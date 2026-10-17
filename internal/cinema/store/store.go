// Package store is the shared base of the feature packages: the PocketBase app, the cache
// folder layout, the FFmpeg locator, and the few operations everybody needs (read the
// settings, write an audit line, check a path is inside the library).
package store

import (
	"log/slog"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/media"
)

// Store is created once and handed to every feature package.
type Store struct {
	App     core.App
	Cache   cache.Layout  // set by Init
	Tools   *ffmpeg.Tools // FFmpeg and ffprobe
	Started time.Time
}

// New returns a store for the app. Call Init once the data folder is known.
func New(app core.App) *Store {
	return &Store{App: app, Tools: ffmpeg.NewTools(), Started: time.Now()}
}

// Init points the cache and the tools folder at the app's data directory and prepares them.
// It runs when the server starts, before any handler or background job.
func (s *Store) Init() {
	dir := s.App.DataDir()
	s.Tools.SetDataDir(dir)
	s.Cache = cache.New(dir)
	_ = s.Cache.Ensure()
	s.Cache.CleanPartialConversions()
}

// Settings returns the single settings record, creating it with defaults if missing.
func (s *Store) Settings() *core.Record {
	recs, err := s.App.FindAllRecords("settings")
	if err == nil && len(recs) > 0 {
		return recs[0]
	}
	c, err := s.App.FindCollectionByNameOrId("settings")
	if err != nil {
		slog.Error("settings collection missing", "err", err)
		return nil
	}
	r := core.NewRecord(c)
	r.Set("cinema_name", "My Cinema")
	r.Set("accent_color", "#7C3AED")
	r.Set("movie_folders", []string{})
	r.Set("watch_folders", true)
	r.Set("global_max_streams", 10)
	r.Set("mdns_name", "cinema")
	r.Set("setup_complete", false)
	r.Set("setup_step", 1)
	if err := s.App.Save(r); err != nil {
		slog.Error("create settings", "err", err)
	}
	return r
}

// Folders lists the movie folders from the settings.
func (s *Store) Folders() []string {
	rec := s.Settings()
	if rec == nil {
		return nil
	}
	var out []string
	_ = rec.UnmarshalJSONField("movie_folders", &out)
	return out
}

// FoldersKey turns a folder list into one comparable string, to notice when it changed.
func FoldersKey(folders []string) string { return strings.Join(folders, "|") }

// InsideLibrary reports whether a file lives in one of the movie folders (symlinks resolved).
func (s *Store) InsideLibrary(path string) bool {
	return media.InsideRoots(path, s.Folders())
}

// Streamable says a movie can be shown to guests: its file exists and the admin has not hidden it.
func Streamable(r *core.Record) bool {
	return r.GetBool("available") && !r.GetBool("hidden")
}

// Touch marks a session as seen now.
func (s *Store) Touch(sess *core.Record) {
	sess.Set("last_seen", types.NowDateTime())
	_ = s.App.Save(sess)
}

// Audit writes one line to the audit log and to the program log.
func (s *Store) Audit(event, user, session, movie string, detail map[string]any) {
	c, err := s.App.FindCollectionByNameOrId("audit_log")
	if err != nil {
		return
	}
	r := core.NewRecord(c)
	r.Set("event", event)
	if user != "" {
		r.Set("user", user)
	}
	if session != "" {
		r.Set("session", session)
	}
	if movie != "" {
		r.Set("movie", movie)
	}
	r.Set("detail", detail)
	_ = s.App.Save(r)
	slog.Info(event, "user", user, "session", session, "movie", movie)
}
