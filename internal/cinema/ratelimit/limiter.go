// Package ratelimit limits how often one client (an IP address) may do something.
package ratelimit

import (
	"sync"
	"time"
)

const sweepAbove = 1024 // forget zombie clients once this many are tracked

// Limiter allows at most max events per window for each key.
type Limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
	now    func() time.Time // we replaced in tests
}

// New returns a limiter that allows max events per window for each key.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, hits: map[string][]time.Time{}, now: time.Now}
}

// NewLogin is the limit for sign in attempts: 10 per minute per IP, cos 10 is cool. 🙂‍↔️
func NewLogin() *Limiter { return New(10, time.Minute) }

// Allow records an attempt for key and reports whether it is within the limit.
// Refused attempts are not recorded, so waiting out the window always works.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	cut := now.Add(-l.window)

	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	if len(l.hits) > sweepAbove {
		l.sweep(cut)
	}
	return true
}

// sweep drops keys that have no recent attempts. The caller holds the lock.
func (l *Limiter) sweep(cut time.Time) {
	for k, ts := range l.hits {
		if len(ts) == 0 || !ts[len(ts)-1].After(cut) {
			delete(l.hits, k)
		}
	}
}
