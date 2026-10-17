// Package scanner finds the movies in the library folders, reads their details (ffprobe,
// TMDB, a frame from the video) and keeps the list up to date when files change.
package scanner

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/media"
	"local-cinema/internal/cinema/progress"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/thumbs"
)

// Service scans the library.
type Service struct {
	st     *store.Store
	thumbs *thumbs.Service
	Meta   *Metadata

	// State is the progress of the scan.
	State progress.Tracker

	issues issueList

	mu          sync.Mutex
	watch       *watcher
	lastFolders string
}

// New returns the scanner.
func New(st *store.Store, th *thumbs.Service, meta *Metadata) *Service {
	return &Service{st: st, thumbs: th, Meta: meta}
}

// Issues lists the problems the last scan ran into (at most 20).
func (s *Service) Issues() []string { return s.issues.all() }

// Run scans every movie folder. It returns at once when a scan is already running.
func (s *Service) Run() {
	if !s.State.Begin() {
		return
	}
	defer s.State.End()

	key := s.Meta.Key()
	s.issues.reset()
	files := media.FindVideos(s.st.Folders())
	s.State.Set(len(files), 0, "")
	if len(files) == 0 {
		s.issues.add("No video files were found in the movie folders.")
	}

	coll, err := s.st.App.FindCollectionByNameOrId("movies")
	if err != nil {
		return
	}
	existing, _ := s.st.App.FindAllRecords("movies")
	byPath := map[string]*core.Record{}
	for _, r := range existing {
		byPath[r.GetString("file_path")] = r
	}
	seen := map[string]bool{}

	for i, p := range files {
		seen[p] = true
		s.State.Set(-1, i, filepath.Base(p))
		st, err := os.Stat(p)
		if err != nil {
			s.issues.add(p + ": cannot read the file (" + err.Error() + ")")
			continue
		}
		rec := byPath[p]
		changed := false
		if rec == nil {
			title, year := media.ParseName(p)
			rec = core.NewRecord(coll)
			rec.Set("title", title)
			rec.Set("year", year)
			rec.Set("file_path", p)
			rec.Set("added_at", types.NowDateTime())
			changed = true
		} else if int64(rec.GetFloat("file_size")) != st.Size() {
			changed = true
			_ = os.Remove(s.st.Cache.Compat(rec.Id)) // the file was replaced, so the old converted copy is stale
		}
		rec.Set("file_size", st.Size())
		rec.Set("container", ContainerOf(p))
		if changed {
			pi := ffmpeg.Probe(s.st.Tools.FFprobe(), p)
			rec.Set("duration_seconds", pi.Duration)
			rec.Set("video_codec", pi.VCodec)
			rec.Set("audio_codec", pi.ACodec)
			rec.Set("resolution", pi.Resolution)
		}
		rec.Set("available", true)
		rec.Set("last_scanned", types.NowDateTime())
		if err := s.st.App.Save(rec); err != nil {
			slog.Warn("save movie", "path", p, "err", err)
			s.issues.add(p + ": could not be added (" + err.Error() + ")")
			continue
		}
		if key != "" && rec.GetFloat("tmdb_id") == 0 && !rec.GetBool("edited") {
			s.Meta.Fetch(rec, key)
			_ = s.st.App.Save(rec)
			time.Sleep(100 * time.Millisecond) // ~10 requests a second
		}
		// No TMDB art (offline, no key, no match): grab a frame from the video itself.
		if s.thumbs.Needed(rec) {
			if err := s.thumbs.Make(rec, 0, false); err != nil {
				s.thumbs.MarkFailed(rec.Id)
				slog.Warn("thumbnail failed", "movie", rec.GetString("title"), "err", err)
			} else {
				_ = s.st.App.Save(rec)
			}
		}
	}

	for p, r := range byPath {
		if !seen[p] && r.GetBool("available") {
			r.Set("available", false)
			_ = s.st.App.Save(r)
		}
	}
	s.State.Set(len(files), len(files), "")
	slog.Info("scan complete", "files", len(files))
}

// ContainerOf is the container name stored for a file: its extension in lower case without
// the dot, for example "mkv".
func ContainerOf(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}

// StartWatching begins watching the library folders for changes.
func (s *Service) StartWatching() {
	wt, err := newWatcher(s.debouncedScan, DebounceDelay)
	if err != nil {
		slog.Warn("folder watching unavailable", "err", err)
		return
	}
	s.mu.Lock()
	s.watch = wt
	s.lastFolders = store.FoldersKey(s.st.Folders())
	s.mu.Unlock()
	go wt.run()
	s.SyncWatches()
}

// SyncWatches rebuilds the watch list from the current settings.
func (s *Service) SyncWatches() {
	s.mu.Lock()
	wt := s.watch
	s.mu.Unlock()
	if wt == nil {
		return
	}
	cfg := s.st.Settings()
	wt.sync(s.st.Folders(), cfg.GetBool("watch_folders") && cfg.GetBool("setup_complete"))
}

// debouncedScan waits for file copies to settle, then rescans.
func (s *Service) debouncedScan() {
	if s.State.Running() {
		s.mu.Lock()
		wt := s.watch
		s.mu.Unlock()
		if wt != nil {
			wt.retry()
		}
		return
	}
	slog.Info("library change detected, rescanning")
	s.Run()
}

// OnSettingsUpdated re-syncs the watches and rescans when the library folders changed.
func (s *Service) OnSettingsUpdated(rec *core.Record) {
	if !rec.GetBool("setup_complete") {
		return
	}
	s.SyncWatches()
	key := store.FoldersKey(s.st.Folders())
	s.mu.Lock()
	changed := key != s.lastFolders
	if changed {
		s.lastFolders = key
	}
	s.mu.Unlock()
	if changed {
		go s.Run()
	}
}

// Close stops watching.
func (s *Service) Close() {
	s.mu.Lock()
	wt := s.watch
	s.mu.Unlock()
	if wt != nil {
		wt.close()
	}
}

// Mount registers the scan endpoint.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.POST("/scan", s.handleScan)
}

func (s *Service) handleScan(e *core.RequestEvent) error {
	if s.State.Running() || s.thumbs.Job.Running() {
		return e.JSON(409, map[string]any{"error": "scan_running", "message": "A scan is already running."})
	}
	go s.Run()
	return e.JSON(202, map[string]any{"ok": true})
}
