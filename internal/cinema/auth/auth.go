// Package auth signs guests in and out and protects the guest endpoints. A guest is a room
// (a PocketBase "users" record) on one device (a "sessions" record).
package auth

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/ratelimit"
	"local-cinema/internal/cinema/sessions"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/util"
)

// touchEvery limits how often a session's last_seen is written, so range requests stay cheap.
const touchEvery = 10 * time.Second

// Service handles sign in, sign out, heartbeats and the guest guard.
type Service struct {
	st      *store.Store
	limiter *ratelimit.Limiter
}

// New returns the auth service. Sign in attempts are limited per IP address.
func New(st *store.Store) *Service {
	return &Service{st: st, limiter: ratelimit.NewLogin()}
}

// Credentials reads the token and the session id from a request. They may arrive in the
// Authorization and X-Session-Id headers, or as ?token= and ?sid= (needed for <video> and
// <img> tags, which cannot send headers). Headers win.
func Credentials(authHeader, sessionHeader string, q url.Values) (token, sid string) {
	token = strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		token = q.Get("token")
	}
	sid = sessionHeader
	if sid == "" {
		sid = q.Get("sid")
	}
	return token, sid
}

// NeedsTouch says a session's last_seen is old enough to be written again.
func NeedsTouch(lastSeen, now time.Time) bool {
	return lastSeen.Before(now.Add(-touchEvery))
}

// Guard validates the JWT and the server side session, then runs next with the room and the session.
func (s *Service) Guard(next httpx.GuestHandler) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		tok, sid := Credentials(e.Request.Header.Get("Authorization"), e.Request.Header.Get("X-Session-Id"), e.Request.URL.Query())
		if tok == "" || sid == "" {
			return httpx.Err(e, 401, "unauthorized", "Please log in.")
		}
		user, err := s.st.App.FindAuthRecordByToken(tok, core.TokenTypeAuth)
		if err != nil || user.Collection().Name != "users" || !user.GetBool("active") {
			return httpx.Err(e, 401, "unauthorized", "Please log in.")
		}
		sess, err := s.st.App.FindRecordById("sessions", sid)
		if err != nil || !sess.GetBool("active") || sess.GetString("user") != user.Id {
			return httpx.Err(e, 401, "session_ended", "Your session ended. Please log in again.")
		}
		if NeedsTouch(sess.GetDateTime("last_seen").Time(), time.Now()) {
			s.st.Touch(sess)
		}
		return next(e, user, sess)
	}
}

// Mount registers the auth endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.POST("/api/auth/login", s.handleLogin)
	rt.Public.POST("/api/auth/logout", rt.Guest(s.handleLogout))
	rt.Public.POST("/api/auth/heartbeat", rt.Guest(s.handleHeartbeat))
	rt.Public.GET("/api/auth/me", rt.Guest(s.handleMe))
}

type loginBody struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

// NormalizeEmail turns what a guest typed into the account email. A room number such as
// "Profile204" is enough: it becomes profile204@cinema.local.
func NormalizeEmail(in string) string {
	email := strings.ToLower(strings.TrimSpace(in))
	if email != "" && !strings.Contains(email, "@") {
		email += "@cinema.local"
	}
	return email
}

func userDTO(user *core.Record) map[string]any {
	return map[string]any{
		"id":           user.Id,
		"room_number":  user.GetString("room_number"),
		"display_name": user.GetString("display_name"),
	}
}

