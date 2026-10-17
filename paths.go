package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// defaultDataDir decides where the database and cache live.
//
//   - CINEMA_DATA_DIR wins when set.
//   - An installed copy (the installer writes a file called "installed" next to the
//     program) keeps its data outside the program folder, so upgrades and uninstalls
//     leave it alone: ProgramData on Windows, /var/lib on Linux, Application Support on macOS.
//   - Otherwise "" is returned and PocketBase uses pb_data next to the program (portable).
//
// The --dir flag still overrides everything.
func defaultDataDir() (dir string, installed bool) {
	if d := os.Getenv("CINEMA_DATA_DIR"); d != "" {
		return d, false
	}
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), "installed")); err != nil {
		return "", false
	}
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("ProgramData")
		if base == "" {
			base = `C:\ProgramData`
		}
		return filepath.Join(base, "Local Cinema", "data"), true
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "Local Cinema", "data"), true
		}
	default:
		return "/var/lib/local-cinema/data", true
	}
	return "", false
}

// logDirFor is where the control panel writes local-cinema.log: next to the data folder for
// an installed copy (C:\ProgramData\Local Cinema\logs), a "logs" folder beside the program otherwise.
func logDirFor(dataDir string, installed bool) string {
	if installed && dataDir != "" {
		return filepath.Join(filepath.Dir(dataDir), "logs")
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "logs")
	}
	return ""
}

// setupFileLogging sends logs to <data folder>/../logs/local-cinema.log.
func setupFileLogging(dataDir string) {
	dir := filepath.Join(filepath.Dir(dataDir), "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "local-cinema.log")

	// The file is moved aside once it passes 5 MB so it never grows without limit
	if st, err := os.Stat(path); err == nil && st.Size() > 5<<20 {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Writer(f), nil)))
	log.SetOutput(f)
}
