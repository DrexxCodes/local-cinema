package ratelimit

import (
	"strconv"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTest(max int, window time.Duration) (*Limiter, *clock) {
	c := &clock{t: time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)}
	l := New(max, window)
	l.now = c.now
	return l, c
}

func TestAllowsUpToMax(t *testing.T) {
	l, _ := newTest(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("1.1.1.1") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if l.Allow("1.1.1.1") {
		t.Fatal("the fourth attempt should be refused")
	}
}

func TestKeysAreIndependent(t *testing.T) {
	l, _ := newTest(1, time.Minute)
	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("different keys must not share a budget")
	}
	if l.Allow("a") {
		t.Fatal("a second attempt for the same key should be refused")
	}
}

func TestWindowExpires(t *testing.T) {
	l, c := newTest(2, time.Minute)
	l.Allow("x")
	l.Allow("x")
	if l.Allow("x") {
		t.Fatal("limit should be reached")
	}
	c.advance(59 * time.Second)
	if l.Allow("x") {
		t.Fatal("still inside the window")
	}
	c.advance(2 * time.Second)
	if !l.Allow("x") {
		t.Fatal("the window has passed, attempts should be allowed again")
	}
}

func TestRefusedAttemptsAreNotRecorded(t *testing.T) {
	l, c := newTest(1, time.Minute)
	l.Allow("x")
	for i := 0; i < 20; i++ {
		c.advance(2 * time.Second)
		if l.Allow("x") {
			t.Fatal("should be refused inside the window")
		}
	}
	// 40s of refused hammering must not push the release time past the first attempt + 1 minute.
	c.advance(21 * time.Second)
	if !l.Allow("x") {
		t.Fatal("refused attempts must not extend the lockout")
	}
}

func TestLoginLimitIsTenPerMinute(t *testing.T) {
	l := NewLogin()
	for i := 0; i < 10; i++ {
		if !l.Allow("ip") {
			t.Fatalf("attempt %d should pass", i+1)
		}
	}
	if l.Allow("ip") {
		t.Fatal("the 11th attempt in a minute should be refused")
	}
}

func TestSweepForgetsIdleKeys(t *testing.T) {
	l, c := newTest(5, time.Minute)
	for i := 0; i < sweepAbove; i++ {
		l.Allow("old-" + strconv.Itoa(i))
	}
	c.advance(2 * time.Minute)
	l.Allow("fresh-1")
	l.Allow("fresh-2") // pushes past the threshold and triggers the sweep
	l.mu.Lock()
	n := len(l.hits)
	l.mu.Unlock()
	if n > 10 {
		t.Errorf("idle keys should be dropped, %d still tracked", n)
	}
	if l.Allow("old-1") != true {
		t.Error("a swept key starts with a fresh budget")
	}
}
