package compat

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	l := cache.New(t.TempDir())
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	return New(&store.Store{Cache: l})
}

func TestJoinFailures(t *testing.T) {
	if got := JoinFailures(nil); got != "" {
		t.Errorf("no failures = %q", got)
	}
	if got := JoinFailures([]string{"A: x", "B: y"}); got != "A: x | B: y" {
		t.Errorf("got %q", got)
	}
}

func TestFileAndExists(t *testing.T) {
	s := testService(t)
	if s.Exists("m1") {
		t.Fatal("no converted copy yet")
	}
	if want := filepath.Join(s.st.Cache.Root, "compat", "m1.mp4"); s.File("m1") != want {
		t.Errorf("File = %q, want %q", s.File("m1"), want)
	}
	if err := os.WriteFile(s.File("m1"), []byte("mp4"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !s.Exists("m1") {
		t.Error("the copy exists now")
	}
	if err := os.WriteFile(s.File("m2"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if s.Exists("m2") {
		t.Error("an empty file is a failed conversion, not a copy")
	}
}

func TestErrorIsRememberedAndCleared(t *testing.T) {
	s := testService(t)
	s.setError("A: boom")
	if s.LastError() != "A: boom" {
		t.Error("error should be kept")
	}
	s.setError("")
	if s.LastError() != "" {
		t.Error("error should be cleared")
	}
}

func TestCancelStopsTheRunningJob(t *testing.T) {
	s := testService(t)
	s.Cancel() // nothing running: must not panic

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.stop = cancel
	s.mu.Unlock()
	s.Cancel()
	select {
	case <-ctx.Done():
	default:
		t.Error("Cancel should cancel the job's context")
	}
}

func TestOnlyOneJobAtATime(t *testing.T) {
	s := testService(t)
	if !s.Job.Begin() {
		t.Fatal("first job starts")
	}
	s.RunJob(nil, true) // a second job must return at once without touching the database
	if !s.Job.Running() {
		t.Error("the first job is still running")
	}
}
