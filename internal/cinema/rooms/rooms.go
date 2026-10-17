// Package rooms is the admin's management of room profiles (the guests' accounts): listing,
// editing, deleting, and creating many at once.
package rooms

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/sessions"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/util"
)

const (
	pageDefault = 20
	pageMax     = 100
	maxBulk     = 1000 // most rooms one bulk request can create
)

// Service manages rooms.
type Service struct {
	st       *store.Store
	sessions *sessions.Service
}

// New returns the rooms service.
func New(st *store.Store, sess *sessions.Service) *Service {
	return &Service{st: st, sessions: sess}
}

// Mount registers the admin endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.POST("/rooms/bulk", s.handleBulk)
	rt.Admin.GET("/users", s.handleList)
	rt.Admin.PATCH("/users/{id}", s.handleUpdate)
	rt.Admin.DELETE("/users/{id}", s.handleDelete)
}

// Username is the part of an email before the @, which is what guests type to sign in.
func Username(email string) string { return strings.SplitN(email, "@", 2)[0] }

// RoomEmail is the sign in email of a room: prefix + number + @cinema.local.
func RoomEmail(prefix string, room int) string {
	return prefix + strconv.Itoa(room) + "@cinema.local"
}

// ValidRange checks a bulk request: from must not exceed to, and at most 1000 rooms.
func ValidRange(from, to int) bool { return to >= from && to-from <= maxBulk }

// BulkDefaults fills in what a bulk request left out: prefix "profile" and 2 devices per profile.
func BulkDefaults(prefix string, maxConnections int) (string, int) {
	if prefix == "" {
		prefix = "profile"
	}
	if maxConnections < 1 {
		maxConnections = 2
	}
	return prefix, maxConnections
}

// NewPassword makes a random 8 character password (4 random bytes in hex).
func NewPassword() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ClampDevices limits a room's device count to 1 to 50.
func ClampDevices(n int) int { return util.ClampInt(n, 1, 50) }

func userDTO(r *core.Record, online int) map[string]any {
	email := r.Email()
	return map[string]any{
		"id":              r.Id,
		"room_number":     r.GetString("room_number"),
		"username":        Username(email),
		"email":           email,
		"display_name":    r.GetString("display_name"),
		"max_connections": int(r.GetFloat("max_connections")),
		"active":          r.GetBool("active"),
		"allow_external":  r.GetBool("allow_external"),
		"notes":           r.GetString("notes"),
		"online":          online,
		"created":         r.GetDateTime("created").String(),
	}
}

