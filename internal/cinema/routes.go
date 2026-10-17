package cinema

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
)

// mount lets every feature register its endpoints. Each package owns its own paths; this
// only decides what is public, what needs a guest and what needs the admin.
func (a *app) mount(r *httpx.Router) {
	admin := r.Group("/api/admin")
	admin.Bind(apis.RequireSuperuserAuth())
	rt := httpx.Routes{Public: r, Admin: admin, Guest: a.auth.Guard}

	r.GET("/api/ping", func(e *core.RequestEvent) error { // link quality probe, see the signal monitor
		e.Response.Header().Set("Cache-Control", "no-store")
		return e.NoContent(204)
	})

	for _, f := range []interface{ Mount(httpx.Routes) }{
		a.setup, a.auth, a.catalog, a.subs, a.stream, a.sessions, a.rooms,
		a.scanner, a.thumbs, a.compat, a.settings, a.updates, a.tools, a.admin,
	} {
		f.Mount(rt)
	}

	// Embedded Next.js export (falls back to index.html). It goes last so it never hides an API path.
	r.GET("/{path...}", apis.Static(a.ui, true))
}
