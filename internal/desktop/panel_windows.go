package desktop

import (
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmLogLines = wmApp + 1 // new log lines are waiting
	wmState    = wmApp + 2 // the server state changed

	idStart = 101
	idStop  = 102
	idAdmin = 103
	idCopy  = 104
	idLogs  = 105
	idLog   = 200

	footerText = "Developed & Managed by Drexx Technologies"

	maxQueued  = 5000   // log lines waiting for the window
	maxLogChar = 200000 // characters kept in the log box
)

// cur is the panel the window procedure talks to (there is only ever one window).
var cur *panel

// windowProc is the callback Windows calls with window messages. Windows can only call a
// plain function, so it hands over to the one panel through cur.
func windowProc(hwnd, m, wp, lp uintptr) uintptr {
	if cur != nil {
		defer func() {
			if r := recover(); r != nil {
				cur.flog.Write("window error: " + toString(r))
			}
		}()
		return cur.handle(hwnd, uint32(m), wp, lp)
	}
	r, _, _ := procDefWindowProc.Call(hwnd, m, wp, lp)
	return r
}

type panel struct {
	opts  Options
	flog  *fileLog
	scale float64
	server *server
	hwnd   atomic.Uintptr
	hInst  windows.Handle
	ready  bool // controls exist (touched only on the window thread)
	fontBody, fontTitle, fontMono uintptr
	title, status, adminLbl       uintptr
	guestLbl                      uintptr
	btnStart, btnStop, btnAdmin   uintptr
	btnCopy, btnLogs, logEdit     uintptr
	footer                        uintptr
	qmu   sync.Mutex
	queue []string
	smu        sync.Mutex
	state      serverState
	detail     string
	autoOpened bool
}

func newPanel(opts Options, flog *fileLog) *panel { return &panel{opts: opts, flog: flog} }

func toString(v any) string {
	if e, ok := v.(error); ok {
		return e.Error()
	}
	if s, ok := v.(string); ok {
		return s
	}
	return "unknown error"
}

func (p *panel) px(n int) int { return int(float64(n)*p.scale + 0.5) }

// log adds a line to the log box and the log file. It is safe to call from any goroutine.
func (p *panel) log(line string) {
	line = time.Now().Format("15:04:05") + "  " + line
	p.flog.Write(line)
	p.qmu.Lock()
	p.queue = append(p.queue, line)
	if len(p.queue) > maxQueued {
		p.queue = p.queue[len(p.queue)-maxQueued:]
	}
	p.qmu.Unlock()
	if h := p.hwnd.Load(); h != 0 {
		_, _, _ = procPostMessage.Call(h, wmLogLines, 0, 0)
	}
}

func (p *panel) onServerState(st serverState, detail string) {
	p.smu.Lock()
	p.state, p.detail = st, detail
	first := st == stateRunning && !p.autoOpened && !p.opts.Background
	if first {
		p.autoOpened = true
	}
	p.smu.Unlock()
	if h := p.hwnd.Load(); h != 0 {
		_, _, _ = procPostMessage.Call(h, wmState, 0, 0)
	}
	if first {
		// On a fresh install, open the setup wizard in the browser without being asked.
		go func() {
			if u := StartURL(p.opts.Port); strings.HasSuffix(u, "/setup/") {
				openBrowser(u)
			}
		}()
	}
}