// handleBulk creates room profiles.
// Body: {"prefix":"profile","from":101,"to":120,"max_connections":2,"password":""}
// An empty password generates one per room. Passwords are returned once.
func (s *Service) handleBulk(e *core.RequestEvent) error {
	var in struct {
		Prefix         string `json:"prefix"`
		From           int    `json:"from"`
		To             int    `json:"to"`
		MaxConnections int    `json:"max_connections"`
		Password       string `json:"password"`
		AllowExternal  bool   `json:"allow_external"`
	}
	if err := e.BindBody(&in); err != nil || !ValidRange(in.From, in.To) {
		return httpx.Err(e, 400, "bad_request", "Provide a valid from/to range (max 1000 profiles).")
	}
	prefix, maxConn := BulkDefaults(in.Prefix, in.MaxConnections)
	c, err := s.st.App.FindCollectionByNameOrId("users")
	if err != nil {
		return httpx.Err(e, 500, "internal", "Server error.")
	}
	type created struct {
		Room     string `json:"room"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var made []created
	var skipped []string
	for n := in.From; n <= in.To; n++ {
		room := strconv.Itoa(n)
		email := RoomEmail(prefix, n)
		if _, err := s.st.App.FindAuthRecordByEmail("users", email); err == nil {
			skipped = append(skipped, room)
			continue
		}
		pw := in.Password
		if pw == "" {
			pw = NewPassword()
		}
		u := core.NewRecord(c)
		u.SetEmail(email)
		u.SetPassword(pw)
		u.SetVerified(true)
		u.Set("room_number", room)
		u.Set("display_name", "Profile "+room)
		u.Set("max_connections", maxConn)
		u.Set("active", true)
		u.Set("allow_external", in.AllowExternal)
		if err := s.st.App.Save(u); err != nil {
			skipped = append(skipped, room)
			continue
		}
		made = append(made, created{room, email, pw})
	}
	return e.JSON(200, map[string]any{"created": made, "skipped": skipped})
}

// handleList lists room profiles a page at a time.
// Query: limit (default 20), offset, q (matches room, name or email).
func (s *Service) handleList(e *core.RequestEvent) error {
	qs := e.Request.URL.Query()
	limit := util.QueryInt(qs.Get("limit"), pageDefault, 1, pageMax)
	offset := util.QueryInt(qs.Get("offset"), 0, 0, 1<<30)
	q := strings.TrimSpace(qs.Get("q"))

	var filters []dbx.Expression
	if q != "" {
		filters = append(filters, dbx.Or(
			dbx.Like("room_number", q), dbx.Like("display_name", q), dbx.Like("email", q)))
	}
	total, err := s.st.App.CountRecords("users", filters...)
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not count users.")
	}

	query := s.st.App.RecordQuery("users").
		OrderBy("CAST(room_number AS INTEGER) ASC", "room_number ASC", "id ASC").
		Limit(int64(limit)).Offset(int64(offset))
	for _, f := range filters {
		query.AndWhere(f)
	}
	var recs []*core.Record
	if err := query.All(&recs); err != nil {
		slog.Warn("admin users: list", "err", err)
		return httpx.Err(e, 500, "internal", "Could not list users: "+err.Error())
	}

	// Open sessions per room, so the list can show who is watching now.
	online := map[string]int{}
	if live, err := s.st.App.FindAllRecords("sessions", dbx.HashExp{"active": true}); err == nil {
		for _, sr := range live {
			online[sr.GetString("user")]++
		}
	}
	items := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		items = append(items, userDTO(r, online[r.Id]))
	}
	return e.JSON(200, map[string]any{
		"items": items, "total": total, "offset": offset,
		"has_more": int64(offset+len(items)) < total,
	})
}

func (s *Service) handleUpdate(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("users", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Profile not found.")
	}
	var in struct {
		DisplayName    *string `json:"display_name"`
		MaxConnections *int    `json:"max_connections"`
		Active         *bool   `json:"active"`
		Notes          *string `json:"notes"`
		Password       *string `json:"password"`
		AllowExternal  *bool   `json:"allow_external"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	if in.DisplayName != nil {
		r.Set("display_name", strings.TrimSpace(*in.DisplayName))
	}
	if in.MaxConnections != nil {
		r.Set("max_connections", ClampDevices(*in.MaxConnections))
	}
	if in.Notes != nil {
		r.Set("notes", *in.Notes)
	}
	if in.AllowExternal != nil {
		r.Set("allow_external", *in.AllowExternal)
		if !*in.AllowExternal {
			s.sessions.EndExternal(r.Id)
		}
	}
	if in.Active != nil {
		r.Set("active", *in.Active)
	}
	if in.Password != nil && *in.Password != "" {
		r.SetPassword(*in.Password)
	}
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 400, "invalid", "Could not save: "+err.Error())
	}
	if (in.Active != nil && !*in.Active) || (in.Password != nil && *in.Password != "") {
		s.sessions.EndForUser(r.Id)
	}
	s.st.Audit("user_updated", r.Id, "", "", nil)
	return e.JSON(200, userDTO(r, 0))
}

func (s *Service) handleDelete(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("users", e.Request.PathValue("id"))
	if err != nil {
		return httpx.Err(e, 404, "not_found", "Profile not found.")
	}
	if err := s.st.App.Delete(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not delete this profile.")
	}
	s.st.Audit("user_deleted", "", "", "", map[string]any{"room": r.GetString("room_number")})
	return httpx.OK(e)
}
