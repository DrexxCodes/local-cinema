// Package desktop is the Windows control panel for Local Cinema.
//
// Double-clicking the program opens a small native window (plain Windows controls, nothing
// to install) that starts the cinema server, shows its addresses and its log, and lets you
// stop and start it. The server itself runs as a separate hidden copy of this program
// (`local-cinema.exe serve`), so Stop and Start really restart it, and its output (including
// crash messages) shows up in the log box and in the log file.
//
// Phones and TVs keep using the cinema through their browsers. The panel is only for the
// person at the computer.
//
// Everything that touches Windows lives in the *_windows.go files. On other systems the
// package does nothing (see supported_other.go) and the program behaves as before.
package desktop

// Options tells the panel how the program was started.
type Options struct {
	// Port is the HTTP port the cinema listens on.
	Port string
	// Background starts the panel minimized (used when Windows starts it at log in).
	Background bool
	// Version is shown in the window title.
	Version string
	// LogDir is where local-cinema.log is written. Empty means no log file.
	LogDir string
}

// WindowTitle is the title of the panel window.
const WindowTitle = "Local Cinema"
