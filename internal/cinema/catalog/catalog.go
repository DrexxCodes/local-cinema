// Package catalog is the movie list: what guests browse, and what the admin edits.
package catalog

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/compat"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/scanner"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/util"
)

// Service serves and edits movies.
type Service struct {
	st     *store.Store
	compat *compat.Service
	meta   *scanner.Metadata
}

// New returns the catalog service.
func New(st *store.Store, c *compat.Service, meta *scanner.Metadata) *Service {
	return &Service{st: st, compat: c, meta: meta}
}

// Mount registers the guest and admin endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.GET("/api/movies", rt.Guest(s.handleList))
	rt.Public.GET("/api/movies/{id}", rt.Guest(s.handleGet))
	rt.Public.GET("/api/movies/{id}/poster", rt.Guest(s.imageHandler("posters")))
	rt.Public.GET("/api/movies/{id}/backdrop", rt.Guest(s.imageHandler("backdrops")))
	// The admin page uses <img>, which cannot send headers, so its token comes in the query.
	rt.Public.GET("/api/admin-img/{kind}/{id}", s.handleAdminImage)

	rt.Admin.GET("/movies", s.handleAdminList)
	rt.Admin.PATCH("/movies/{id}", s.handleAdminUpdate)
	rt.Admin.DELETE("/movies/{id}", s.handleAdminDelete)
	rt.Admin.POST("/movies/{id}/tmdb", s.handleAdminLookup)
	rt.Admin.PUT("/movies/{id}/art/{kind}", s.handleAdminArt)
	rt.Admin.POST("/movies/visibility", s.handleAdminVisibility)
}

// DTO is the guest view of a movie.
func DTO(r *core.Record) map[string]any {
	var genres []string
	_ = r.UnmarshalJSONField("genres", &genres)
	if genres == nil {
		genres = []string{}
	}
	return map[string]any{
		"id":               r.Id,
		"title":            r.GetString("title"),
		"year":             int(r.GetFloat("year")),
		"duration_seconds": int(r.GetFloat("duration_seconds")),
		"container":        r.GetString("container"),
		"video_codec":      r.GetString("video_codec"),
		"audio_codec":      r.GetString("audio_codec"),
		"resolution":       r.GetString("resolution"),
		"synopsis":         r.GetString("synopsis"),
		"genres":           genres,
		"rating":           r.GetFloat("rating"),
		"file_size":        int64(r.GetFloat("file_size")),
		"has_poster":       r.GetBool("has_poster"),
		"has_backdrop":     r.GetBool("has_backdrop"),
		"added_at":         r.GetDateTime("added_at").String(),
		"thumb_v":          int64(r.GetFloat("thumb_version")),
	}
}

// AdminDTO is the guest view plus what only the admin needs.
func (s *Service) AdminDTO(r *core.Record) map[string]any {
	d := DTO(r)
	d["compat"] = s.compat.Status(r)
	d["hidden"] = r.GetBool("hidden")
	d["edited"] = r.GetBool("edited")
	d["available"] = r.GetBool("available")
	d["tmdb_id"] = int(r.GetFloat("tmdb_id"))
	d["file_name"] = filepath.Base(r.GetString("file_path"))
	return d
}

func (s *Service) handleList(e *core.RequestEvent, _, _ *core.Record) error {
	recs, err := s.st.App.FindAllRecords("movies", dbx.HashExp{"available": true}, dbx.NewExp("COALESCE(hidden, 0) = 0"))
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not load the catalog.")
	}
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		out = append(out, DTO(r))
	}
	return e.JSON(200, out)
}

func (s *Service) handleGet(e *core.RequestEvent, _, _ *core.Record) error {
	r, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil || !store.Streamable(r) {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	return e.JSON(200, DTO(r))
}

func (s *Service) imageHandler(kind string) httpx.GuestHandler {
	return func(e *core.RequestEvent, _, _ *core.Record) error {
		return s.serveImage(e, kind, e.Request.PathValue("id"))
	}
}

// serveImage sends a cached poster or backdrop. The version query parameter (?v=) changes
// whenever a thumbnail is regenerated, so the browser can keep it for a week.
func (s *Service) serveImage(e *core.RequestEvent, kind, id string) error {
	p := s.st.Cache.Image(kind, id) // reduced to a base name, no path tricks
	if !util.FileExists(p) {
		return httpx.Err(e, 404, "not_found", "No image.")
	}
	e.Response.Header().Set("Cache-Control", "private, max-age=604800")
	return e.FileFS(os.DirFS(filepath.Dir(p)), filepath.Base(p))
}

func (s *Service) handleAdminImage(e *core.RequestEvent) error {
	tok := e.Request.URL.Query().Get("token")
	if tok == "" {
		tok = strings.TrimPrefix(e.Request.Header.Get("Authorization"), "Bearer ")
	}
	rec, err := s.st.App.FindAuthRecordByToken(tok, core.TokenTypeAuth)
	if err != nil || rec.Collection().Name != core.CollectionNameSuperusers {
		return httpx.Err(e, 401, "unauthorized", "Admin sign in required.")
	}
	kind := e.Request.PathValue("kind")
	if kind != "posters" && kind != "backdrops" {
		return httpx.Err(e, 404, "not_found", "No image.")
	}
	return s.serveImage(e, kind, e.Request.PathValue("id"))
}

// handleAdminList lists the library a page at a time.
// Query: limit, offset, q (title), show = "" | "hidden" | "missing".
func (s *Service) handleAdminList(e *core.RequestEvent) error {
	qs := e.Request.URL.Query()
	limit := util.QueryInt(qs.Get("limit"), 24, 1, 100)
	offset := util.QueryInt(qs.Get("offset"), 0, 0, 1<<30)
	q := strings.TrimSpace(qs.Get("q"))

	filter := "available = true"
	counts := []dbx.Expression{dbx.HashExp{"available": true}}
	switch qs.Get("show") {
	case "hidden":
		filter = "available = true && hidden = true"
		counts = []dbx.Expression{dbx.HashExp{"available": true, "hidden": true}}
	case "missing":
		filter = "available = false"
		counts = []dbx.Expression{dbx.HashExp{"available": false}}
	}
	params := dbx.Params{}
	if q != "" {
		filter += " && title ~ {:q}"
		params["q"] = q
		counts = append(counts, dbx.Like("title", q))
	}
	total, err := s.st.App.CountRecords("movies", counts...)
	if err != nil {
		slog.Warn("admin movies: count", "err", err)
		return httpx.Err(e, 500, "internal", "Could not count movies: "+err.Error())
	}
	recs, err := s.st.App.FindRecordsByFilter("movies", filter, "title", limit, offset, params)
	if err != nil {
		slog.Warn("admin movies: list", "err", err)
		return httpx.Err(e, 500, "internal", "Could not list movies: "+err.Error())
	}
	items := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		items = append(items, s.AdminDTO(r))
	}
	return e.JSON(200, map[string]any{"items": items, "total": total, "has_more": int64(offset+len(items)) < total})
}