// run creates the window and runs its message loop until it is closed.
func (p *panel) run() error {
	cur = p
	p.scale = dpiScale()
	enableVisualStyles()
	p.fontBody = makeFont("Segoe UI", 9, 400, p.scale)
	p.fontTitle = makeFont("Segoe UI", 16, 700, p.scale)
	p.fontMono = makeFont("Consolas", 9, 400, p.scale)

	srv, err := newServer(p.opts.Port, p.log, p.onServerState)
	if err != nil {
		return err
	}
	p.server = srv
	p.state = stateStarting // the server is started as soon as the window is up

	h, _, e := procGetModuleHandle.Call(0)
	if h == 0 {
		return e
	}
	hInst := windows.Handle(h)
	p.hInst = hInst

	clsName, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return err
	}
	cursor, _, _ := procLoadCursor.Call(0, idcArrow)
	wc := wndClassEx{
		lpfnWndProc:   syscall.NewCallback(windowProc),
		hInstance:     hInst,
		hCursor:       windows.Handle(cursor),
		hbrBackground: windows.Handle(colorBtnFace + 1),
		lpszClassName: clsName,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	if atom, _, e := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		return e
	}

	title := WindowTitle
	if v := p.opts.Version; v != "" && v != "dev" {
		title += " " + v
	}
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	hwnd, _, e := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(clsName)), uintptr(unsafe.Pointer(titlePtr)),
		wsOverlappedWindow, cwUseDefault, cwUseDefault, u(p.px(700)), u(p.px(560)), 0, 0, uintptr(hInst), 0)
	if hwnd == 0 {
		return e
	}
	p.hwnd.Store(hwnd)
	setWindowIcon(hwnd)

	p.createControls(title)
	p.ready = true
	p.layout()
	p.refreshState()

	show := uintptr(swShowNormal)
	if p.opts.Background {
		show = swShowMinNoActive
	}
	_, _, _ = procShowWindow.Call(hwnd, show)

	if p.opts.LogDir != "" {
		p.log(title + " started. Logs are saved in " + p.opts.LogDir)
	} else {
		p.log(title + " started.")
	}
	go p.server.Start()

	var m msg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 { // 0 is WM_QUIT, -1 is an error
			break
		}
		if d, _, _ := procIsDialogMessage.Call(hwnd, uintptr(unsafe.Pointer(&m))); d != 0 {
			continue // Tab and arrow keys moved between the buttons
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

// child creates one control inside the window.
func (p *panel) child(ex uint32, class, text string, style uint32, id int, font uintptr) uintptr {
	c, err1 := windows.UTF16PtrFromString(class)
	t, err2 := windows.UTF16PtrFromString(text)
	if err1 != nil || err2 != nil {
		return 0
	}
	h, _, _ := procCreateWindowEx.Call(uintptr(ex), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(t)),
		uintptr(style|wsChild|wsVisible), 0, 0, 0, 0, p.hwnd.Load(), u(id), uintptr(p.hInst), 0)
	if h != 0 && font != 0 {
		_, _, _ = procSendMessage.Call(h, wmSetFont, font, 1)
	}
	return h
}

func (p *panel) createControls(title string) {
	p.title = p.child(0, "STATIC", title, ssLeft|ssNoPrefix, 0, p.fontTitle)
	p.status = p.child(0, "STATIC", "Status:  Starting...", ssLeft|ssNoPrefix, 0, p.fontBody)
	p.adminLbl = p.child(0, "STATIC", "This computer:  "+BaseURL(p.opts.Port)+"/admin/", ssLeft|ssNoPrefix, 0, p.fontBody)
	p.guestLbl = p.child(0, "STATIC", "", ssLeft|ssNoPrefix, 0, p.fontBody)

	p.btnStart = p.child(0, "BUTTON", "Start server", wsTabstop|bsPushBtn, idStart, p.fontBody)
	p.btnStop = p.child(0, "BUTTON", "Stop server", wsTabstop|bsPushBtn, idStop, p.fontBody)
	p.btnAdmin = p.child(0, "BUTTON", "Open admin", wsTabstop|bsPushBtn, idAdmin, p.fontBody)
	p.btnCopy = p.child(0, "BUTTON", "Copy guest link", wsTabstop|bsPushBtn, idCopy, p.fontBody)
	p.btnLogs = p.child(0, "BUTTON", "Open logs folder", wsTabstop|bsPushBtn, idLogs, p.fontBody)

	p.logEdit = p.child(wsExClientEdge, "EDIT", "", wsVscroll|esMultiline|esAutoVScr|esReadOnly, idLog, p.fontMono)
	_, _, _ = procSendMessage.Call(p.logEdit, emLimitText, 1<<24, 0)

	// No ampersand trouble here: ssNoPrefix makes Windows show "&" as a normal character.
	p.footer = p.child(0, "STATIC", footerText, ssCenter|ssNoPrefix, 0, p.fontBody)
}

func moveTo(h uintptr, x, y, w, hgt int) {
	_, _, _ = procMoveWindow.Call(h, u(x), u(y), u(w), u(hgt), 1)
}

// layout places the controls to fit the window.
func (p *panel) layout() {
	var rc rect
	_, _, _ = procGetClientRect.Call(p.hwnd.Load(), uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.right-rc.left), int(rc.bottom-rc.top)
	m := p.px(16)
	full := w - 2*m
	y := m

	moveTo(p.title, m, y, full, p.px(30))
	y += p.px(38)
	for _, lbl := range []uintptr{p.status, p.adminLbl, p.guestLbl} {
		moveTo(lbl, m, y, full, p.px(20))
		y += p.px(22)
	}
	y += p.px(8)

	x := m
	for _, b := range []struct {
		h uintptr
		w int
	}{{p.btnStart, 100}, {p.btnStop, 100}, {p.btnAdmin, 110}, {p.btnCopy, 130}, {p.btnLogs, 130}} {
		moveTo(b.h, x, y, p.px(b.w), p.px(30))
		x += p.px(b.w) + p.px(8)
	}
	y += p.px(30) + p.px(12)

	footerH := p.px(22)
	moveTo(p.footer, m, h-footerH-p.px(8), full, footerH)
	moveTo(p.logEdit, m, y, full, h-y-m-footerH-p.px(6))
}

