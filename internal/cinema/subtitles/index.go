package subtitles

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A subtitle track comes from one of three places:
//
//	file      a .srt or .vtt next to the movie (Movie.srt, Movie.en.srt ...), never modified
//	upload    added by the admin, stored in the cache folder
//	embedded  extracted by the admin from a track inside the video file
const (
	SourceFile     = "file"
	SourceUpload   = "upload"
	SourceEmbedded = "embedded"
)

// Track is one subtitle track as the player sees it.
type Track struct {
	Key     string `json:"key"`
	Lang    string `json:"lang"`
	Label   string `json:"label"`
	Default bool   `json:"default"`
	Source  string `json:"source"`
	path    string
}

// Entry is a stored (uploaded or extracted) track.
type Entry struct {
	Key    string `json:"key"`
	Lang   string `json:"lang"`
	Label  string `json:"label"`
	Source string `json:"source"`
}

// Index is saved as <cache>/subs/<movie id>/index.json.
type Index struct {
	Default string  `json:"default"` // key of the track that starts selected, empty for none
	Tracks  []Entry `json:"tracks"`
}

var langRe = regexp.MustCompile(`^[A-Za-z0-9\-]{1,16}$`)

// ValidLang reports whether s is usable as a language code in a label or file name.
func ValidLang(s string) bool { return langRe.MatchString(s) }

// randKey makes a short random track key such as "u3fa09c".
func randKey(prefix string) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 6)
	for i := range b {
		b[i] = hex[rand.Intn(16)]
	}
	return prefix + string(b)
}

// position finds a stored track by key, -1 when it is not stored.
func (idx *Index) position(key string) int {
	for i, t := range idx.Tracks {
		if t.Key == key {
			return i
		}
	}
	return -1
}

// setDefault makes key the track that starts selected, or clears it when on is false and key
// is the current default.
func (idx *Index) setDefault(key string, on bool) {
	if on {
		idx.Default = key
	} else if idx.Default == key {
		idx.Default = ""
	}
}

// remove deletes a stored track and reports whether it existed. The default is cleared when it
// pointed at the removed track.
func (idx *Index) remove(key string) bool {
	pos := idx.position(key)
	if pos < 0 {
		return false
	}
	idx.Tracks = append(idx.Tracks[:pos:pos], idx.Tracks[pos+1:]...)
	if idx.Default == key {
		idx.Default = ""
	}
	return true
}

// upsert adds a stored track, or updates the language and label of one with the same key.
func (idx *Index) upsert(e Entry) {
	if pos := idx.position(e.Key); pos >= 0 {
		idx.Tracks[pos].Lang, idx.Tracks[pos].Label = e.Lang, e.Label
		return
	}
	idx.Tracks = append(idx.Tracks, e)
}

// cleanLabel trims a track name and checks its length (1 to 40 characters).
func cleanLabel(s string) (string, bool) {
	l := strings.TrimSpace(s)
	return l, l != "" && len(l) <= 40
}

func (s *Service) loadIndex(id string) Index {
	var idx Index
	if raw, err := os.ReadFile(filepath.Join(s.st.Cache.SubsDir(id), "index.json")); err == nil {
		_ = json.Unmarshal(raw, &idx)
	}
	return idx
}

func (s *Service) saveIndex(id string, idx Index) error {
	dir := s.st.Cache.SubsDir(id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(idx, "", "  ")
	return os.WriteFile(filepath.Join(dir, "index.json"), raw, 0o644)
}

// trackFile is where a stored track's WebVTT lives.
func (s *Service) trackFile(movieID, key string) string {
	return filepath.Join(s.st.Cache.SubsDir(movieID), filepath.Base(key)+".vtt")
}

// FindSidecars looks for subtitle files next to a movie: Movie.srt, Movie.en.srt, Movie.fr.vtt ...
func FindSidecars(moviePath string) []Track {
	dir := filepath.Dir(moviePath)
	base := strings.TrimSuffix(filepath.Base(moviePath), filepath.Ext(moviePath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []Track
	for _, en := range entries {
		if en.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(en.Name()))
		if ext != ".srt" && ext != ".vtt" {
			continue
		}
		stem := strings.TrimSuffix(en.Name(), filepath.Ext(en.Name()))
		lang := ""
		switch {
		case stem == base:
			lang = "und"
		case strings.HasPrefix(stem, base+"."):
			lang = stem[len(base)+1:]
		default:
			continue
		}
		if !langRe.MatchString(lang) {
			continue
		}
		label := lang
		if lang == "und" {
			label = "Subtitles"
		}
		out = append(out, Track{Key: lang, Lang: lang, Label: label, Source: SourceFile, path: filepath.Join(dir, en.Name())})
	}
	return out
}
