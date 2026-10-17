package desktop

import (
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// manifestXML asks Windows for the modern look of buttons and boxes (Common Controls 6).
const manifestXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <dependency>
    <dependentAssembly>
      <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
    </dependentAssembly>
  </dependency>
</assembly>`

// enableVisualStyles turns on the modern look at run time, so the program needs no embedded
// manifest. If anything fails the panel simply looks classic.
func enableVisualStyles() {
	path := filepath.Join(cacheDir(), "panel.manifest")
	if err := os.WriteFile(path, []byte(manifestXML), 0o644); err != nil {
		return
	}
	src, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	ctx := actCtx{lpSource: src}
	ctx.cbSize = uint32(unsafe.Sizeof(ctx))
	h, _, _ := procCreateActCtx.Call(uintptr(unsafe.Pointer(&ctx)))
	if h == 0 || h == ^uintptr(0) {
		return
	}
	var cookie uintptr
	if ok, _, _ := procActivateActCtx.Call(h, uintptr(unsafe.Pointer(&cookie))); ok == 0 {
		return
	}
	icc := initCommonControlsEx{icc: iccStandardClasses}
	icc.size = uint32(unsafe.Sizeof(icc))
	_, _, _ = procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
}

// dpiScale makes the panel sharp on high resolution screens (1.0 = 100%, 1.5 = 150%).
func dpiScale() float64 {
	_, _, _ = procSetProcessDPIAware.Call()
	if procGetDpiForSystem.Find() == nil {
		if dpi, _, _ := procGetDpiForSystem.Call(); dpi > 0 {
			return float64(dpi) / 96
		}
	}
	return 1
}

// makeFont creates a font of the given size in points.
func makeFont(face string, points int, weight int, scale float64) uintptr {
	name, err := windows.UTF16PtrFromString(face)
	if err != nil {
		return 0
	}
	height := -int(float64(points)*96/72*scale + 0.5)
	h, _, _ := procCreateFont.Call(u(height), 0, 0, 0, u(weight), 0, 0, 0,
		1 /* DEFAULT_CHARSET */, 0, 0, 5 /* CLEARTYPE_QUALITY */, 0, uintptr(unsafe.Pointer(name)))
	return h
}
