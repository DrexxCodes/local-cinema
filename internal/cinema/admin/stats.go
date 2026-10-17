// Package admin is the dashboard summary the admin page polls.
package admin

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/compat"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/progress"
	"local-cinema/internal/cinema/scanner"
	"local-cinema/internal/cinema/sessions"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/subtitles"
	"local-cinema/internal/cinema/thumbs"
	"local-cinema/internal/cinema/tools"
	"local-cinema/internal/cinema/updates"
)

// Service gathers the numbers for the dashboard.
type Service struct {
	St       *store.Store
	Scanner  *scanner.Service
	Thumbs   *thumbs.Service
	Compat   *compat.Service
	Subs     *subtitles.Service
	Sessions *sessions.Service
	Updates  *updates.Service
	Tools    *tools.Service
}

// Stats is everything the dashboard shows. Build turns it into the JSON the UI reads.
type Stats struct {
	ActiveSessions  int64
	Movies          int64
	Uptime          time.Duration
	Scan            progress.Snapshot
	Thumbs          progress.Snapshot
	Convert         progress.Snapshot
	ConvertError    string
	ScanIssues      []string
	Subs            progress.Snapshot
	SubsMessage     string
	PendingUpgrades int64
	UpdateAvailable string
	FFmpegInstall   progress.Snapshot
	HasFFmpeg       bool
}

// Build is the JSON shape of GET /api/admin/stats. The key names are read by the admin UI.
func (s Stats) Build() map[string]any {
	issues := s.ScanIssues
	if issues == nil {
		issues = []string{}
	}
	return map[string]any{
		"active_sessions":  s.ActiveSessions,
		"movies":           s.Movies,
		"uptime_seconds":   int(s.Uptime.Seconds()),
		"scan":             s.Scan,
		"thumbs":           s.Thumbs,
		"convert":          s.Convert,
		"convert_error":    s.ConvertError,
		"scan_issues":      issues,
		"subs":             s.Subs,
		"subs_msg":         s.SubsMessage,
		"pending_upgrades": s.PendingUpgrades,
		"update_available": s.UpdateAvailable,
		"ffmpeg_install":   s.FFmpegInstall,
		"ffmpeg":           s.HasFFmpeg,
	}
}

// Mount registers the endpoint.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.GET("/stats", s.handleStats)
}

func (s *Service) handleStats(e *core.RequestEvent) error {
	active, _ := s.St.App.CountRecords("sessions", dbx.HashExp{"active": true})
	movies, _ := s.St.App.CountRecords("movies", dbx.HashExp{"available": true})
	return e.JSON(200, Stats{
		ActiveSessions:  active,
		Movies:          movies,
		Uptime:          time.Since(s.St.Started),
		Scan:            s.Scanner.State.Snapshot(),
		Thumbs:          s.Thumbs.Job.Snapshot(),
		Convert:         s.Compat.Job.Snapshot(),
		ConvertError:    s.Compat.LastError(),
		ScanIssues:      s.Scanner.Issues(),
		Subs:            s.Subs.Extract.Snapshot(),
		SubsMessage:     s.Subs.Note(),
		PendingUpgrades: s.Sessions.PendingUpgrades(),
		UpdateAvailable: s.Updates.Available(),
		FFmpegInstall:   s.Tools.Job.Snapshot(),
		HasFFmpeg:       s.St.Tools.FFmpeg() != "",
	}.Build())
}
