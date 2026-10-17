package media

import (
	"path/filepath"
	"strings"
)

// InsideRoots resolves symlinks and reports whether the file lives under one of the
// library folders. This is what stops a link inside the library from exposing other files.
func InsideRoots(path string, roots []string) bool {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	for _, root := range roots {
		rr, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(rr, real)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
