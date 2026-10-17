package scanner

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestContainerOf(t *testing.T) {
	cases := map[string]string{
		"/m/Movie.MKV":    "mkv",
		"Movie.mp4":       "mp4",
		"a.b.c.WebM":      "webm",
		"noextension":     "",
		`C:\movies\a.AVI`: "avi",
	}
	for in, want := range cases {
		if got := ContainerOf(in); got != want {
			t.Errorf("ContainerOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIssueListKeepsTheFirstTwenty(t *testing.T) {
	var l issueList
	for i := 0; i < 30; i++ {
		l.add("problem " + strconv.Itoa(i))
	}
	got := l.all()
	if len(got) != maxIssues || got[0] != "problem 0" || got[maxIssues-1] != "problem 19" {
		t.Errorf("got %d issues, first %q, last %q", len(got), got[0], got[len(got)-1])
	}
	got[0] = "changed"
	if l.all()[0] != "problem 0" {
		t.Error("all() must return a copy")
	}
}

func TestIssueListResetAndJSONShape(t *testing.T) {
	var l issueList
	if got := l.all(); got == nil || len(got) != 0 {
		t.Errorf("an empty list must be [] not null, got %#v", got)
	}
	l.add("x")
	l.reset()
	if len(l.all()) != 0 {
		t.Error("reset should clear the list")
	}
}

func TestRelevant(t *testing.T) {
	cases := []struct {
		name         string
		video, isDir bool
	}{
		{"/lib/a.mp4", true, false},
		{"/lib/A.MKV", true, false},
		{"/lib/Action", false, true}, // no extension: probably a folder
		{"/lib/a.srt", false, false},
		{"/lib/a.txt", false, false},
		{"/lib/a.mp4.part", false, false},
	}
	for _, c := range cases {
		v, d := relevant(c.name)
		if v != c.video || d != c.isDir {
			t.Errorf("relevant(%q) = (%v, %v), want (%v, %v)", c.name, v, d, c.video, c.isDir)
		}
	}
}

func newTestWatcher(t *testing.T, delay time.Duration) (*watcher, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	wt, err := newWatcher(func() { n.Add(1) }, delay)
	if err != nil {
		t.Skipf("file watching is not available here: %v", err)
	}
	t.Cleanup(wt.close)
	return wt, &n
}

func TestWatcherCoalescesBursts(t *testing.T) {
	wt, n := newTestWatcher(t, 60*time.Millisecond)
	for i := 0; i < 5; i++ {
		wt.onEvent(fsnotify.Event{Name: "/lib/new.mp4", Op: fsnotify.Write})
		time.Sleep(10 * time.Millisecond)
	}
	if n.Load() != 0 {
		t.Fatal("must wait for the quiet period")
	}
	time.Sleep(200 * time.Millisecond)
	if got := n.Load(); got != 1 {
		t.Errorf("a burst of changes should trigger one rescan, got %d", got)
	}
}

func TestWatcherIgnoresUnrelatedFiles(t *testing.T) {
	wt, n := newTestWatcher(t, 20*time.Millisecond)
	wt.onEvent(fsnotify.Event{Name: "/lib/readme.txt", Op: fsnotify.Create})
	wt.onEvent(fsnotify.Event{Name: "/lib/a.srt", Op: fsnotify.Write})
	time.Sleep(100 * time.Millisecond)
	if n.Load() != 0 {
		t.Error("only videos and folders should start a rescan")
	}
}

func TestWatcherRetryRestartsTheWait(t *testing.T) {
	wt, n := newTestWatcher(t, 30*time.Millisecond)
	wt.retry()
	time.Sleep(120 * time.Millisecond)
	if n.Load() != 1 {
		t.Errorf("retry should trigger once after the delay, got %d", n.Load())
	}
}

func TestWatcherCloseStopsPendingTrigger(t *testing.T) {
	wt, n := newTestWatcher(t, 50*time.Millisecond)
	wt.onEvent(fsnotify.Event{Name: "/lib/a.mp4", Op: fsnotify.Create})
	wt.close()
	time.Sleep(150 * time.Millisecond)
	if n.Load() != 0 {
		t.Error("a closed watcher must not trigger a scan")
	}
}

func TestWatcherSyncDisabledWatchesNothing(t *testing.T) {
	wt, _ := newTestWatcher(t, time.Second)
	dir := t.TempDir()
	wt.sync([]string{dir}, false)
	if len(wt.watched) != 0 {
		t.Errorf("nothing should be watched when disabled, got %v", wt.watched)
	}
	wt.sync([]string{dir}, true)
	if !wt.watched[dir] {
		t.Error("the folder should be watched when enabled")
	}
	wt.sync(nil, true)
	if len(wt.watched) != 0 {
		t.Error("a folder removed from the settings must stop being watched")
	}
}
