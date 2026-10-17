// Package version holds build information. The release workflow sets these with
// -ldflags "-X local-cinema/internal/version.Version=1.2.0" and so on.
package version

var (
	// Version is the release number without the leading "v". "dev" for local builds.
	Version = "dev"
	// Commit is the git commit the binary was built from.
	Commit = ""
	// Repo is the GitHub "owner/name" used for update checks and the FFmpeg download.
	Repo = ""
	// FFmpegBundle names the FFmpeg download this version expects (release tag ffmpeg-<bundle>).
	FFmpegBundle = ""
)
