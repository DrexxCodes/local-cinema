package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathFor(t *testing.T) {
	cases := map[string]string{
		`{"setup_complete":true}`:  "/admin/",
		`{"setup_complete":false}`: "/setup/",
		`{}`:                       "/admin/",
		`not json`:                 "/admin/",
	}
	for body, want := range cases {
		if got := pathFor([]byte(body)); got != want {
			t.Errorf("pathFor(%s) = %s, want %s", body, got, want)
		}
	}
}

func TestCleanLine(t *testing.T) {
	if got := cleanLine("\x1b[32mServer started\x1b[0m\r\n"); got != "Server started" {
		t.Errorf("cleanLine = %q", got)
	}
}

func TestReadLines(t *testing.T) {
	var got []string
	readLines(strings.NewReader("one\r\n\r\n\x1b[1mtwo\x1b[0m\nthree"), func(s string) { got = append(got, s) })
	if strings.Join(got, "|") != "one|two|three" {
		t.Errorf("lines = %v", got)
	}
}

func TestFileLog(t *testing.T) {
	dir := t.TempDir()
	l := openFileLog(dir)
	if l == nil {
		t.Fatal("log not opened")
	}
	l.Write("hello")
	l.Close()
	b, err := os.ReadFile(filepath.Join(dir, "local-cinema.log"))
	if err != nil || string(b) != "hello\n" {
		t.Errorf("log = %q, %v", b, err)
	}
	var none *fileLog
	none.Write("ignored") // a nil log must not panic
	none.Close()
}
