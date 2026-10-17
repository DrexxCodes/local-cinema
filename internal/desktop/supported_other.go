//go:build !windows

package desktop

// On Linux and macOS the program stays a plain server, as before.

// Supported says whether this system has the control panel.
func Supported() bool { return false }

// Run does nothing here.
func Run(Options) {}

// ClaimInstance always succeeds here.
func ClaimInstance(bool) bool { return true }

// ShowError does nothing here, the message is in the log.
func ShowError(string, string) {}
