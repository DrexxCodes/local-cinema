package tools

import (
	"context"
	"strings"
	"testing"

	"local-cinema/internal/cinema/ffmpeg"
)

func TestInstallFailureSetsMessageAndFinishes(t *testing.T) {
	s := New(ffmpeg.NewTools(), &ffmpeg.Installer{GOOS: "plan9", GOARCH: "mips"})
	s.Install(context.Background())
	if !strings.HasPrefix(s.Message(), "Could not install FFmpeg:") {
		t.Errorf("message = %q", s.Message())
	}
	if s.Job.Running() {
		t.Error("the job must end after a failure")
	}
}

func TestInstallSkipsWhenBusy(t *testing.T) {
	s := New(ffmpeg.NewTools(), &ffmpeg.Installer{GOOS: "plan9"})
	s.Job.Begin()
	s.Install(context.Background())
	if s.Message() != "" {
		t.Errorf("a second install must not run, message = %q", s.Message())
	}
	if !s.Job.Running() {
		t.Error("the running job must be left alone")
	}
}

func TestStatusShape(t *testing.T) {
	s := New(ffmpeg.NewTools(), &ffmpeg.Installer{Repo: "o/r", Bundle: "1", GOOS: "linux", GOARCH: "amd64"})
	got := s.Status()
	for _, k := range []string{"ffmpeg", "ffprobe", "ffmpeg_version", "ffmpeg_path", "can_install", "install", "message", "platform"} {
		if _, ok := got[k]; !ok {
			t.Errorf("status is missing %q", k)
		}
	}
	if got["can_install"] != true {
		t.Error("linux/amd64 with a repo and bundle can install")
	}
}

func TestHooksFeedTheTracker(t *testing.T) {
	s := New(ffmpeg.NewTools(), &ffmpeg.Installer{})
	s.Job.Begin()
	s.Installer.OnStep("Downloading FFmpeg")
	s.Installer.OnPercent(40)
	snap := s.Job.Snapshot()
	if snap.Current != "Downloading FFmpeg" || snap.Percent != 40 {
		t.Errorf("snapshot = %+v", snap)
	}
}
