package version

import "testing"

func TestDefaultsAreDevelopmentBuild(t *testing.T) {
	// A plain `go build` has no -ldflags, which must read as a development build so the
	// update check and the FFmpeg download stay off.
	if Version != "dev" {
		t.Errorf("Version = %q, want dev", Version)
	}
	if Repo != "" || FFmpegBundle != "" {
		t.Error("Repo and FFmpegBundle are only set by the release build")
	}
}
