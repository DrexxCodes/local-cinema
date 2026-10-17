// Package sessions is everything about who is connected: a room's devices, the connection
// limit, stream links for other players, upgrade requests, and the reaper that ends
// sessions that went quiet.
package sessions

import (
	"log/slog"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/subtitles"
)

// Service handles sessions.
type Service struct {
	st   *store.Store
	subs *subtitles.Service
}

// New returns the sessions service.
func New(st *store.Store, subs *subtitles.Service) *Service {
	return &Service{st: st, subs: subs}
}

// Limit is how many devices a room may use at once: its max_connections, at least 1.
func Limit(maxConnections float64) int {
	if n := int(maxConnections); n >= 1 {
		return n
	}
	return 1
}

// Active lists a room's open sessions, oldest first.
func (s *Service) Active(userID string) []*core.Record {
	recs, _ := s.st.App.FindRecordsByFilter("sessions", "user = {:u} && active = true", "started_at", 100, 0, dbx.Params{"u": userID})
	return recs
}

func (s *Service) end(filter dbx.HashExp) {
	recs, err := s.st.App.FindAllRecords("sessions", filter)
	if err != nil {
		return
	}
	for _, r := range recs {
		r.Set("active", false)
		_ = s.st.App.Save(r)
	}
}

// EndForUser closes every open session of a room.
func (s *Service) EndForUser(userID string) {
	s.end(dbx.HashExp{"user": userID, "active": true})
}

// EndExternal closes the stream link sessions of a room.
func (s *Service) EndExternal(userID string) {
	s.end(dbx.HashExp{"user": userID, "active": true, "kind": "external"})
}

// DTO is how a session is shown to its owner. currentID marks the device asking.
func DTO(r *core.Record, currentID string) map[string]any {
	return map[string]any{
		"id": r.Id, "device_name": r.GetString("device_name"), "ip": r.GetString("ip"),
		"kind": r.GetString("kind"), "current": r.Id == currentID,
		"started_at": r.GetDateTime("started_at").String(),
		"last_seen":  r.GetDateTime("last_seen").String(),
	}
}

// Mount registers every session endpoint.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.GET("/api/sessions", rt.Guest(s.handleMine))
	rt.Public.DELETE("/api/sessions/{id}", rt.Guest(s.handleEndMine))
	rt.Public.POST("/api/movies/{id}/stream-url", rt.Guest(s.handleStreamURL))
	rt.Public.POST("/api/upgrade-requests", rt.Guest(s.handleUpgradeRequest))

	rt.Admin.GET("/sessions", s.handleAdminList)
	rt.Admin.DELETE("/sessions/{id}", s.handleAdminEnd)
	rt.Admin.GET("/upgrade-requests", s.handleAdminUpgrades)
	rt.Admin.PATCH("/upgrade-requests/{id}", s.handleAdminUpgradeResolve)
}

// handleMine lists the devices connected to the guest's room.
func (s *Service) handleMine(e *core.RequestEvent, user, sess *core.Record) error {
	out := []map[string]any{}
	for _, r := range s.Active(user.Id) {
		out = append(out, DTO(r, sess.Id))
	}
	return e.JSON(200, out)
}

// handleEndMine lets a guest free a connection by disconnecting another of their devices.
func (s *Service) handleEndMine(e *core.RequestEvent, user, sess *core.Record) error {
	id := e.Request.PathValue("id")
	if id == sess.Id {
		return httpx.Err(e, 400, "bad_request", "This is the device you are using. Use Log out instead.")
	}
	r, err := s.st.App.FindRecordById("sessions", id)
	if err != nil || r.GetString("user") != user.Id {
		return httpx.Err(e, 404, "not_found", "Device not found.")
	}
	r.Set("active", false)
	_ = s.st.App.Save(r)
	return httpx.OK(e)
}

func (s *Service) handleAdminList(e *core.RequestEvent) error {
	recs, err := s.st.App.FindRecordsByFilter("sessions", "active = true", "-last_seen", 500, 0)
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not list sessions.")
	}
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		out = append(out, map[string]any{
			"id": r.Id, "room_number": r.GetString("room_number"),
			"device_name": r.GetString("device_name"), "ip": r.GetString("ip"),
			"started_at": r.GetDateTime("started_at").String(),
			"last_seen":  r.GetDateTime("last_seen").String(),
			"rtt_ms":     int(r.GetFloat("rtt_ms")),
			"signal":     int(r.GetFloat("signal")),
			"kind":       r.GetString("kind"),
		})
	}
	return e.JSON(200, out)
}

func (s *Service) handleAdminEnd(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("sessions", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Session not found.")
	}
	r.Set("active", false)
	_ = s.st.App.Save(r)
	s.st.Audit("session_killed", r.GetString("user"), r.Id, "", nil)
	return httpx.OK(e)
}

const (
	staleAfter     = 2 * time.Minute  // a browser sends a heartbeat every 15 seconds
	extIdleAfter   = 30 * time.Minute // external players send no heartbeat, so they get longer
	reaperInterval = 60 * time.Second
	dateLayout     = "2006-01-02 15:04:05.000Z" // how PocketBase stores dates
)

// Cutoffs are the "last seen" times before which a browser session and an external player
// session count as gone, in the database's date format.
func Cutoffs(now time.Time) (browser, external string) {
	now = now.UTC()
	return now.Add(-staleAfter).Format(dateLayout), now.Add(-extIdleAfter).Format(dateLayout)
}

// RunReaper marks sessions inactive after a while without a heartbeat. It never returns,
// so start it in its own goroutine.
func (s *Service) RunReaper() {
	t := time.NewTicker(reaperInterval)
	defer t.Stop()
	for range t.C {
		cutoff, extCutoff := Cutoffs(time.Now())
		recs, err := s.st.App.FindAllRecords("sessions",
			dbx.NewExp("active = 1 AND ((COALESCE(kind, '') != 'external' AND last_seen < {:c}) OR (kind = 'external' AND last_seen < {:x}))",
				dbx.Params{"c": cutoff, "x": extCutoff}))
		if err != nil {
			continue
		}
		for _, r := range recs {
			r.Set("active", false)
			_ = s.st.App.Save(r)
		}
		if len(recs) > 0 {
			slog.Info("reaper ended stale sessions", "count", len(recs))
		}
	}
}
