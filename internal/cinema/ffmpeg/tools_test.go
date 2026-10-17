package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// isolate removes every place a real ffmpeg could be found except the test's own folders.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	t.Setenv("CINEMA_FFMPEG_DIR", "")
}

func TestNotInstalled(t *testing.T) {
	isolate(t)
	tools := NewTools()
	tools.SetDataDir(filepath.Join(t.TempDir(), "data"))
	if got := tools.FFmpeg(); got != "" {
		t.Errorf("FFmpeg() = %q, want empty", got)
	}
	if got := tools.FFprobe(); got != "" {
		t.Errorf("FFprobe() = %q, want empty", got)
	}
}

func TestFindsToolsInBinFolderNextToData(t *testing.T) {
	isolate(t)
	root := t.TempDir()
	touch(t, filepath.Join(root, "bin", ExeName("ffmpeg")))
	touch(t, filepath.Join(root, "bin", ExeName("ffprobe")))
	tools := NewTools()
	tools.SetDataDir(filepath.Join(root, "data"))
	if got, want := tools.FFmpeg(), filepath.Join(root, "bin", ExeName("ffmpeg")); got != want {
		t.Errorf("FFmpeg() = %q, want %q", got, want)
	}
	if got, want := tools.FFprobe(), filepath.Join(root, "bin", ExeName("ffprobe")); got != want {
		t.Errorf("FFprobe() = %q, want %q", got, want)
	}
	if got, want := tools.Dir(), filepath.Join(root, "bin"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestEnvFolderBeatsBinFolder(t *testing.T) {
	isolate(t)
	root, custom := t.TempDir(), t.TempDir()
	touch(t, filepath.Join(root, "bin", ExeName("ffmpeg")))
	touch(t, filepath.Join(custom, ExeName("ffmpeg")))
	t.Setenv("CINEMA_FFMPEG_DIR", custom)
	tools := NewTools()
	tools.SetDataDir(filepath.Join(root, "data"))
	if got, want := tools.FFmpeg(), filepath.Join(custom, ExeName("ffmpeg")); got != want {
		t.Errorf("FFmpeg() = %q, want the CINEMA_FFMPEG_DIR copy %q", got, want)
	}
}

func TestFindsToolOnPath(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, ExeName("ffmpeg")))
	t.Setenv("PATH", dir)
	t.Setenv("CINEMA_FFMPEG_DIR", "")
	tools := NewTools()
	tools.SetDataDir(filepath.Join(t.TempDir(), "data"))
	if got := tools.FFmpeg(); got == "" {
		t.Error("ffmpeg on the PATH should be found")
	}
}

func TestResetSearchesAgain(t *testing.T) {
	isolate(t)
	root := t.TempDir()
	tools := NewTools()
	tools.SetDataDir(filepath.Join(root, "data"))
	if tools.FFmpeg() != "" {
		t.Fatal("nothing is installed yet")
	}
	touch(t, filepath.Join(root, "bin", ExeName("ffmpeg")))
	if tools.FFmpeg() != "" {
		t.Error("the first answer is cached until Reset")
	}
	tools.Reset()
	if tools.FFmpeg() == "" {
		t.Error("after Reset the new install should be found")
	}
}

func TestSetDataDirSearchesAgain(t *testing.T) {
	isolate(t)
	a, b := t.TempDir(), t.TempDir()
	touch(t, filepath.Join(b, "bin", ExeName("ffmpeg")))
	tools := NewTools()
	tools.SetDataDir(filepath.Join(a, "data"))
	if tools.FFmpeg() != "" {
		t.Fatal("not installed under a")
	}
	tools.SetDataDir(filepath.Join(b, "data"))
	if tools.FFmpeg() == "" {
		t.Error("changing the data directory should search again")
	}
}

func TestExeName(t *testing.T) {
	got := ExeName("ffmpeg")
	if got != "ffmpeg" && got != "ffmpeg.exe" {
		t.Errorf("ExeName = %q", got)
	}
}

func TestVersionReadsFirstLineAndCaches(t *testing.T) {
	bin := useFake(t, "version")
	tools := NewTools()
	want := "ffmpeg version 9.9-test Copyright (c) the test authors"
	if got := tools.Version(bin); got != want {
		t.Fatalf("Version = %q, want %q", got, want)
	}
	useFake(t, "convert-fail") // would give a different answer if asked again
	if got := tools.Version(bin); got != want {
		t.Errorf("the version should be cached, got %q", got)
	}
	tools.Reset()
	if got := tools.Version(bin); got == want {
		t.Error("Reset should clear the cached version")
	}
}

func TestVersionOfNothing(t *testing.T) {
	tools := NewTools()
	if got := tools.Version(""); got != "" {
		t.Errorf("Version(\"\") = %q, want empty", got)
	}
	if got := tools.Version(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Errorf("a program that cannot run should give empty, got %q", got)
	}
}
