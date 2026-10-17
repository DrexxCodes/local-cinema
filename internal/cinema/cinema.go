// Package cinema wires Local Cinema into a PocketBase app. The work itself lives in the
// sub packages, one folder per feature (see README.md).
package cinema

import (
	"io/fs"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/admin"
	"local-cinema/internal/cinema/auth"
	"local-cinema/internal/cinema/catalog"
	"local-cinema/internal/cinema/compat"
	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/mdns"
	"local-cinema/internal/cinema/netinfo"
	"local-cinema/internal/cinema/rooms"
	"local-cinema/internal/cinema/scanner"
	"local-cinema/internal/cinema/sessions"
	"local-cinema/internal/cinema/settings"
	"local-cinema/internal/cinema/setup"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/stream"
	"local-cinema/internal/cinema/subtitles"
	"local-cinema/internal/cinema/thumbs"
	"local-cinema/internal/cinema/tools"
	"local-cinema/internal/cinema/updates"
	"local-cinema/internal/version"
)

// ListenPort returns the HTTP port (CINEMA_PORT overrides the 8765 default).
func ListenPort() string { return netinfo.ListenPort() }

// app holds every feature service.
type app struct {
	ui fs.FS

	st       *store.Store
	auth     *auth.Service
	subs     *subtitles.Service
	thumbs   *thumbs.Service
	compat   *compat.Service
	scanner  *scanner.Service
	sessions *sessions.Service
	rooms    *rooms.Service
	settings *settings.Service
	setup    *setup.Service
	catalog  *catalog.Service
	stream   *stream.Service
	updates  *updates.Service
	tools    *tools.Service
	mdns     *mdns.Service
	admin    *admin.Service
}

// newApp creates the services and connects the ones that need each other.
func newApp(pb core.App, ui fs.FS) *app {
	st := store.New(pb)
	a := &app{ui: ui, st: st}

	a.auth = auth.New(st)
	a.subs = subtitles.New(st)
	a.thumbs = thumbs.New(st)
	a.compat = compat.New(st)
	meta := scanner.NewMetadata(st)
	a.scanner = scanner.New(st, a.thumbs, meta)
	a.sessions = sessions.New(st, a.subs)
	a.rooms = rooms.New(st, a.sessions)
	a.settings = settings.New(st)
	a.setup = setup.New(st, a.scanner)
	a.catalog = catalog.New(st, a.compat, meta)
	a.stream = stream.New(st, a.compat, a.sessions)
	a.updates = updates.New(version.Version, version.Repo)
	a.tools = tools.New(st.Tools, &ffmpeg.Installer{
		Repo: version.Repo, Bundle: version.FFmpegBundle, UserAgent: "local-cinema/" + version.Version,
	})
	a.mdns = mdns.New(st)
	a.admin = &admin.Service{
		St: st, Scanner: a.scanner, Thumbs: a.thumbs, Compat: a.compat, Subs: a.subs,
		Sessions: a.sessions, Updates: a.updates, Tools: a.tools,
	}

	// The packages do not import each other where that would loop, so the links are set here.
	a.thumbs.Busy = a.scanner.State.Running
	a.settings.OnIdentityChanged = func() { go a.mdns.Restart() }
	return a
}

// Register wires Local Cinema into a PocketBase app.
func Register(pb *pocketbase.PocketBase, ui fs.FS) {
	a := newApp(pb, ui)

	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := a.st.EnsureCollections(); err != nil {
			return err
		}
		a.st.Init()
		a.mount(e.Router)
		a.startBackground()
		return e.Next()
	})

	// Re-sync watches and rescan when the library folders change after setup.
	pb.OnRecordAfterUpdateSuccess("settings").BindFunc(func(e *core.RecordEvent) error {
		a.scanner.OnSettingsUpdated(e.Record)
		return e.Next()
	})

	pb.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		a.mdns.Stop()
		a.scanner.Close()
		return e.Next()
	})
}

// startBackground starts the jobs that run for as long as the server does.
func (a *app) startBackground() {
	go a.sessions.RunReaper()
	a.mdns.Start()
	a.scanner.StartWatching()
	go a.updates.Loop() // checks GitHub for a newer release now and then daily
	// Give movies that already exist (and have no art yet) a thumbnail in the background.
	if a.st.Settings().GetBool("setup_complete") && a.st.Tools.FFmpeg() != "" {
		go a.thumbs.RunJob(nil, true, 0, false)
	}
}
