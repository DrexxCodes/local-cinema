package scanner

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"local-cinema/internal/cinema/media"
)

// DebounceDelay is how long the library must stay quiet before a rescan starts, so a long
// file copy triggers one scan and not hundreds.
const DebounceDelay = 10 * time.Second

// watcher watches the library folders and calls trigger once changes stop.
type watcher struct {
	mu       sync.Mutex
	w        *fsnotify.Watcher
	watched  map[string]bool
	timer    *time.Timer
	debounce time.Duration
	trigger  func()
}

func newWatcher(trigger func(), debounce time.Duration) (*watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{w: w, watched: map[string]bool{}, debounce: debounce, trigger: trigger}, nil
}

// run reads file system events until the watcher is closed.
func (wt *watcher) run() {
	for {
		select {
		case ev, ok := <-wt.w.Events:
			if !ok {
				return
			}
			wt.onEvent(ev)
		case err, ok := <-wt.w.Errors:
			if !ok {
				return
			}
			slog.Warn("watcher error", "err", err)
		}
	}
}

// relevant says whether a change to name can affect the library: a video file, or a folder
// (or any name without an extension).
func relevant(name string) (isVideo, isDir bool) {
	ext := strings.ToLower(filepath.Ext(name))
	return media.IsVideoExt(ext), ext == ""
}

func (wt *watcher) onEvent(ev fsnotify.Event) {
	isVideo, isDir := relevant(ev.Name)
	if !isVideo && !isDir {
		return
	}
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if ev.Has(fsnotify.Create) && isDir {
		wt.addTreeLocked(ev.Name) // new sub-folder: watch it too
	}
	wt.restartTimerLocked()
}

// restartTimerLocked (re)starts the quiet period. The caller holds the lock.
func (wt *watcher) restartTimerLocked() {
	if wt.timer != nil {
		wt.timer.Stop()
	}
	wt.timer = time.AfterFunc(wt.debounce, wt.trigger)
}

// retry asks for another quiet period, used when a scan is already running.
func (wt *watcher) retry() {
	wt.mu.Lock()
	wt.restartTimerLocked()
	wt.mu.Unlock()
}

// sync replaces the watch list with the given roots. Nothing is watched when enabled is false.
func (wt *watcher) sync(roots []string, enabled bool) {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	for p := range wt.watched {
		_ = wt.w.Remove(p)
	}
	wt.watched = map[string]bool{}
	if !enabled {
		return
	}
	for _, root := range roots {
		wt.addTreeLocked(root)
	}
}

func (wt *watcher) addTreeLocked(root string) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if !wt.watched[p] {
			if err := wt.w.Add(p); err == nil {
				wt.watched[p] = true
			}
		}
		return nil
	})
}

func (wt *watcher) close() {
	wt.mu.Lock()
	if wt.timer != nil {
		wt.timer.Stop()
	}
	wt.mu.Unlock()
	_ = wt.w.Close()
}
