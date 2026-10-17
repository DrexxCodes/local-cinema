// Package compat makes "universal" MP4 copies of movies that some devices cannot play
// ("Make compatible"), and tells the admin page which movies need it.
package compat

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/progress"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/util"
)

// Service converts movies in the background, one at a time.
type Service struct {
	st *store.Store

	// Job is the progress of the conversion job (with the percent of the current movie).
	Job progress.Tracker

	mu      sync.Mutex
	stop    context.CancelFunc
	lastErr string
}

// New returns the compatibility service.
func New(st *store.Store) *Service { return &Service{st: st} }

// File is where the converted copy of a movie lives.
func (s *Service) File(id string) string { return s.st.Cache.Compat(id) }

// Exists reports whether a converted copy of the movie exists.
func (s *Service) Exists(id string) bool { return util.FileExists(s.File(id)) }

// Status is the compatibility status shown on the admin page (see ffmpeg.Status*).
func (s *Service) Status(rec *core.Record) string {
	return ffmpeg.Classify(rec.GetString("container"), rec.GetString("video_codec"),
		rec.GetString("audio_codec"), s.Exists(rec.Id))
}

// LastError is what went wrong in the last job, "" when it went well.
func (s *Service) LastError() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

func (s *Service) setError(msg string) {
	s.mu.Lock()
	s.lastErr = msg
	s.mu.Unlock()
}

// JoinFailures makes one message out of the per movie failures.
func JoinFailures(failed []string) string { return strings.Join(failed, " | ") }

// convertMovie writes the universal MP4 copy of one movie.
func (s *Service) convertMovie(ctx context.Context, rec *core.Record) error {
	src := rec.GetString("file_path")
	if !s.st.InsideLibrary(src) {
		return errors.New("file is outside the library folders")
	}
	return ffmpeg.Convert(ctx, s.st.Tools.FFmpeg(), ffmpeg.ConvertOptions{
		Src: src, Out: s.File(rec.Id),
		VCodec: rec.GetString("video_codec"), ACodec: rec.GetString("audio_codec"),
		Duration:  rec.GetFloat("duration_seconds"),
		OnPercent: s.Job.SetPercent,
	})
}

// RunJob converts the given movies one after another (or every movie that needs it when all
// is true). It stops at the first cancel and records the failures.
func (s *Service) RunJob(ids []string, all bool) {
	if !s.Job.Begin() {
		return
	}
	defer s.Job.End()
	s.setError("")

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.stop = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		s.stop = nil
		s.mu.Unlock()
	}()

	var recs []*core.Record
	var err error
	if all {
		recs, err = s.st.App.FindAllRecords("movies", dbx.HashExp{"available": true})
	} else {
		recs, err = s.st.App.FindRecordsByIds("movies", ids)
	}
	if err != nil {
		s.setError("Could not read the movie list.")
		return
	}
	todo := recs[:0]
	for _, r := range recs {
		if r.GetBool("available") && ffmpeg.NeedsConversion(s.Status(r), !all) {
			todo = append(todo, r)
		}
	}
	s.Job.Set(len(todo), 0, "")
	var failed []string
	for i, r := range todo {
		s.Job.Set(-1, i, r.GetString("title"))
		s.Job.SetPercent(0)
		started := time.Now()
		if err := s.convertMovie(ctx, r); err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("conversion cancelled")
				break
			}
			slog.Warn("conversion failed", "movie", r.GetString("title"), "err", err)
			failed = append(failed, r.GetString("title")+": "+err.Error())
			continue
		}
		slog.Info("converted", "movie", r.GetString("title"), "took", time.Since(started).Round(time.Second))
		s.st.Audit("converted", "", "", r.Id, nil)
	}
	s.Job.Set(len(todo), len(todo), "")
	if len(failed) > 0 {
		s.setError(JoinFailures(failed))
	}
}

// Cancel stops the running conversion. The file in progress is discarded.
func (s *Service) Cancel() {
	s.mu.Lock()
	stop := s.stop
	s.mu.Unlock()
	if stop != nil {
		stop()
	}
}

// Mount registers the "Make compatible" endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.POST("/compat", s.handleStart)
	rt.Admin.POST("/compat/cancel", s.handleCancel)
	rt.Admin.DELETE("/compat/{id}", s.handleRemove)
}

// handleStart starts a background job.
// Body: {"ids":["..."]} for chosen movies, or {"all":true} for everything that needs it.
func (s *Service) handleStart(e *core.RequestEvent) error {
	var in struct {
		IDs []string `json:"ids"`
		All bool     `json:"all"`
	}
	if err := e.BindBody(&in); err != nil || (!in.All && len(in.IDs) == 0) {
		return httpx.Err(e, 400, "bad_request", "Pick at least one movie.")
	}
	if s.st.Tools.FFmpeg() == "" {
		return httpx.Err(e, 400, "no_ffmpeg", "ffmpeg is not installed on the host, so movies cannot be converted. Install FFmpeg and restart.")
	}
	if s.Job.Running() {
		return httpx.Err(e, 409, "busy", "A conversion is already running.")
	}
	go s.RunJob(in.IDs, in.All)
	return e.JSON(202, map[string]any{"ok": true})
}

func (s *Service) handleCancel(e *core.RequestEvent) error {
	s.Cancel()
	return httpx.OK(e)
}

// handleRemove deletes the converted copy, so guests get the original again.
func (s *Service) handleRemove(e *core.RequestEvent) error {
	if err := os.Remove(s.File(e.Request.PathValue("id"))); err != nil && !os.IsNotExist(err) {
		return httpx.Err(e, 500, "internal", "Could not remove the converted copy.")
	}
	return httpx.OK(e)
}
