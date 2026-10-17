// Package progress tracks a background job (scan, thumbnails, conversion, download)
// so the admin page can poll how far along it is.
package progress

import (
	"sync"

	"local-cinema/internal/cinema/util"
)

// Snapshot is the JSON shape the UI polls.
type Snapshot struct {
	Running bool   `json:"running"`
	Total   int    `json:"total"`
	Done    int    `json:"done"`
	Current string `json:"current"`
	Percent int    `json:"percent"` // progress of the current item, 0 to 100
}

// Tracker is safe for concurrent use. Only one run can be active at a time.
type Tracker struct {
	mu sync.Mutex
	s  Snapshot
}

// Begin starts a run and resets the counters. It returns false when one is already running.
func (t *Tracker) Begin() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.s.Running {
		return false
	}
	t.s = Snapshot{Running: true}
	return true
}

// End marks the run as finished. The final Total and Done are kept for the UI.
func (t *Tracker) End() {
	t.mu.Lock()
	t.s.Running, t.s.Current = false, ""
	t.mu.Unlock()
}

// SetPercent sets the progress of the current item, limited to 0 to 100.
func (t *Tracker) SetPercent(p int) {
	t.mu.Lock()
	t.s.Percent = util.ClampInt(p, 0, 100)
	t.mu.Unlock()
}

// Running reports whether a run is active.
func (t *Tracker) Running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.s.Running
}

// Set updates the counters. A negative total leaves the total unchanged.
func (t *Tracker) Set(total, done int, current string) {
	t.mu.Lock()
	if total >= 0 {
		t.s.Total = total
	}
	t.s.Done, t.s.Current = done, current
	t.mu.Unlock()
}

// Snapshot returns a copy of the current state.
func (t *Tracker) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.s
}
