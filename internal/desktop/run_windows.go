package desktop

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// Supported says whether this system has the control panel.
// Set CINEMA_NO_DESKTOP=1 to run as a plain server on Windows too.
func Supported() bool { return os.Getenv("CINEMA_NO_DESKTOP") == "" }

// Run opens the panel, starts the server and blocks until the window is closed. Every failure
// ends up in the log file and in a message box, because the installed program has no console.
func Run(opts Options) {
	runtime.LockOSThread() // the window and its message loop live on this thread

	flog := openFileLog(opts.LogDir)
	defer flog.Close()
	defer func() {
		if r := recover(); r != nil {
			flog.Write(fmt.Sprintf("panel crashed: %v\n%s", r, debug.Stack()))
			ShowError(WindowTitle, fmt.Sprintf("Local Cinema hit an unexpected error and will close.\n\n%v", r))
		}
	}()

	p := newPanel(opts, flog)
	if err := p.run(); err != nil {
		flog.Write("panel failed: " + err.Error())
		ShowError(WindowTitle, "Local Cinema could not open its window.\n\n"+err.Error()+
			"\n\nYou can still use it from a web browser if you run:\n  local-cinema.exe serve")
	}
}
