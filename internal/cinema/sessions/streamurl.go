package sessions

import (
	"context"
	"net/url"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/netinfo"
	"local-cinema/internal/cinema/store"
)

// KeepAlive refreshes an external session while its request is open, because players do not
// send heartbeats and one request can last the whole movie. It returns when ctx ends.
func (s *Service) KeepAlive(ctx context.Context, sess *core.Record) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.st.Touch(sess)
		}
	}
}

// DeviceKey names the stream link session of a device, so asking again gives the same link.
// A browser without a device id falls back to its session id.
func DeviceKey(deviceID, sessionID string) string {
	if deviceID == "" {
		return "ext:" + sessionID
	}
	return "ext:" + deviceID
}

// AuthQuery is the query string that signs a link: the token and the session id.
func AuthQuery(token, sessionID string) string {
	return "token=" + url.QueryEscape(token) + "&sid=" + url.QueryEscape(sessionID)
}

// StreamLink is the address of a movie's file for another player.
func StreamLink(base, movieID, query string) string {
	return base + "/stream/" + movieID + "?" + query
}

// SubtitleLink is the address of one subtitle track for another player.
func SubtitleLink(base, movieID, trackKey, query string) string {
	return base + "/api/movies/" + movieID + "/subtitles/" + url.PathEscape(trackKey) + "?" + query
}

// limitReached is the answer when a room has used all its connections. It lists the devices
// holding them and the latest upgrade request, so the guest can act.
func (s *Service) limitReached(e *core.RequestEvent, user, sess *core.Record, max int) error {
	devices := []map[string]any{}
	for _, r := range s.Active(user.Id) {
		devices = append(devices, DTO(r, sess.Id))
	}
	var req any
	if last, err := s.st.App.FindRecordsByFilter("upgrade_requests", "user = {:u}", "-created", 1, 0, dbx.Params{"u": user.Id}); err == nil && len(last) > 0 {
		req = map[string]any{"status": last[0].GetString("status"), "slots": int(last[0].GetFloat("slots")), "granted": int(last[0].GetFloat("granted"))}
	}
	s.st.Audit("cap_exceeded", user.Id, sess.Id, "", map[string]any{"max": max, "via": "stream_url"})
	return e.JSON(429, map[string]any{
		"error":   "max_connections_reached",
		"message": "You have reached the maximum connection for this profile.",
		"max":     max, "devices": devices, "request": req,
	})
}

// handleStreamURL gives the guest a link for another player (VLC, Infuse and so on).
// The link carries its own session, which counts against the room's connection limit. A
// device keeps one such session, so asking again returns the same link.
func (s *Service) handleStreamURL(e *core.RequestEvent, user, sess *core.Record) error {
	m, err := s.st.App.FindRecordById("movies", e.Request.PathValue("id"))
	if err != nil || !store.Streamable(m) {
		return httpx.Err(e, 404, "not_found", "This movie is no longer available.")
	}
	if !user.GetBool("allow_external") {
		return httpx.Err(e, 403, "external_not_allowed", "External streaming is not turned on for this profile. Ask the front desk to enable it.")
	}

	deviceKey := DeviceKey(sess.GetString("device_id"), sess.Id)
	ext, err := s.st.App.FindFirstRecordByFilter("sessions", "user = {:u} && device_id = {:d} && active = true",
		dbx.Params{"u": user.Id, "d": deviceKey})
	if err != nil || ext == nil {
		max := Limit(user.GetFloat("max_connections"))
		cur, _ := s.st.App.CountRecords("sessions", dbx.HashExp{"user": user.Id, "active": true})
		if int(cur) >= max {
			return s.limitReached(e, user, sess, max)
		}
		if gm := int(s.st.Settings().GetFloat("global_max_streams")); gm > 0 {
			if total, _ := s.st.App.CountRecords("sessions", dbx.HashExp{"active": true}); int(total) >= gm {
				return httpx.Err(e, 503, "server_at_capacity", "The cinema is full right now. Try again in a few minutes.")
			}
		}
		c, err := s.st.App.FindCollectionByNameOrId("sessions")
		if err != nil {
			return httpx.Err(e, 500, "internal", "Server error.")
		}
		ext = core.NewRecord(c)
		ext.Set("user", user.Id)
		ext.Set("device_id", deviceKey)
		ext.Set("device_name", "External player ("+sess.GetString("device_name")+")")
		ext.Set("ip", e.RealIP())
		ext.Set("room_number", user.GetString("room_number"))
		ext.Set("kind", "external")
		ext.Set("started_at", types.NowDateTime())
		ext.Set("last_seen", types.NowDateTime())
		ext.Set("active", true)
		if err := s.st.App.Save(ext); err != nil {
			return httpx.Err(e, 500, "internal", "Could not create the link.")
		}
		s.st.Audit("stream_url", user.Id, ext.Id, m.Id, nil)
	} else {
		s.st.Touch(ext)
	}

	tok, err := user.NewAuthToken()
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not issue a link.")
	}
	q := AuthQuery(tok, ext.Id)
	base := netinfo.ExternalBase(e.Request.Host, e.Request.TLS != nil, netinfo.LAN())
	subs := []map[string]string{}
	for _, t := range s.subs.Tracks(m) {
		subs = append(subs, map[string]string{"label": t.Label, "url": SubtitleLink(base, m.Id, t.Key, q)})
	}
	return e.JSON(200, map[string]any{
		"url": StreamLink(base, m.Id, q), "subtitles": subs, "title": m.GetString("title"),
	})
}
