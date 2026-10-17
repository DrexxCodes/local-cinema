package desktop

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

//go:embed assets/icon.ico
var iconICO []byte

// cacheDir holds small per-user files (a copy of the icon, the theme manifest).
func cacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "Local Cinema")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// focusWindow restores a minimized window and brings it to the front.
func focusWindow(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	_, _, _ = procShowWindow.Call(hwnd, swRestore)
	_, _, _ = procSetForegroundWindow.Call(hwnd)
}

// setWindowIcon gives the window the Local Cinema icon (title bar and taskbar).
func setWindowIcon(hwnd uintptr) {
	path := filepath.Join(cacheDir(), "icon.ico")
	if err := os.WriteFile(path, iconICO, 0o644); err != nil {
		return
	}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	big, _, _ := procLoadImage.Call(0, uintptr(unsafe.Pointer(p)), imageIcon, 0, 0, lrLoadFromFile|lrDefaultSize)
	small, _, _ := procLoadImage.Call(0, uintptr(unsafe.Pointer(p)), imageIcon, 16, 16, lrLoadFromFile)
	if big != 0 {
		_, _, _ = procSendMessage.Call(hwnd, wmSetIcon, iconBig, big)
	}
	if small != 0 {
		_, _, _ = procSendMessage.Call(hwnd, wmSetIcon, iconSmall, small)
	}
}

// openBrowser opens a link in the default web browser.
func openBrowser(url string) {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err == nil {
		go func() { _ = cmd.Wait() }()
	}
}

// openFolder shows a folder in Explorer.
func openFolder(dir string) {
	_ = os.MkdirAll(dir, 0o755)
	cmd := exec.Command("explorer.exe", dir)
	if err := cmd.Start(); err == nil {
		go func() { _ = cmd.Wait() }()
	}
}

// copyText puts text on the clipboard using the clip tool that ships with Windows.
func copyText(text string) error {
	cmd := exec.Command("clip")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// ShowError shows a message box. The installed program has no console, so this is how
// a start up failure becomes visible.
func ShowError(title, msg string) {
	messageBox(0, title, msg, mbOk|mbIconError)
}
