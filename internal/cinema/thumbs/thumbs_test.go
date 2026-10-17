package thumbs

import (
	"os"
	"path/filepath"
	"testing"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/store"
)

// The tests run this test binary as a stand-in for ffmpeg: when CINEMA_FAKE_FFMPEG is set the
// program behaves like ffmpeg instead of running tests. That works the same on every system.
func TestMain(m *testing.M) {
	if mode := os.Getenv("CINEMA_FAKE_FFMPEG"); mode != "" {
		os.Exit(fakeFFmpeg(mode, os.Args[1:]))
	}
	os.Exit(m.Run())
}

func fakeFFmpeg(mode string, args []string) int {
	switch mode {
	case "ok":
		for i, a := range args {
			if a == "[bd]" && i+5 < len(args) {
				_ = os.WriteFile(args[i+5], []byte("backdrop-frame"), 0o644)
			}
		}
		_ = os.WriteFile(args[len(args)-1], []byte("poster-frame"), 0o644)
		return 0
	case "fail":
		return 1
	}
	return 2
}

func fake(t *testing.T, mode string) string {
	t.Helper()
	t.Setenv("CINEMA_FAKE_FFMPEG", mode)
	return os.Args[0]
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestMissingArt(t *testing.T) {
	cases := []struct{ poster, backdrop, want bool }{
		{false, false, true}, {true, false, true}, {false, true, true}, {true, true, false},
	}
	for _, c := range cases {
		if got := MissingArt(c.poster, c.backdrop); got != c.want {
			t.Errorf("MissingArt(%v, %v) = %v, want %v", c.poster, c.backdrop, got, c.want)
		}
	}
}

func TestWriteFramesWritesBothPictures(t *testing.T) {
	bin := fake(t, "ok")
	dir := t.TempDir()
	poster, backdrop := filepath.Join(dir, "p.jpg"), filepath.Join(dir, "b.jpg")
	if err := writeFrames(bin, "movie.mp4", 10, poster, backdrop, true, true); err != nil {
		t.Fatal(err)
	}
	if read(t, poster) != "poster-frame" || read(t, backdrop) != "backdrop-frame" {
		t.Error("both pictures should be written")
	}
	left, _ := filepath.Glob(filepath.Join(dir, "*.tmp.jpg"))
	if len(left) != 0 {
		t.Errorf("temporary files should be removed, found %v", left)
	}
}

func TestWriteFramesKeepsPicturesNotWanted(t *testing.T) {
	bin := fake(t, "ok")
	dir := t.TempDir()
	poster, backdrop := filepath.Join(dir, "p.jpg"), filepath.Join(dir, "b.jpg")
	if err := os.WriteFile(poster, []byte("TMDB poster"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFrames(bin, "movie.mp4", 10, poster, backdrop, false, true); err != nil {
		t.Fatal(err)
	}
	if read(t, poster) != "TMDB poster" {
		t.Error("a poster that was not asked for (TMDB art) must not be replaced")
	}
	if read(t, backdrop) != "backdrop-frame" {
		t.Error("the backdrop was asked for and should be written")
	}
}

func TestWriteFramesFailureLeavesExistingPictures(t *testing.T) {
	bin := fake(t, "fail")
	dir := t.TempDir()
	poster, backdrop := filepath.Join(dir, "p.jpg"), filepath.Join(dir, "b.jpg")
	for _, p := range []string{poster, backdrop} {
		if err := os.WriteFile(p, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeFrames(bin, "movie.mp4", 10, poster, backdrop, true, true); err == nil {
		t.Fatal("a failing ffmpeg must report an error")
	}
	if read(t, poster) != "old" || read(t, backdrop) != "old" {
		t.Error("a failed run must keep the pictures that were there")
	}
}

func TestWriteFramesWithoutFFmpeg(t *testing.T) {
	if err := writeFrames("", "movie.mp4", 10, "p.jpg", "b.jpg", true, true); err == nil {
		t.Error("an empty ffmpeg path should be an error")
	}
}

func TestMarkFailedStopsRetries(t *testing.T) {
	s := New(&store.Store{Cache: cache.New(t.TempDir()), Tools: ffmpeg.NewTools()})
	if _, failed := s.bad.Load("m1"); failed {
		t.Fatal("nothing has failed yet")
	}
	s.MarkFailed("m1")
	if _, failed := s.bad.Load("m1"); !failed {
		t.Error("MarkFailed should be remembered")
	}
}
