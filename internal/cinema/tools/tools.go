// Package tools is the admin Settings page for FFmpeg: whether it is installed, and the
// button that downloads it.
package tools

import (
	"context"
	"runtime"
	"sync"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/progress"
)

// Service reports on FFmpeg and installs it.
type Service struct {
	Tools     *ffmpeg.Tools
	Installer *ffmpeg.Installer
	Job       progress.Tracker

	mu  sync.Mutex
	msg string
}

// New returns the service. The installer's step and percent hooks feed the job tracker.
func New(t *ffmpeg.Tools, in *ffmpeg.Installer) *Service {
	s := &Service{Tools: t, Installer: in}
	in.OnStep = func(m string) { s.Job.Set(1, 0, m) }
	in.OnPercent = s.Job.SetPercent
	return s
}

func (s *Service) setMsg(m string) { s.mu.Lock(); s.msg = m; s.mu.Unlock() }

// Message is the outcome of the last install, for the page to show.
func (s *Service) Message() string { s.mu.Lock(); defer s.mu.Unlock(); return s.msg }

// Install downloads FFmpeg and makes the program find it. It does nothing when an install
// is already running.
func (s *Service) Install(ctx context.Context) {
	if !s.Job.Begin() {
		return
	}
	defer s.Job.End()
	s.setMsg("")
	if err := s.Installer.Install(ctx); err != nil {
		s.setMsg("Could not install FFmpeg: " + err.Error())
		return
	}
	s.Tools.Reset()
	s.setMsg("FFmpeg is installed.")
}

// Status is the JSON for GET /tools.
func (s *Service) Status() map[string]any {
	ff, fp := s.Tools.FFmpeg(), s.Tools.FFprobe()
	return map[string]any{
		"ffmpeg": ff != "", "ffprobe": fp != "",
		"ffmpeg_version": s.Tools.Version(ff), "ffmpeg_path": ff,
		"can_install": s.Installer.Supported(), "install": s.Job.Snapshot(), "message": s.Message(),
		"platform": runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// Mount registers the admin endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.GET("/tools", func(e *core.RequestEvent) error { return e.JSON(200, s.Status()) })
	rt.Admin.POST("/tools/install", s.handleInstall)
}

func (s *Service) handleInstall(e *core.RequestEvent) error {
	if !s.Installer.Supported() {
		return httpx.Err(e, 400, "unsupported", "This build can't download FFmpeg for this computer. Install it yourself (for example brew install ffmpeg) and restart.")
	}
	if s.Job.Running() {
		return httpx.Err(e, 409, "busy", "FFmpeg is already being installed.")
	}
	go s.Install(context.Background())
	return e.JSON(202, map[string]any{"ok": true})
}
