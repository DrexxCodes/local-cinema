// Package httpx holds the small pieces every handler package shares: the JSON error shape,
// the type of a handler that needs a signed in guest, and the route groups packages mount into.
package httpx

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// Router and Group are the PocketBase router types, named once here.
type (
	Router = router.Router[*core.RequestEvent]
	Group  = router.RouterGroup[*core.RequestEvent]
)

// GuestHandler is a handler that runs only for a signed in guest. It receives the room (user)
// and the device session.
type GuestHandler func(e *core.RequestEvent, user, sess *core.Record) error

// Guard turns a GuestHandler into a normal handler that checks the guest first.
type Guard func(GuestHandler) func(*core.RequestEvent) error

// Routes is what a package gets to mount its endpoints.
type Routes struct {
	Public *Router // no sign in needed
	Admin  *Group  // PocketBase superuser only, mounted under /api/admin
	Guest  Guard   // wraps a handler so only signed in guests reach it
}

// Err sends the JSON error the UI understands: {"error": key, "message": text}.
// key is stable and read by code, message is shown to people.
func Err(e *core.RequestEvent, code int, key, msg string) error {
	return e.JSON(code, map[string]any{"error": key, "message": msg})
}

// OK sends {"ok": true}.
func OK(e *core.RequestEvent) error {
	return e.JSON(200, map[string]any{"ok": true})
}
