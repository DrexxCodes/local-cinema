package desktop

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// The few Windows calls the panel needs, declared once.

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	comctl32 = windows.NewLazySystemDLL("comctl32.dll")

	procRegisterClassEx     = user32.NewProc("RegisterClassExW")
	procCreateWindowEx      = user32.NewProc("CreateWindowExW")
	procDefWindowProc       = user32.NewProc("DefWindowProcW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procGetMessage          = user32.NewProc("GetMessageW")
	procIsDialogMessage     = user32.NewProc("IsDialogMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
	procPostMessage         = user32.NewProc("PostMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procSendMessage         = user32.NewProc("SendMessageW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procMoveWindow          = user32.NewProc("MoveWindow")
	procGetClientRect       = user32.NewProc("GetClientRect")
	procSetWindowText       = user32.NewProc("SetWindowTextW")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	procEnableWindow        = user32.NewProc("EnableWindow")
	procLoadCursor          = user32.NewProc("LoadCursorW")
	procLoadImage           = user32.NewProc("LoadImageW")
	procMessageBox          = user32.NewProc("MessageBoxW")
	procFindWindow          = user32.NewProc("FindWindowW")
	procSetProcessDPIAware  = user32.NewProc("SetProcessDPIAware")
	procGetDpiForSystem     = user32.NewProc("GetDpiForSystem")

	procCreateFont = gdi32.NewProc("CreateFontW")

	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
	procCreateActCtx   = kernel32.NewProc("CreateActCtxW")
	procActivateActCtx = kernel32.NewProc("ActivateActCtx")

	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
)

const className = "LocalCinemaPanel"

// Window and control styles, messages and other numbers from the Windows headers.
const (
	wsOverlappedWindow = 0x00CF0000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsTabstop          = 0x00010000
	wsVscroll          = 0x00200000
	wsExClientEdge     = 0x00000200

	ssLeft      = 0x0000
	ssCenter    = 0x0001
	ssNoPrefix  = 0x0080
	bsPushBtn   = 0x0000
	esMultiline = 0x0004
	esAutoVScr  = 0x0040
	esReadOnly  = 0x0800

	wmDestroy    = 0x0002
	wmSize       = 0x0005
	wmClose      = 0x0010
	wmSetFont    = 0x0030
	wmSetIcon    = 0x0080
	wmCommand    = 0x0111
	wmApp        = 0x8000
	emSetSel     = 0x00B1
	emScrollCar  = 0x00B7
	emReplaceSel = 0x00C2
	emLimitText  = 0x00C5

	swShowNormal       = 1
	swShowMinNoActive  = 7
	swRestore          = 9
	cwUseDefault       = 0x80000000
	idcArrow           = 32512
	colorBtnFace       = 15
	iccStandardClasses = 0x4000

	mbOk           = 0x00000000
	mbYesNo        = 0x00000004
	mbIconError    = 0x00000010
	mbIconQuestion = 0x00000020
	mbDefButton2   = 0x00000100
	idYes          = 6

	iconSmall     = 0
	iconBig       = 1
	imageIcon     = 1
	lrLoadFromFile = 0x00000010
	lrDefaultSize = 0x00000040
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
}

type point struct{ x, y int32 }

type msg struct {
	hwnd    windows.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
	private uint32
}

type rect struct{ left, top, right, bottom int32 }

type actCtx struct {
	cbSize                 uint32
	dwFlags                uint32
	lpSource               *uint16
	wProcessorArchitecture uint16
	wLangID                uint16
	lpAssemblyDirectory    *uint16
	lpResourceName         *uint16
	lpApplicationName      *uint16
	hModule                windows.Handle
}

type initCommonControlsEx struct {
	size uint32
	icc  uint32
}

// messageBox shows a message box and returns the button that was pressed.
func messageBox(owner uintptr, title, text string, flags uint32) int {
	t, err1 := windows.UTF16PtrFromString(title)
	m, err2 := windows.UTF16PtrFromString(text)
	if err1 != nil || err2 != nil {
		return 0
	}
	r, _, _ := procMessageBox.Call(owner, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(flags))
	return int(r)
}

// u turns an int into a call argument. Negative numbers are sign extended, as Windows expects.
func u(n int) uintptr { return uintptr(int32(n)) }
