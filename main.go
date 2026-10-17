package main

import (
	"embed"
	"io/fs"
	"log"
	"os"

	"github.com/pocketbase/pocketbase"

	"local-cinema/internal/cinema"
	"local-cinema/internal/desktop"
	"local-cinema/internal/version"
)

// The Next.js static export is copied into ui/out by `make ui`.
//
//go:embed all:ui/out
var uiFS embed.FS

func main() {
	// --background starts the control panel minimized (Windows uses it when it starts the cinema at log in).
	background := takeFlag("--background")

	port := cinema.ListenPort()
	plain := len(os.Args) == 1 // double-click / no arguments
	dataDir, installed := defaultDataDir()

	// On Windows a double-click opens the control panel. The panel runs the server as a hidden
	// copy of this program (`serve`), so Stop and Start work and its log is visible.
	if plain && desktop.Supported() {
		if !desktop.ClaimInstance(background) {
			return // another panel is open, it has been brought to the front
		}
		desktop.Run(desktop.Options{
			Port:       port,
			Background: background,
			Version:    version.Version,
			LogDir:     logDirFor(dataDir, installed),
		})
		return
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: dataDir})
	app.RootCmd.Version = version.Version
	// The installed app has no console window. When the panel started this copy it reads
	// the output instead, and writes the log file itself.
	if installed && os.Getenv("CINEMA_CHILD") == "" {
		setupFileLogging(dataDir)
	}

	sub, err := fs.Sub(uiFS, "ui/out")
	if err != nil {
		log.Fatal(err)
	}
	cinema.Register(app, sub)

	// Double-click / no-args behaviour: serve on all interfaces.
	if plain {
		os.Args = append(os.Args, "serve", "--http=0.0.0.0:"+port)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
