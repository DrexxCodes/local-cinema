package desktop

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const maxLogSize = 5 << 20 // the log is moved aside once it passes 5 MB

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// cleanLine removes colour codes and the line ending.
func cleanLine(s string) string {
	return strings.TrimRight(ansiRE.ReplaceAllString(s, ""), "\r\n")
}

// readLines calls fn for every non-empty line r produces, until r ends.
func readLines(r io.Reader, fn func(string)) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		if s := cleanLine(sc.Text()); s != "" {
			fn(s)
		}
	}
}

// fileLog appends lines to <dir>/local-cinema.log. A nil *fileLog does nothing.
type fileLog struct {
	mu sync.Mutex
	f  *os.File
}

func openFileLog(dir string) *fileLog {
	if dir == "" || os.MkdirAll(dir, 0o755) != nil {
		return nil
	}
	path := filepath.Join(dir, "local-cinema.log")
	if st, err := os.Stat(path); err == nil && st.Size() > maxLogSize {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil
	}
	return &fileLog{f: f}
}

func (l *fileLog) Write(line string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.f.WriteString(line + "\n")
}

func (l *fileLog) Close() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.f.Close()
}
