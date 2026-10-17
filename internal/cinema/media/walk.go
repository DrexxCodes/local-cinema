package media

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FindVideos lists every video file under the given folders. Unreadable folders are skipped.
func FindVideos(roots []string) []string {
	var files []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && IsVideo(p) {
				files = append(files, p)
			}
			return nil
		})
	}
	return files
}

// FolderReport says what a library folder holds. The JSON names are what the UI reads.
type FolderReport struct {
	Valid      bool     `json:"valid"`
	VideoCount int      `json:"video_count"`
	TotalBytes int64    `json:"total_bytes"`
	Warnings   []string `json:"warnings"`
}

// ValidateFolder inspects a library folder: does it exist and how many videos does it hold.
func ValidateFolder(path string) FolderReport {
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		return FolderReport{Warnings: []string{"That folder does not exist or is not readable."}}
	}
	r := FolderReport{Valid: true, Warnings: []string{}}
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && IsVideo(p) {
			r.VideoCount++
			if info, err := d.Info(); err == nil {
				r.TotalBytes += info.Size()
			}
		}
		return nil
	})
	if r.VideoCount == 0 {
		r.Warnings = append(r.Warnings, "No videos found in this folder.")
	}
	return r
}
