// Package subtitles finds, converts, stores and serves subtitle tracks: files next to the
// movie, files the admin uploads, and tracks extracted from inside the video.
package subtitles

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/progress"
	"local-cinema/internal/cinema/store"
)

// Service owns the subtitle index files and the extraction job.
type Service struct {
	st *store.Store
	mu sync.Mutex // guards every index.json

	// Extract is the progress of "Extract from video".
	Extract progress.Tracker

	noteMu sync.Mutex
	note   string
}

// New returns the subtitle service.
func New(st *store.Store) *Service { return &Service{st: st} }

// Note is the message left by the last extraction ("Extracted 2 subtitle tracks.").
func (s *Service) Note() string {
	s.noteMu.Lock()
	defer s.noteMu.Unlock()
	return s.note
}

func (s *Service) setNote(msg string) {
	s.noteMu.Lock()
	s.note = msg
	s.noteMu.Unlock()
}

// Tracks returns every subtitle track of a movie, with the default marked.
func (s *Service) Tracks(m *core.Record) []Track {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := FindSidecars(m.GetString("file_path"))
	idx := s.loadIndex(m.Id)
	for _, t := range idx.Tracks {
		p := s.trackFile(m.Id, t.Key)
		if _, err := os.Stat(p); err == nil {
			out = append(out, Track{Key: t.Key, Lang: t.Lang, Label: t.Label, Source: t.Source, path: p})
		}
	}
	for i := range out {
		out[i].Default = out[i].Key == idx.Default
	}
	if out == nil {
		out = []Track{}
	}
	return out
}

// Mount registers the subtitle endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.GET("/api/movies/{id}/subtitles", rt.Guest(s.handleList))
	rt.Public.GET("/api/movies/{id}/subtitles/{lang}", rt.Guest(s.handleTrack))

	rt.Admin.GET("/movies/{id}/subtitles", s.handleAdminList)
	rt.Admin.POST("/movies/{id}/subtitles", s.handleAdminUpload)
	rt.Admin.POST("/movies/{id}/subtitles/extract", s.handleAdminExtract)
	rt.Admin.PATCH("/movies/{id}/subtitles/{key}", s.handleAdminUpdate)
	rt.Admin.DELETE("/movies/{id}/subtitles/{key}", s.handleAdminDelete)
}

func (s *Service) handleList(e *core.RequestEvent, _, _ *core.Record) error {
	m, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil || !store.Streamable(m) {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	return e.JSON(200, s.Tracks(m))
}

func (s *Service) handleTrack(e *core.RequestEvent, _, _ *core.Record) error {
	m, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil || !store.Streamable(m) {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	key := e.Request.PathValue("lang") // the route keeps its old name, the value is the track key
	for _, t := range s.Tracks(m) {
		if t.Key != key {
			continue
		}
		raw, err := os.ReadFile(t.path)
		if err != nil {
			break
		}
		txt := strings.TrimPrefix(string(raw), "\ufeff")
		txt = strings.ReplaceAll(txt, "\r\n", "\n")
		if strings.ToLower(filepath.Ext(t.path)) == ".srt" {
			txt = SRTToVTT(txt)
		}
		h := e.Response.Header()
		h.Set("Content-Type", "text/vtt; charset=utf-8")
		h.Set("Cache-Control", "private, max-age=300")
		h.Set("Access-Control-Allow-Origin", "*") // external players load these links
		_, _ = e.Response.Write([]byte(txt))
		return nil
	}
	return httpx.Err(e, 404, "not_found", "No subtitles for that track.")
}
