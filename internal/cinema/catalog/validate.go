package catalog

import (
	"errors"
	"strings"

	"local-cinema/internal/cinema/util"
)

// Title checks and trims a movie title (1 to 200 characters).
func Title(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 200 {
		return "", errors.New("The title must be 1 to 200 characters.")
	}
	return s, nil
}

// Year accepts 0 (unknown) or a year up to 2100.
func Year(y int) error {
	if y < 0 || y > 2100 {
		return errors.New("Enter a year between 1888 and 2100, or 0 for unknown.")
	}
	return nil
}

// Synopsis trims the text and refuses more than 5000 characters.
func Synopsis(s string) (string, error) {
	if len(s) > 5000 {
		return "", errors.New("The synopsis can be at most 5000 characters.")
	}
	return strings.TrimSpace(s), nil
}

// Genres keeps up to 20 non empty names of at most 40 characters.
func Genres(in []string) []string {
	clean := []string{}
	for _, g := range in {
		if g = strings.TrimSpace(g); g != "" && len(g) <= 40 && len(clean) < 20 {
			clean = append(clean, g)
		}
	}
	return clean
}

// Rating limits a rating to 0..10.
func Rating(r float64) float64 { return util.ClampFloat(r, 0, 10) }

// ArtKind maps the URL part "poster" or "backdrop" to the cache folder and the widest size
// kept. ok is false for anything else.
func ArtKind(kind string) (dir string, maxWidth int, ok bool) {
	switch kind {
	case "poster":
		return "posters", 600, true
	case "backdrop":
		return "backdrops", 1920, true
	}
	return "", 0, false
}

// ImageExt maps a sniffed content type to the extension of an accepted upload.
func ImageExt(contentType string) (string, bool) {
	ext, ok := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
	return ext, ok
}
