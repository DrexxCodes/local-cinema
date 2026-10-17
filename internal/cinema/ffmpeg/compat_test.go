package ffmpeg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestUniversal(t *testing.T) {
	cases := []struct {
		container, v, a string
		want            bool
	}{
		{"mp4", "h264", "aac", true},
		{"mp4", "h264", "mp3", true},
		{"mp4", "h264", "", true}, // no audio track
		{"m4v", "h264", "aac", true},
		{"mov", "h264", "aac", true},
		{"mp4", "hevc", "aac", false},
		{"mp4", "h264", "ac3", false},
		{"mp4", "h264", "dts", false},
		{"mkv", "h264", "aac", false},
		{"avi", "h264", "mp3", false},
		{"webm", "vp9", "opus", false},
	}
	for _, c := range cases {
		if got := Universal(c.container, c.v, c.a); got != c.want {
			t.Errorf("Universal(%q, %q, %q) = %v, want %v", c.container, c.v, c.a, got, c.want)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name            string
		container, v, a string
		converted       bool
		want            string
	}{
		{"converted copy wins", "mkv", "hevc", "dts", true, StatusConverted},
		{"converted copy wins over unknown", "mkv", "", "", true, StatusConverted},
		{"no codec info", "mp4", "", "", false, StatusUnknown},
		{"plays everywhere", "mp4", "h264", "aac", false, StatusOK},
		{"needs conversion", "mkv", "h264", "ac3", false, StatusConvert},
	}
	for _, c := range cases {
		if got := Classify(c.container, c.v, c.a, c.converted); got != c.want {
			t.Errorf("%s: Classify = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestNeedsConversion(t *testing.T) {
	cases := []struct {
		status   string
		explicit bool
		want     bool
	}{
		{StatusConvert, false, true},
		{StatusConvert, true, true},
		{StatusUnknown, false, false}, // "Make all compatible" skips what it cannot judge
		{StatusUnknown, true, true},   // but a movie picked by hand is tried
		{StatusOK, true, false},
		{StatusConverted, true, false},
	}
	for _, c := range cases {
		if got := NeedsConversion(c.status, c.explicit); got != c.want {
			t.Errorf("NeedsConversion(%q, %v) = %v, want %v", c.status, c.explicit, got, c.want)
		}
	}
}

func hasPair(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestConvertArgsCopiesWhatItCan(t *testing.T) {
	args := ConvertArgs("in.mkv", "out.part.mp4", "h264", "aac")
	if !hasPair(args, "-c:v", "copy") || !hasPair(args, "-c:a", "copy") {
		t.Errorf("h264 + aac should be copied: %v", args)
	}
	if slices.Contains(args, "libx264") {
		t.Errorf("must not re-encode h264: %v", args)
	}
}

func TestConvertArgsReencodes(t *testing.T) {
	args := ConvertArgs("in.mkv", "out.part.mp4", "hevc", "dts")
	if !hasPair(args, "-c:v", "libx264") || !hasPair(args, "-pix_fmt", "yuv420p") {
		t.Errorf("hevc should become h264: %v", args)
	}
	if !hasPair(args, "-c:a", "aac") || !hasPair(args, "-ac", "2") {
		t.Errorf("dts should become stereo aac: %v", args)
	}
}

func TestConvertArgsShape(t *testing.T) {
	args := ConvertArgs("in.mkv", "out.part.mp4", "h264", "aac")
	if !hasPair(args, "-i", "in.mkv") {
		t.Errorf("input missing: %v", args)
	}
	if args[len(args)-1] != "out.part.mp4" {
		t.Errorf("the output file must be last: %v", args)
	}
	for _, want := range []string{"-nostdin", "-y", "+faststart", "pipe:1"} {
		if !slices.Contains(args, want) {
			t.Errorf("%q missing from %v", want, args)
		}
	}
	if !hasPair(args, "-map", "0:a:0?") {
		t.Errorf("audio must be optional so silent films convert: %v", args)
	}
}

func TestParseProgress(t *testing.T) {
	cases := []struct {
		line  string
		total float64
		pct   int
		ok    bool
	}{
		{"out_time_us=5000000", 10, 50, true},
		{"out_time_us=10000000", 10, 100, true},
		{"out_time_us=99000000", 10, 100, true}, // never above 100
		{"out_time_us=0", 10, 0, true},
		{"out_time_us=-1000000", 10, 0, true}, // ffmpeg sometimes starts negative
		{"out_time_us=N/A", 10, 0, false},
		{"out_time_us=5000000", 0, 0, false}, // length unknown
		{"frame=120", 10, 0, false},
		{"progress=end", 10, 0, false},
		{"", 10, 0, false},
	}
	for _, c := range cases {
		pct, ok := ParseProgress(c.line, c.total)
		if pct != c.pct || ok != c.ok {
			t.Errorf("ParseProgress(%q, %v) = %d, %v; want %d, %v", c.line, c.total, pct, ok, c.pct, c.ok)
		}
	}
}

func TestConvertSucceeds(t *testing.T) {
	bin := useFake(t, "convert-ok")
	dir := t.TempDir()
	out := filepath.Join(dir, "m1.mp4")
	var mu sync.Mutex
	var seen []int
	err := Convert(context.Background(), bin, ConvertOptions{
		Src: "in.mkv", Out: out, VCodec: "h264", ACodec: "aac", Duration: 10,
		OnPercent: func(p int) { mu.Lock(); seen = append(seen, p); mu.Unlock() },
	})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(out); err != nil || string(b) != "fake mp4" {
		t.Errorf("the finished file is missing or wrong: %q, %v", b, err)
	}
	if _, err := os.Stat(strings.TrimSuffix(out, ".mp4") + ".part.mp4"); !os.IsNotExist(err) {
		t.Error("the .part file must be gone")
	}
	if !slices.Equal(seen, []int{50, 100}) {
		t.Errorf("progress = %v, want [50 100] (the N/A line is skipped)", seen)
	}
}

func TestConvertReportsFfmpegErrors(t *testing.T) {
	bin := useFake(t, "convert-fail")
	out := filepath.Join(t.TempDir(), "m.mp4")
	err := Convert(context.Background(), bin, ConvertOptions{Src: "in.mkv", Out: out})
	if err == nil {
		t.Fatal("a failing ffmpeg should be an error")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "ffmpeg failed: ") || !strings.HasSuffix(msg, "THE-END") {
		t.Errorf("the message should end with the tail of stderr: %q", msg)
	}
	if len(msg) > len("ffmpeg failed: ")+300 {
		t.Errorf("stderr should be trimmed to 300 characters, message is %d long", len(msg))
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Error("a failed conversion must not leave an output file")
	}
}

func TestConvertWithoutOutputFile(t *testing.T) {
	bin := useFake(t, "convert-empty")
	err := Convert(context.Background(), bin, ConvertOptions{Src: "in.mkv", Out: filepath.Join(t.TempDir(), "m.mp4")})
	if err == nil || !strings.Contains(err.Error(), "no file") {
		t.Errorf("want a 'produced no file' error, got %v", err)
	}
}

func TestConvertCancel(t *testing.T) {
	bin := useFake(t, "convert-hang")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Convert(ctx, bin, ConvertOptions{Src: "in.mkv", Out: filepath.Join(t.TempDir(), "m.mp4")})
	}()
	time.Sleep(300 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("a cancelled conversion should return context.Canceled, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Convert did not stop after the context was cancelled")
	}
}

func TestConvertNeedsFfmpeg(t *testing.T) {
	if err := Convert(context.Background(), "", ConvertOptions{}); err == nil {
		t.Error("an empty program path should be an error")
	}
}
