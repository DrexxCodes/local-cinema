package progress

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestBeginIsExclusive(t *testing.T) {
	var tr Tracker
	if !tr.Begin() {
		t.Fatal("first Begin should succeed")
	}
	if tr.Begin() {
		t.Fatal("second Begin while running should fail")
	}
	tr.End()
	if !tr.Begin() {
		t.Fatal("Begin after End should succeed")
	}
}

func TestBeginResetsCounters(t *testing.T) {
	var tr Tracker
	tr.Begin()
	tr.Set(10, 4, "movie.mp4")
	tr.SetPercent(80)
	tr.End()
	tr.Begin()
	got := tr.Snapshot()
	want := Snapshot{Running: true}
	if got != want {
		t.Errorf("snapshot after Begin = %+v, want %+v", got, want)
	}
}

func TestEndKeepsTotalsButClearsCurrent(t *testing.T) {
	var tr Tracker
	tr.Begin()
	tr.Set(5, 5, "last.mkv")
	tr.End()
	got := tr.Snapshot()
	if got.Running || got.Current != "" || got.Total != 5 || got.Done != 5 {
		t.Errorf("unexpected snapshot after End: %+v", got)
	}
	if tr.Running() {
		t.Error("Running should be false after End")
	}
}

func TestSetNegativeTotalKeepsTotal(t *testing.T) {
	var tr Tracker
	tr.Begin()
	tr.Set(7, 0, "")
	tr.Set(-1, 3, "x")
	got := tr.Snapshot()
	if got.Total != 7 || got.Done != 3 || got.Current != "x" {
		t.Errorf("unexpected snapshot: %+v", got)
	}
}

func TestSetPercentClamps(t *testing.T) {
	var tr Tracker
	for in, want := range map[int]int{-5: 0, 0: 0, 42: 42, 100: 100, 250: 100} {
		tr.SetPercent(in)
		if got := tr.Snapshot().Percent; got != want {
			t.Errorf("SetPercent(%d) gave %d, want %d", in, got, want)
		}
	}
}

func TestSnapshotJSONKeys(t *testing.T) {
	var tr Tracker
	tr.Begin()
	tr.Set(2, 1, "a")
	raw, err := json.Marshal(tr.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"running", "total", "done", "current", "percent"} {
		if _, ok := m[k]; !ok {
			t.Errorf("JSON is missing the %q key the UI reads: %s", k, raw)
		}
	}
}

func TestConcurrentUse(t *testing.T) {
	var tr Tracker
	tr.Begin()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tr.Set(100, i, "x")
			tr.SetPercent(i)
			_ = tr.Snapshot()
			_ = tr.Running()
		}(i)
	}
	wg.Wait()
}