func (s *Service) handleLogin(e *core.RequestEvent) error {
	var in loginBody
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	if !s.limiter.Allow(e.RealIP()) {
		return httpx.Err(e, 429, "rate_limited", "Too many attempts. Wait a minute and try again.")
	}
	cfg := s.st.Settings()
	if !cfg.GetBool("setup_complete") {
		return httpx.Err(e, 503, "setup_required", "The cinema has not been set up yet.")
	}

	email := NormalizeEmail(in.Email)
	user, err := s.st.App.FindAuthRecordByEmail("users", email)
	if err != nil || !user.ValidatePassword(in.Password) {
		s.st.Audit("login_failed", "", "", "", map[string]any{"email": email, "ip": e.RealIP()})
		return httpx.Err(e, 401, "invalid_credentials", "Wrong profile or password.")
	}
	if !user.GetBool("active") {
		return httpx.Err(e, 403, "account_disabled", "This profile is not currently active. Contact the front desk.")
	}

	// Same device + same user: reuse the active session.
	if in.DeviceID != "" {
		existing, err := s.st.App.FindFirstRecordByFilter("sessions",
			"user = {:u} && device_id = {:d} && active = true",
			dbx.Params{"u": user.Id, "d": in.DeviceID})
		if err == nil && existing != nil {
			s.st.Touch(existing)
			return s.loginOK(e, user, existing)
		}
	}

	max := sessions.Limit(user.GetFloat("max_connections"))
	cur, _ := s.st.App.CountRecords("sessions", dbx.HashExp{"user": user.Id, "active": true})
	if int(cur) >= max {
		s.st.Audit("cap_exceeded", user.Id, "", "", map[string]any{"max": max, "current": cur})
		return e.JSON(http.StatusTooManyRequests, map[string]any{
			"error":   "max_connections_reached",
			"message": "Room " + user.GetString("room_number") + " is already streaming on " + strconv.Itoa(int(cur)) + " devices. Ask the other viewer to stop or wait a few minutes.",
			"room":    user.GetString("room_number"),
			"max":     max,
			"current": cur,
		})
	}

	globalMax := int(cfg.GetFloat("global_max_streams"))
	if globalMax > 0 {
		total, _ := s.st.App.CountRecords("sessions", dbx.HashExp{"active": true})
		if int(total) >= globalMax {
			s.st.Audit("global_cap_exceeded", user.Id, "", "", nil)
			return httpx.Err(e, 503, "server_at_capacity", "The cinema is full right now. Try again in a few minutes.")
		}
	}

	c, err := s.st.App.FindCollectionByNameOrId("sessions")
	if err != nil {
		return httpx.Err(e, 500, "internal", "Server error.")
	}
	sess := core.NewRecord(c)
	sess.Set("user", user.Id)
	sess.Set("device_id", in.DeviceID)
	sess.Set("device_name", in.DeviceName)
	sess.Set("ip", e.RealIP())
	sess.Set("user_agent", e.Request.UserAgent())
	sess.Set("room_number", user.GetString("room_number"))
	sess.Set("started_at", types.NowDateTime())
	sess.Set("last_seen", types.NowDateTime())
	sess.Set("active", true)
	if err := s.st.App.Save(sess); err != nil {
		return httpx.Err(e, 500, "internal", "Could not start a session.")
	}
	s.st.Audit("login", user.Id, sess.Id, "", map[string]any{"ip": e.RealIP()})
	return s.loginOK(e, user, sess)
}

func (s *Service) loginOK(e *core.RequestEvent, user, sess *core.Record) error {
	tok, err := user.NewAuthToken()
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not issue a token.")
	}
	return e.JSON(200, map[string]any{"token": tok, "session_id": sess.Id, "user": userDTO(user)})
}

func (s *Service) handleLogout(e *core.RequestEvent, user, sess *core.Record) error {
	sess.Set("active", false)
	_ = s.st.App.Save(sess)
	s.st.Audit("logout", user.Id, sess.Id, "", nil)
	return httpx.OK(e)
}

// HeartbeatValues limits the link quality a device reports: round trip up to a minute, signal 0 to 4.
func HeartbeatValues(rttMs float64, signal int) (float64, int) {
	return util.ClampFloat(rttMs, 0, 60000), util.ClampInt(signal, 0, 4)
}

// handleHeartbeat keeps the session alive and stores the client measured link quality (round
// trip in ms and a 0 to 4 signal level) for the admin monitor.
func (s *Service) handleHeartbeat(e *core.RequestEvent, _, sess *core.Record) error {
	if e.Request.ContentLength != 0 {
		var in struct {
			RTT    float64 `json:"rtt_ms"`
			Signal int     `json:"signal"`
		}
		if err := e.BindBody(&in); err == nil {
			rtt, sig := HeartbeatValues(in.RTT, in.Signal)
			sess.Set("rtt_ms", rtt)
			sess.Set("signal", sig)
		}
	}
	s.st.Touch(sess)
	return httpx.OK(e)
}

func (s *Service) handleMe(e *core.RequestEvent, user, sess *core.Record) error {
	return e.JSON(200, map[string]any{"user": userDTO(user), "session_id": sess.Id})
}
