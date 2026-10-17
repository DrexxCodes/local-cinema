package desktop

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// instanceMutex is held for as long as the program runs.
var instanceMutex windows.Handle

// ClaimInstance makes sure only one panel runs per Windows user. When another copy is already
// running it brings that window to the front (unless this one was started in the background)
// and returns false, which means the caller should exit.
func ClaimInstance(background bool) bool {
	name, err := windows.UTF16PtrFromString(`Local\LocalCinemaDesktop`)
	if err != nil {
		return true
	}
	h, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) || errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		if h != 0 {
			_ = windows.CloseHandle(h)
		}
		if !background {
			bringToFront()
		}
		return false
	}
	if err == nil {
		instanceMutex = h
	}
	return true
}

// bringToFront finds the running panel's window. It retries because that copy may still be
// opening its window.
func bringToFront() {
	cls, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return
	}
	for i := 0; i < 20; i++ {
		hwnd, _, _ := procFindWindow.Call(uintptr(unsafe.Pointer(cls)), 0)
		if hwnd != 0 {
			focusWindow(hwnd)
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}