// handleAdminUpdate edits the details TMDB would otherwise fill in. Editing any of them
// locks the movie so a rescan never overwrites the admin's text.
func (s *Service) handleAdminUpdate(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Movie not found.")
	}
	var in struct {
		Title    *string   `json:"title"`
		Year     *int      `json:"year"`
		Synopsis *string   `json:"synopsis"`
		Genres   *[]string `json:"genres"`
		Rating   *float64  `json:"rating"`
		Hidden   *bool     `json:"hidden"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	edited := false
	if in.Title != nil {
		t, err := Title(*in.Title)
		if err != nil {
			return httpx.Err(e, 400, "invalid", err.Error())
		}
		r.Set("title", t)
		edited = true
	}
	if in.Year != nil {
		if err := Year(*in.Year); err != nil {
			return httpx.Err(e, 400, "invalid", err.Error())
		}
		r.Set("year", *in.Year)
		edited = true
	}
	if in.Synopsis != nil {
		t, err := Synopsis(*in.Synopsis)
		if err != nil {
			return httpx.Err(e, 400, "invalid", err.Error())
		}
		r.Set("synopsis", t)
		edited = true
	}
	if in.Genres != nil {
		r.Set("genres", Genres(*in.Genres))
		edited = true
	}
	if in.Rating != nil {
		r.Set("rating", Rating(*in.Rating))
		edited = true
	}
	if edited {
		r.Set("edited", true)
	}
	if in.Hidden != nil {
		r.Set("hidden", *in.Hidden)
	}
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 400, "invalid", "Could not save: "+err.Error())
	}
	return e.JSON(200, s.AdminDTO(r))
}

// handleAdminVisibility hides or shows several movies at once.
// Body: {"ids":["..."],"hidden":true}
func (s *Service) handleAdminVisibility(e *core.RequestEvent) error {
	var in struct {
		IDs    []string `json:"ids"`
		Hidden bool     `json:"hidden"`
	}
	if err := e.BindBody(&in); err != nil || len(in.IDs) == 0 {
		return httpx.Err(e, 400, "bad_request", "Pick at least one movie.")
	}
	recs, err := s.st.App.FindRecordsByIds("movies", in.IDs)
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not load the movies.")
	}
	for _, r := range recs {
		r.Set("hidden", in.Hidden)
		if err := s.st.App.Save(r); err != nil {
			return httpx.Err(e, 500, "internal", "Could not save: "+err.Error())
		}
	}
	return e.JSON(200, map[string]any{"ok": true, "changed": len(recs)})
}

// handleAdminDelete forgets a movie whose file is gone. A movie with a file on disk cannot be
// deleted (a rescan would add it back): hide it instead.
func (s *Service) handleAdminDelete(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Movie not found.")
	}
	if r.GetBool("available") {
		return httpx.Err(e, 400, "in_use", "This movie's file still exists. Hide it from guests instead.")
	}
	for _, p := range []string{
		s.st.Cache.Image("posters", r.Id),
		s.st.Cache.Image("backdrops", r.Id),
		s.compat.File(r.Id),
	} {
		_ = os.Remove(p)
	}
	_ = os.RemoveAll(s.st.Cache.SubsDir(r.Id))
	if err := s.st.App.Delete(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not remove it: "+err.Error())
	}
	return httpx.OK(e)
}

// handleAdminLookup fills the details from TMDB using the movie's current title and year,
// so fixing a wrong title and looking up again works.
func (s *Service) handleAdminLookup(e *core.RequestEvent) error {
	key := s.meta.Key()
	if key == "" {
		return httpx.Err(e, 400, "no_key", "Add a TMDB API key in Settings first.")
	}
	r, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Movie not found.")
	}
	r.Set("tmdb_id", 0)
	s.meta.Fetch(r, key)
	if r.GetFloat("tmdb_id") == 0 {
		return httpx.Err(e, 404, "no_match", fmt.Sprintf("TMDB found nothing for \"%s\" (or the internet is unreachable). Fix the title or year and try again.", r.GetString("title")))
	}
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not save: "+err.Error())
	}
	return e.JSON(200, s.AdminDTO(r))
}
