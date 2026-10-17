// Package ffmpeg finds and runs FFmpeg and ffprobe, and installs them when missing.
//
// Local Cinema runs FFmpeg as a separate program. It is never linked into the binary.
package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Tools locates the ffmpeg and ffprobe programs. They are found in this order:
// CINEMA_FFMPEG_DIR, the "bin" folder next to the data folder (where the installer and the
// Settings button put them), the folder of the program, and finally the PATH.
type Tools struct {
	mu       sync.Mutex
	binDir   string // <parent of data dir>/bin
	ffmpeg   string
	ffprobe  string
	searched bool
	versions map[string]string
}

// NewTools returns a locator with no data directory set yet.
func NewTools() *Tools { return &Tools{versions: map[string]string{}} }

// SetDataDir tells the locator where the data folder is. The tools folder is "bin" next to it.
func (t *Tools) SetDataDir(dir string) {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	t.mu.Lock()
	t.binDir = filepath.Join(filepath.Dir(dir), "bin")
	t.searched = false
	t.mu.Unlock()
}

// Dir is where downloaded FFmpeg files are stored.
func (t *Tools) Dir() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.binDir
}

// FFmpeg returns the ffmpeg program, or "" when it is not installed.
func (t *Tools) FFmpeg() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.search()
	return t.ffmpeg
}

// FFprobe returns the ffprobe program, or "" when it is not installed.
func (t *Tools) FFprobe() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.search()
	return t.ffprobe
}

// Reset makes the next lookup search again, for example after FFmpeg was installed.
func (t *Tools) Reset() {
	t.mu.Lock()
	t.searched = false
	t.versions = map[string]string{}
	t.mu.Unlock()
}

// Version returns the first line of "<bin> -version", for example "ffmpeg version 7.1".
// The answer is cached per program. It is "" when bin is empty or cannot be run.
func (t *Tools) Version(bin string) string {
	if bin == "" {
		return ""
	}
	t.mu.Lock()
	if v, ok := t.versions[bin]; ok {
		t.mu.Unlock()
		return v
	}
	t.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, _ := exec.CommandContext(ctx, bin, "-version").Output()
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])

	t.mu.Lock()
	t.versions[bin] = line
	t.mu.Unlock()
	return line
}

// search fills in the program paths once. The caller holds the lock.
func (t *Tools) search() {
	if t.searched {
		return
	}
	t.ffmpeg, t.ffprobe = t.find("ffmpeg"), t.find("ffprobe")
	t.searched = true
}

// find looks for one program. The caller holds the lock.
func (t *Tools) find(name string) string {
	file := ExeName(name)
	var dirs []string
	if d := os.Getenv("CINEMA_FFMPEG_DIR"); d != "" {
		dirs = append(dirs, d)
	}
	dirs = append(dirs, t.binDir)
	if exe, err := os.Executable(); err == nil {
		ed := filepath.Dir(exe)
		dirs = append(dirs, ed, filepath.Join(ed, "bin"))
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		p := filepath.Join(d, file)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	p, _ := exec.LookPath(name)
	return p
}

// ExeName adds ".exe" on Windows.
func ExeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
