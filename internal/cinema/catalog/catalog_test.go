package catalog

import (
	"strings"
	"testing"
)

func TestTitle(t *testing.T) {
	if got, err := Title("  Heat  "); err != nil || got != "Heat" {
		t.Errorf("got %q, %v", got, err)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("a", 201)} {
		if _, err := Title(bad); err == nil {
			t.Errorf("%d chars should be refused", len(bad))
		}
	}
	if _, err := Title(strings.Repeat("a", 200)); err != nil {
		t.Error("200 characters is allowed")
	}
}

func TestYear(t *testing.T) {
	for _, ok := range []int{0, 1999, 2100} {
		if Year(ok) != nil {
			t.Errorf("%d should be accepted", ok)
		}
	}
	for _, bad := range []int{-1, 2101} {
		if Year(bad) == nil {
			t.Errorf("%d should be refused", bad)
		}
	}
}

func TestSynopsis(t *testing.T) {
	if got, err := Synopsis(" hi \n"); err != nil || got != "hi" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, err := Synopsis(strings.Repeat("x", 5001)); err == nil {
		t.Error("5001 characters should be refused")
	}
}

func TestGenres(t *testing.T) {
	got := Genres([]string{" Drama ", "", "  ", strings.Repeat("g", 41), "Crime"})
	if len(got) != 2 || got[0] != "Drama" || got[1] != "Crime" {
		t.Errorf("got %v", got)
	}
	many := make([]string, 30)
	for i := range many {
		many[i] = "g"
	}
	if n := len(Genres(many)); n != 20 {
		t.Errorf("at most 20 genres, got %d", n)
	}
	if got := Genres(nil); got == nil || len(got) != 0 {
		t.Error("an empty list must be an empty slice, not nil (it becomes JSON [])")
	}
}

func TestRating(t *testing.T) {
	if Rating(-3) != 0 || Rating(11) != 10 || Rating(7.5) != 7.5 {
		t.Error("rating is clamped to 0..10")
	}
}

func TestArtKind(t *testing.T) {
	if d, w, ok := ArtKind("poster"); !ok || d != "posters" || w != 600 {
		t.Errorf("poster: %q %d %v", d, w, ok)
	}
	if d, w, ok := ArtKind("backdrop"); !ok || d != "backdrops" || w != 1920 {
		t.Errorf("backdrop: %q %d %v", d, w, ok)
	}
	if _, _, ok := ArtKind("../x"); ok {
		t.Error("unknown kinds are refused")
	}
}

func TestImageExt(t *testing.T) {
	for ct, ext := range map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"} {
		if got, ok := ImageExt(ct); !ok || got != ext {
			t.Errorf("%s: %q %v", ct, got, ok)
		}
	}
	if _, ok := ImageExt("image/gif"); ok {
		t.Error("gif is not accepted")
	}
	if _, ok := ImageExt("text/html"); ok {
		t.Error("html is not accepted")
	}
}
