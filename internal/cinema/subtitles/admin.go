package subtitles

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/util"
)

const maxUpload = 2 << 20 // 2 MB

// adminDTO lists a movie's tracks for the admin page. Files next to the movie can't be edited.
func (s *Service) adminDTO(m *core.Record) []map[string]any {
	out := []map[string]any{}
	for _, t := range s.Tracks(m) {
		out = append(out, map[string]any{
			"key": t.Key, "lang": t.Lang, "label": t.Label, "default": t.Default,
			"source": t.Source, "editable": t.Source != SourceFile,
		})
	}
	return out
}

// movieFor loads the movie named in the URL. When it is missing the 404 has been sent
// already, so the caller returns the error as it is.
func (s *Service) movieFor(e *core.RequestEvent) (*core.Record, error) {
	m, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil {
		return nil, httpx.Err(e, 404, "not_found", "Movie not found.")
	}
	return m, nil
}

func (s *Service) handleAdminList(e *core.RequestEvent) error {
	m, err := s.movieFor(e)
	if m == nil {
		return err
	}
	return e.JSON(200, s.adminDTO(m))
}

// handleAdminUpload adds a subtitle file (raw body, 2 MB max).
// Query: lang (for example en), label (shown in the player), name (original file name, gives the format).
func (s *Service) handleAdminUpload(e *core.RequestEvent) error {
	m, err := s.movieFor(e)
	if m == nil {
		return err
	}
	q := e.Request.URL.Query()
	lang := strings.TrimSpace(q.Get("lang"))
	if !ValidLang(lang) {
		return httpx.Err(e, 400, "invalid", "Pick a language for these subtitles.")
	}
	label := strings.TrimSpace(q.Get("label"))
	if label == "" {
		label = lang
	}
	if len(label) > 40 {
		label = label[:40]
	}
	raw, err := io.ReadAll(http.MaxBytesReader(e.Response, e.Request.Body, maxUpload))
	if err != nil {
		return httpx.Err(e, 413, "too_large", "That file is too large (2 MB at most).")
	}
	vtt, err := ToVTT(raw, strings.ToLower(filepath.Ext(q.Get("name"))), s.st.Tools.FFmpeg())
	if err != nil {
		return httpx.Err(e, 400, "bad_subtitle", err.Error())
	}
	if err := s.addUpload(m.Id, lang, label, vtt); err != nil {
		return httpx.Err(e, 500, "internal", "Could not save the subtitles.")
	}
	return e.JSON(200, s.adminDTO(m))
}

// addUpload stores a converted upload and lists it in the movie's index.
func (s *Service) addUpload(movieID, lang, label string, vtt []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.loadIndex(movieID)
	key := randKey("u")
	if err := os.MkdirAll(s.st.Cache.SubsDir(movieID), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(s.trackFile(movieID, key), vtt, 0o644); err != nil {
		return err
	}
	idx.Tracks = append(idx.Tracks, Entry{Key: key, Lang: lang, Label: label, Source: SourceUpload})
	return s.saveIndex(movieID, idx)
}

type updateBody struct {
	Label   *string `json:"label"`
	Default *bool   `json:"default"`
	ShiftMs *int    `json:"shift_ms"`
}

// reqError is an error that already knows its HTTP answer.
type reqError struct {
	code     int
	key, msg string
}

func (r *reqError) Error() string { return r.msg }

// update renames a stored track, makes it the default, or shifts its timing.
// Files next to the movie can only be made the default, they are never changed.
func (s *Service) update(movieID, key string, in updateBody) *reqError {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.loadIndex(movieID)
	pos := idx.position(key)
	if pos < 0 && (in.Label != nil || in.ShiftMs != nil) {
		return &reqError{400, "read_only", "Subtitle files next to the movie can't be edited here. Upload a copy to change its name or timing."}
	}
	if in.Label != nil && pos >= 0 {
		l, ok := cleanLabel(*in.Label)
		if !ok {
			return &reqError{400, "invalid", "The name must be 1 to 40 characters."}
		}
		idx.Tracks[pos].Label = l
	}
	if in.ShiftMs != nil && pos >= 0 {
		p := s.trackFile(movieID, key)
		raw, err := os.ReadFile(p)
		if err != nil {
			return &reqError{404, "not_found", "Track not found."}
		}
		_ = os.WriteFile(p, []byte(ShiftVTT(string(raw), util.ClampInt(*in.ShiftMs, -600000, 600000))), 0o644)
	}
	if in.Default != nil {
		idx.setDefault(key, *in.Default)
	}
	if err := s.saveIndex(movieID, idx); err != nil {
		return &reqError{500, "internal", "Could not save."}
	}
	return nil
}

// handleAdminUpdate: Body {"label":"English","default":true,"shift_ms":-500}.
func (s *Service) handleAdminUpdate(e *core.RequestEvent) error {
	m, err := s.movieFor(e)
	if m == nil {
		return err
	}
	var in updateBody
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	if re := s.update(m.Id, e.Request.PathValue("key"), in); re != nil {
		return httpx.Err(e, re.code, re.key, re.msg)
	}
	return e.JSON(200, s.adminDTO(m))
}

func (s *Service) handleAdminDelete(e *core.RequestEvent) error {
	m, err := s.movieFor(e)
	if m == nil {
		return err
	}
	key := e.Request.PathValue("key")
	s.mu.Lock()
	idx := s.loadIndex(m.Id)
	if !idx.remove(key) {
		s.mu.Unlock()
		return httpx.Err(e, 400, "read_only", "Only uploaded or extracted subtitles can be deleted. Remove the file from the movie folder instead.")
	}
	_ = os.Remove(s.trackFile(m.Id, key))
	err = s.saveIndex(m.Id, idx)
	s.mu.Unlock()
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not save.")
	}
	return e.JSON(200, s.adminDTO(m))
}
