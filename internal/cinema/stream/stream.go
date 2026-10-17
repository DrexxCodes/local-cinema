// Package stream sends movie files to players and remembers where each room stopped.
package stream

import (
	"net/http"
	"os"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/compat"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/media"
	"local-cinema/internal/cinema/sessions"
	"local-cinema/internal/cinema/store"
)

// Service streams movies.
type Service struct {
	st       *store.Store
	compat   *compat.Service
	sessions *sessions.Service
}

// New returns the streaming service.
func New(st *store.Store, c *compat.Service, sess *sessions.Service) *Service {
	return &Service{st: st, compat: c, sessions: sess}
}

// Mount registers the guest endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.GET("/stream/{id}", rt.Guest(s.handleStream))
	rt.Public.GET("/api/playback", rt.Guest(s.handlePlaybackList))
	rt.Public.GET("/api/playback/{id}", rt.Guest(s.handlePlaybackGet))
	rt.Public.PUT("/api/playback/{id}", rt.Guest(s.handlePlaybackPut))
}

// IsStreamStart tells us whether a request begins a viewing: no Range header, or one that starts
// at the first byte. Players then ask for later ranges, which definitelyyyyyyyyyy must not count again.
func IsStreamStart(rangeHeader string) bool {
	return rangeHeader == "" || strings.HasPrefix(rangeHeader, "bytes=0-")
}

// Source picks the file to send and its container. Browsers get the converted copy when one
// exists; other apps handle most formats themselves and get the original.
func Source(original, container, converted string, external bool) (path, ctr string, usedConverted bool) {
	if !external && converted != "" {
		return converted, "mp4", true
	}
	return original, container, false
}

func (s *Service) handleStream(e *core.RequestEvent, user, sess *core.Record) error {
	m, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil || !store.Streamable(m) {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	external := sess.GetString("kind") == "external"
	if external {
		go s.sessions.KeepAlive(e.Request.Context(), sess) // one request can last the WHOLE movie
	}
	converted := ""
	if s.compat.Exists(m.Id) {
		converted = s.compat.File(m.Id)
	}
	path, container, usedConverted := Source(m.GetString("file_path"), m.GetString("container"), converted, external)
	if !usedConverted && !s.st.InsideLibrary(path) {
		s.st.Audit("path_escape", user.Id, sess.Id, m.Id, map[string]any{"path": path})
		return httpx.Err(e, 403, "forbidden", "Forbidden.")
	}
	f, err := os.Open(path)
	if err != nil {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not read the file.")
	}
	if IsStreamStart(e.Request.Header.Get("Range")) {
		s.st.Audit("stream_start", user.Id, sess.Id, m.Id, map[string]any{"ip": e.RealIP()})
	}
	h := e.Response.Header()
	h.Set("Content-Type", media.ContentType(container))
	h.Set("Accept-Ranges", "bytes")
	h.Set("Cache-Control", "private, max-age=0")
	http.ServeContent(e.Response, e.Request, st.Name(), st.ModTime(), f)
	return nil
}

func (s *Service) findPlayback(userID, movieID string) *core.Record {
	r, err := s.st.App.FindFirstRecordByFilter("playback_state", "user = {:u} && movie = {:m}",
		dbx.Params{"u": userID, "m": movieID})
	if err != nil {
		return nil
	}
	return r
}

func (s *Service) handlePlaybackList(e *core.RequestEvent, user, _ *core.Record) error {
	recs, _ := s.st.App.FindAllRecords("playback_state", dbx.HashExp{"user": user.Id})
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		out = append(out, map[string]any{
			"movie":            r.GetString("movie"),
			"position_seconds": r.GetFloat("position_seconds"),
			"completed":        r.GetBool("completed"),
			"updated_at":       r.GetDateTime("updated").String(),
		})
	}
	return e.JSON(200, out)
}

func (s *Service) handlePlaybackGet(e *core.RequestEvent, user, _ *core.Record) error {
	r := s.findPlayback(user.Id, e.Request.PathValue("id"))
	if r == nil {
		return e.JSON(200, map[string]any{"position_seconds": 0, "completed": false})
	}
	return e.JSON(200, map[string]any{
		"position_seconds": r.GetFloat("position_seconds"),
		"completed":        r.GetBool("completed"),
	})
}

func (s *Service) handlePlaybackPut(e *core.RequestEvent, user, _ *core.Record) error {
	var in struct {
		Position  float64 `json:"position_seconds"`
		Completed bool    `json:"completed"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	movieID := e.Request.PathValue("id")
	if _, err := s.st.App.FindRecordById("movies", movieID); err != nil {
		return httpx.Err(e, 404, "not_found", "Unknown movie.")
	}
	r := s.findPlayback(user.Id, movieID)
	if r == nil {
		c, _ := s.st.App.FindCollectionByNameOrId("playback_state")
		r = core.NewRecord(c)
		r.Set("user", user.Id)
		r.Set("movie", movieID)
	}
	r.Set("position_seconds", in.Position)
	r.Set("completed", in.Completed)
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not save progress.")
	}
	return httpx.OK(e)
}
