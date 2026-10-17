// Package thumbs makes posters and banners from a frame of the movie, for movies that have
// no art from TMDB or from the admin.
package thumbs

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/progress"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/util"
)

// Service makes thumbnails and runs the "Create missing thumbnails" job.
type Service struct {
	st *store.Store

	// Job is the progress of the thumbnail job.
	Job progress.Tracker

	// Busy tells whether another library job (a scan) is running. It is set by the code that
	// wires the packages together, because the scanner itself uses this package.
	Busy func() bool

	bad sync.Map // movie id -> true when frame extraction failed, so scans don't retry it every time
}

// New returns the thumbnail service.
func New(st *store.Store) *Service { return &Service{st: st} }

// MissingArt reports whether a movie lacks a poster or a banner.
func MissingArt(hasPoster, hasBackdrop bool) bool { return !(hasPoster && hasBackdrop) }

// Needed reports whether a movie still lacks art and is worth a try now.
func (s *Service) Needed(rec *core.Record) bool {
	if !MissingArt(rec.GetBool("has_poster"), rec.GetBool("has_backdrop")) {
		return false
	}
	if s.st.Tools.FFmpeg() == "" {
		return false
	}
	_, failed := s.bad.Load(rec.Id)
	return !failed
}

// MarkFailed remembers that no frame could be taken from a movie, so scans stop retrying.
func (s *Service) MarkFailed(id string) { s.bad.Store(id, true) }

// Make takes one frame from the video and writes a 2:3 poster and a wide backdrop to the
// cache. at <= 0 picks a position automatically. Existing art (for example from TMDB) is
// kept unless force is true. The caller saves rec.
func (s *Service) Make(rec *core.Record, at float64, force bool) error {
	bin := s.st.Tools.FFmpeg()
	if bin == "" {
		return errors.New("ffmpeg is not installed")
	}
	src := rec.GetString("file_path")
	if !s.st.InsideLibrary(src) {
		return errors.New("file is outside the library folders")
	}
	if at <= 0 {
		at = ffmpeg.PickSeek(rec.GetFloat("duration_seconds"))
	}
	wantPoster := force || !rec.GetBool("has_poster")
	wantBackdrop := force || !rec.GetBool("has_backdrop")

	err := writeFrames(bin, src, at,
		s.st.Cache.Image(cache.Posters, rec.Id), s.st.Cache.Image(cache.Backdrops, rec.Id),
		wantPoster, wantBackdrop)
	if err != nil {
		return err
	}
	if wantPoster {
		rec.Set("has_poster", true)
	}
	if wantBackdrop {
		rec.Set("has_backdrop", true)
	}
	rec.Set("thumb_version", time.Now().Unix())
	s.bad.Delete(rec.Id)
	return nil
}

// writeFrames extracts one frame to temporary files and moves the wanted pictures into
// place. A failed run leaves the existing pictures untouched.
func writeFrames(bin, src string, at float64, posterOut, backdropOut string, wantPoster, wantBackdrop bool) error {
	posterTmp := posterOut + ".tmp.jpg"
	backdropTmp := backdropOut + ".tmp.jpg"
	defer os.Remove(posterTmp)
	defer os.Remove(backdropTmp)

	if err := ffmpeg.ExtractThumbs(bin, src, at, backdropTmp, posterTmp); err != nil {
		return err
	}
	if wantPoster {
		if err := os.Rename(posterTmp, posterOut); err != nil {
			return err
		}
	}
	if wantBackdrop {
		if err := os.Rename(backdropTmp, backdropOut); err != nil {
			return err
		}
	}
	return nil
}

// RunJob makes thumbnails for the given movies, or for every movie without art when all is
// true. With force it replaces existing art too.
func (s *Service) RunJob(ids []string, all bool, at float64, force bool) {
	if !s.Job.Begin() {
		return
	}
	defer s.Job.End()

	var recs []*core.Record
	var err error
	if all {
		recs, err = s.st.App.FindAllRecords("movies", dbx.HashExp{"available": true})
	} else {
		recs, err = s.st.App.FindRecordsByIds("movies", ids)
	}
	if err != nil {
		return
	}
	todo := recs[:0]
	for _, r := range recs {
		if !r.GetBool("available") {
			continue
		}
		if force || MissingArt(r.GetBool("has_poster"), r.GetBool("has_backdrop")) {
			todo = append(todo, r)
		}
	}
	s.Job.Set(len(todo), 0, "")
	for i, r := range todo {
		s.Job.Set(-1, i, r.GetString("title"))
		if err := s.Make(r, at, force); err != nil {
			s.MarkFailed(r.Id)
			continue
		}
		_ = s.st.App.Save(r)
	}
	s.Job.Set(len(todo), len(todo), "")
}

// Mount registers the thumbnail endpoint.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.POST("/thumbnails", s.handleStart)
}

// handleStart starts a background thumbnail job.
// Body: {"ids":["..."],"all":false,"at":0,"force":true}
// "at" is the position in seconds (0 picks one automatically).
func (s *Service) handleStart(e *core.RequestEvent) error {
	var in struct {
		IDs   []string `json:"ids"`
		All   bool     `json:"all"`
		At    float64  `json:"at"`
		Force bool     `json:"force"`
	}
	if err := e.BindBody(&in); err != nil || (!in.All && len(in.IDs) == 0) {
		return httpx.Err(e, 400, "bad_request", "Pick at least one movie.")
	}
	if s.st.Tools.FFmpeg() == "" {
		return httpx.Err(e, 400, "no_ffmpeg", "ffmpeg is not installed on the host, so thumbnails cannot be extracted. Install FFmpeg and restart.")
	}
	if (s.Busy != nil && s.Busy()) || s.Job.Running() {
		return httpx.Err(e, 409, "busy", "Another library job is running. Try again in a moment.")
	}
	go s.RunJob(in.IDs, in.All, util.ClampFloat(in.At, 0, 24*3600), in.Force)
	return e.JSON(202, map[string]any{"ok": true})
}
