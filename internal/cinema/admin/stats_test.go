package admin

import (
	"encoding/json"
	"sort"
	"testing"
	"time"
)

// The admin UI reads these keys by name, so renaming one breaks the dashboard.
func TestBuildKeyContract(t *testing.T) {
	got := Stats{}.Build()
	want := []string{
		"active_sessions", "convert", "convert_error", "ffmpeg", "ffmpeg_install", "movies",
		"pending_upgrades", "scan", "scan_issues", "subs", "subs_msg", "thumbs",
		"update_available", "uptime_seconds",
	}
	var keys []string
	for k := range got {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
}

func TestBuildValues(t *testing.T) {
	got := Stats{ActiveSessions: 3, Movies: 120, Uptime: 90*time.Second + 400*time.Millisecond, HasFFmpeg: true,
		UpdateAvailable: "1.2.0", ScanIssues: []string{"a.mkv"}}.Build()
	if got["active_sessions"] != int64(3) || got["movies"] != int64(120) {
		t.Error("counts must pass through")
	}
	if got["uptime_seconds"] != 90 {
		t.Errorf("uptime_seconds = %v, want 90", got["uptime_seconds"])
	}
	if got["ffmpeg"] != true || got["update_available"] != "1.2.0" {
		t.Error("flags must pass through")
	}
}

func TestBuildJSONShape(t *testing.T) {
	b, err := json.Marshal(Stats{}.Build())
	if err != nil {
		t.Fatal(err)
	}
	var back map[string]json.RawMessage
	_ = json.Unmarshal(b, &back)
	if string(back["scan_issues"]) != "[]" {
		t.Errorf("no issues must be [] and not null, got %s", back["scan_issues"])
	}
	var scan map[string]any
	_ = json.Unmarshal(back["scan"], &scan)
	for _, k := range []string{"running", "total", "done", "current", "percent"} {
		if _, ok := scan[k]; !ok {
			t.Errorf("scan is missing %q", k)
		}
	}
}