func setText(h uintptr, s string) {
	if ptr, err := windows.UTF16PtrFromString(s); err == nil {
		_, _, _ = procSetWindowText.Call(h, uintptr(unsafe.Pointer(ptr)))
	}
}

func enable(h uintptr, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	_, _, _ = procEnableWindow.Call(h, v)
}

// refreshState updates the status line and the buttons.
func (p *panel) refreshState() {
	p.smu.Lock()
	st, detail := p.state, p.detail
	p.smu.Unlock()

	switch st {
	case stateStarting:
		setText(p.status, "Status:  Starting...")
	case stateRunning:
		setText(p.status, "Status:  Running")
	default:
		s := "Status:  Stopped"
		if detail != "" && detail != "stopped" {
			s += "  (" + detail + ")"
		}
		setText(p.status, s)
	}
	enable(p.btnStart, st == stateStopped)
	enable(p.btnStop, st != stateStopped)
	setText(p.guestLbl, "Guests (phones, TVs):  "+guestLink(p.opts.Port))
}

// flushLog moves waiting log lines into the log box.
func (p *panel) flushLog() {
	p.qmu.Lock()
	lines := p.queue
	p.queue = nil
	p.qmu.Unlock()
	if len(lines) == 0 {
		return
	}
	text, err := windows.UTF16PtrFromString(strings.Join(lines, "\r\n") + "\r\n")
	if err != nil {
		return
	}
	n, _, _ := procGetWindowTextLength.Call(p.logEdit)
	if n > maxLogChar { // forget the oldest half so the box never grows without limit
		_, _, _ = procSendMessage.Call(p.logEdit, emSetSel, 0, n/2)
		_, _, _ = procSendMessage.Call(p.logEdit, emReplaceSel, 0, uintptr(unsafe.Pointer(emptyText)))
		n, _, _ = procGetWindowTextLength.Call(p.logEdit)
	}
	_, _, _ = procSendMessage.Call(p.logEdit, emSetSel, n, n)
	_, _, _ = procSendMessage.Call(p.logEdit, emReplaceSel, 0, uintptr(unsafe.Pointer(text)))
	_, _, _ = procSendMessage.Call(p.logEdit, emScrollCar, 0, 0)
}

var emptyText, _ = windows.UTF16PtrFromString("")

// command handles a button click.
func (p *panel) command(id int) {
	switch id {
	case idStart:
		go p.server.Start()
	case idStop:
		go p.server.Stop()
	case idAdmin:
		go openBrowser(StartURL(p.opts.Port))
	case idCopy:
		go func() {
			if err := copyText(guestLink(p.opts.Port)); err != nil {
				p.log("Could not copy the link: " + err.Error())
				return
			}
			setText(p.btnCopy, "Copied!")
			time.AfterFunc(2*time.Second, func() { setText(p.btnCopy, "Copy guest link") })
		}()
	case idLogs:
		if p.opts.LogDir != "" {
			go openFolder(p.opts.LogDir)
		}
	}
}

// confirmClose asks before stopping a cinema that people may be watching.
func (p *panel) confirmClose() bool {
	if p.server.State() == stateStopped {
		return true
	}
	ans := messageBox(p.hwnd.Load(), WindowTitle,
		"Closing this window stops Local Cinema, and anyone watching will be disconnected.\n\nTo keep it running, minimize the window instead.\n\nClose and stop the cinema?",
		mbYesNo|mbIconQuestion|mbDefButton2)
	return ans == idYes
}

// handle answers the window messages we care about.
func (p *panel) handle(hwnd uintptr, m uint32, wp, lp uintptr) uintptr {
	switch m {
	case wmSize:
		if p.ready {
			p.layout()
		}
		return 0
	case wmCommand:
		p.command(int(wp & 0xFFFF))
		return 0
	case wmLogLines:
		p.flushLog()
		return 0
	case wmState:
		p.refreshState()
		return 0
	case wmClose:
		if p.confirmClose() {
			p.server.Stop()
			_, _, _ = procDestroyWindow.Call(hwnd)
		}
		return 0
	case wmDestroy:
		_, _, _ = procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(hwnd, uintptr(m), wp, lp)
	return r
}

