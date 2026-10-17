package sessions

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/util"
)

const (
	minSlots = 1
	maxSlots = 20 // most extra connections one request can ask for or be granted
	maxLimit = 50 // most connections a room can have
)

// ValidSlots reports whether n is an allowed number of extra connections.
func ValidSlots(n int) bool { return n >= minSlots && n <= maxSlots }

// RaisedLimit is a room's limit after granting extra connections.
func RaisedLimit(current, extra int) int { return util.ClampInt(current+extra, 1, maxLimit) }

// PendingUpgrades counts the requests waiting for the admin.
func (s *Service) PendingUpgrades() int64 {
	n, _ := s.st.App.CountRecords("upgrade_requests", dbx.HashExp{"status": "pending"})
	return n
}

// handleUpgradeRequest records how many extra connections the guest needs.
// A room has at most one open request, asking again updates it.
func (s *Service) handleUpgradeRequest(e *core.RequestEvent, user, _ *core.Record) error {
	var in struct {
		Slots int `json:"slots"`
	}
	if err := e.BindBody(&in); err != nil || !ValidSlots(in.Slots) {
		return httpx.Err(e, 400, "invalid", "Enter how many extra connections you need (1 to 20).")
	}
	r, err := s.st.App.FindFirstRecordByFilter("upgrade_requests", "user = {:u} && status = 'pending'", dbx.Params{"u": user.Id})
	if err != nil || r == nil {
		c, err := s.st.App.FindCollectionByNameOrId("upgrade_requests")
		if err != nil {
			return httpx.Err(e, 500, "internal", "Server error.")
		}
		r = core.NewRecord(c)
		r.Set("user", user.Id)
		r.Set("status", "pending")
	}
	r.Set("slots", in.Slots)
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not send the request.")
	}
	s.st.Audit("upgrade_requested", user.Id, "", "", map[string]any{"slots": in.Slots})
	return e.JSON(200, map[string]any{"status": "pending", "slots": in.Slots})
}

// handleAdminUpgrades lists requests. Query: status=pending (default) or resolved.
func (s *Service) handleAdminUpgrades(e *core.RequestEvent) error {
	filter := "status = 'pending'"
	if e.Request.URL.Query().Get("status") == "resolved" {
		filter = "status != 'pending'"
	}
	recs, err := s.st.App.FindRecordsByFilter("upgrade_requests", filter, "-created", 100, 0)
	if err != nil {
		return httpx.Err(e, 500, "internal", "Could not list requests: "+err.Error())
	}
	out := []map[string]any{}
	for _, r := range recs {
		u, err := s.st.App.FindRecordById("users", r.GetString("user"))
		if err != nil {
			continue
		}
		online, _ := s.st.App.CountRecords("sessions", dbx.HashExp{"user": u.Id, "active": true})
		out = append(out, map[string]any{
			"id": r.Id, "room_number": u.GetString("room_number"), "display_name": u.GetString("display_name"),
			"max_connections": int(u.GetFloat("max_connections")), "online": online,
			"slots": int(r.GetFloat("slots")), "granted": int(r.GetFloat("granted")), "status": r.GetString("status"),
			"created": r.GetDateTime("created").String(),
		})
	}
	return e.JSON(200, out)
}

// handleAdminUpgradeResolve approves (adding the given number of connections, which the admin
// may change from what was asked) or declines a request.
// Body: {"action":"approve"|"decline","slots":2}
func (s *Service) handleAdminUpgradeResolve(e *core.RequestEvent) error {
	r, err := s.st.App.FindRecordById("upgrade_requests", e.Request.PathValue("id"))
	if err != nil || r.GetString("status") != "pending" {
		return httpx.Err(e, 404, "not_found", "That request is not open any more.")
	}
	var in struct {
		Action string `json:"action"`
		Slots  int    `json:"slots"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	switch in.Action {
	case "decline":
		r.Set("status", "declined")
	case "approve":
		if !ValidSlots(in.Slots) {
			return httpx.Err(e, 400, "invalid", "Approve between 1 and 20 extra connections.")
		}
		u, err := s.st.App.FindRecordById("users", r.GetString("user"))
		if err != nil {
			return httpx.Err(e, 404, "not_found", "That profile no longer exists.")
		}
		u.Set("max_connections", RaisedLimit(int(u.GetFloat("max_connections")), in.Slots))
		if err := s.st.App.Save(u); err != nil {
			return httpx.Err(e, 500, "internal", "Could not update the profile: "+err.Error())
		}
		r.Set("status", "approved")
		r.Set("granted", in.Slots)
	default:
		return httpx.Err(e, 400, "bad_request", "Action must be approve or decline.")
	}
	if err := s.st.App.Save(r); err != nil {
		return httpx.Err(e, 500, "internal", "Could not save: "+err.Error())
	}
	s.st.Audit("upgrade_"+in.Action, r.GetString("user"), "", "", map[string]any{"slots": in.Slots})
	return httpx.OK(e)
}
