// Package cache knows where the program keeps generated files: posters, backdrops,
// converted movies and subtitle tracks. Everything lives under one folder inside the
// data directory, never in the movie library, so the scanner cannot mistake it for a movie.
package cache

import (
	"os"
	"path/filepath"
)

// Image kinds, also the folder names under the cache root.
const (
	Posters   = "posters"
	Backdrops = "backdrops"
)

// Layout maps ids to file paths under Root.
type Layout struct{ Root string }

// New returns the layout for <dataDir>/cache.
func New(dataDir string) Layout { return Layout{Root: filepath.Join(dataDir, "cache")} }

// Image is the cached JPEG for a movie. kind is Posters or Backdrops. The id is reduced to its
// base name, so it can never point outside the folder.
func (l Layout) Image(kind, id string) string {
	return filepath.Join(l.Root, kind, filepath.Base(id)+".jpg")
}

// ImageDir is the folder holding one kind of image.
func (l Layout) ImageDir(kind string) string { return filepath.Join(l.Root, kind) }

// Compat is where the converted MP4 copy of a movie lives. Whether a movie is converted
// is simply whether this file exists.
func (l Layout) Compat(id string) string {
	return filepath.Join(l.Root, "compat", filepath.Base(id)+".mp4")
}

// SubsDir is the folder with a movie's uploaded and extracted subtitle tracks.
func (l Layout) SubsDir(id string) string {
	return filepath.Join(l.Root, "subs", filepath.Base(id))
}

// Ensure creates the folders the program writes into.
func (l Layout) Ensure() error {
	for _, d := range []string{Posters, Backdrops, "compat"} {
		if err := os.MkdirAll(filepath.Join(l.Root, d), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// CleanPartialConversions removes half written conversions left by a crash or power cut.
func (l Layout) CleanPartialConversions() {
	parts, _ := filepath.Glob(filepath.Join(l.Root, "compat", "*.part.mp4"))
	for _, p := range parts {
		_ = os.Remove(p)
	}
}
