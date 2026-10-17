package ffmpeg

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestPickSeek(t *testing.T) {
	cases := []struct {
		duration, want float64
	}{
		{0, 5},       // unknown length
		{-10, 5},     // nonsense length
		{7200, 300},  // 12% would be 864s, capped at 5 minutes
		{1000, 120},  // 12%
		{100, 12},    // 12%
		{5, 2.5},     // very short clip: the middle
		{0.5, 0.25},  // tiny clip
		{2500, 300},  // exactly at the cap
		{8.4, 1.008}, // just past the one second floor
	}
	for _, c := range cases {
		got := PickSeek(c.duration)
		if diff := got - c.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("PickSeek(%v) = %v, want %v", c.duration, got, c.want)
		}
	}
}

func TestThumbArgs(t *testing.T) {
	args := ThumbArgs("in.mp4", 12.345, "bd.tmp.jpg", "po.tmp.jpg")
	if !hasPair(args, "-ss", "12.35") && !hasPair(args, "-ss", "12.34") {
		t.Errorf("seek position missing: %v", args)
	}
	if !hasPair(args, "-i", "in.mp4") {
		t.Errorf("input missing: %v", args)
	}
	idxBd := slices.Index(args, "bd.tmp.jpg")
	idxPo := slices.Index(args, "po.tmp.jpg")
	if idxBd < 0 || idxPo < 0 || idxBd > idxPo {
		t.Errorf("backdrop output must come before the poster output: %v", args)
	}
	if args[idxBd-6] != "-map" || args[idxBd-5] != "[bd]" {
		t.Errorf("backdrop should be mapped from [bd]: %v", args)
	}
	if args[idxPo-6] != "-map" || args[idxPo-5] != "[po]" {
		t.Errorf("poster should be mapped from [po]: %v", args)
	}
}

func TestExtractThumbs(t *testing.T) {
	bin := useFake(t, "thumbs-ok")
	dir := t.TempDir()
	bd, po := filepath.Join(dir, "bd.jpg"), filepath.Join(dir, "po.jpg")
	if err := ExtractThumbs(bin, "in.mp4", 30, bd, po); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{bd, po} {
		if st, err := os.Stat(p); err != nil || st.Size() == 0 {
			t.Errorf("%s should exist and have content", p)
		}
	}
}

func TestExtractThumbsFallsBackToOneSecond(t *testing.T) {
	bin := useFake(t, "thumbs-fail-late")
	dir := t.TempDir()
	bd, po := filepath.Join(dir, "bd.jpg"), filepath.Join(dir, "po.jpg")
	if err := ExtractThumbs(bin, "in.mp4", 9999, bd, po); err != nil {
		t.Fatalf("the retry at 1 second should have worked: %v", err)
	}
	if _, err := os.Stat(po); err != nil {
		t.Error("the poster should exist after the retry")
	}
}

func TestExtractThumbsDoesNotRetryAtOneSecond(t *testing.T) {
	// At 1 second or less there is nothing earlier to fall back to.
	bin := useFake(t, "thumbs-fail-late")
	dir := t.TempDir()
	if err := ExtractThumbs(bin, "in.mp4", 0.5, filepath.Join(dir, "a.jpg"), filepath.Join(dir, "b.jpg")); err == nil {
		t.Error("a failure at 0.5 seconds should be reported")
	}
}

func TestExtractThumbsFailures(t *testing.T) {
	dir := t.TempDir()
	bd, po := filepath.Join(dir, "bd.jpg"), filepath.Join(dir, "po.jpg")

	bin := useFake(t, "thumbs-fail")
	if err := ExtractThumbs(bin, "in.mp4", 30, bd, po); err == nil {
		t.Error("ffmpeg failing should be an error")
	}
	bin = useFake(t, "thumbs-empty")
	if err := ExtractThumbs(bin, "in.mp4", 30, bd, po); err == nil {
		t.Error("ffmpeg writing nothing should be an error")
	}
	if err := ExtractThumbs("", "in.mp4", 30, bd, po); err == nil {
		t.Error("a missing program should be an error")
	}
}
