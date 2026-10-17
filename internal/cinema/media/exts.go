package media

import (
	"path/filepath"
	"strings"
)

// videoExts are the formats the scanner lists. The first six play directly in most
// browsers. The rest are listed too and become playable through "Make compatible".
var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".mov": true, ".avi": true, ".webm": true, ".m4v": true,
	".wmv": true, ".flv": true, ".mpg": true, ".mpeg": true, ".ts": true, ".m2ts": true, ".3gp": true, ".ogv": true,
}

// IsVideoExt reports whether ext (with its dot, any case) is a video format we list.
func IsVideoExt(ext string) bool { return videoExts[strings.ToLower(ext)] }

// IsVideo reports whether the file name has a video extension.
func IsVideo(path string) bool { return IsVideoExt(filepath.Ext(path)) }

var contentTypes = map[string]string{
	"mp4": "video/mp4", "m4v": "video/x-m4v", "mov": "video/quicktime",
	"webm": "video/webm", "mkv": "video/x-matroska", "avi": "video/x-msvideo",
	"wmv": "video/x-ms-wmv", "flv": "video/x-flv", "mpg": "video/mpeg", "mpeg": "video/mpeg",
	"ts": "video/mp2t", "m2ts": "video/mp2t", "3gp": "video/3gpp", "ogv": "video/ogg",
}

// ContentType is the HTTP content type for a container name such as "mkv".
func ContentType(container string) string {
	if ct := contentTypes[container]; ct != "" {
		return ct
	}
	return "application/octet-stream"
}
