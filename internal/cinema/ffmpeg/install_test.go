package ffmpeg

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sumOf(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// release serves a fake GitHub release: SHA256SUMS and one zip.
func release(t *testing.T, asset string, zipBytes []byte, sums string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sums))
	})
	mux.HandleFunc("/"+asset, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(zipBytes)))
		_, _ = w.Write(zipBytes)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newInstaller(srv *httptest.Server, dir string) *Installer {
	return &Installer{BaseURL: srv.URL, Dir: dir, GOOS: "windows", GOARCH: "amd64", Client: srv.Client(), UserAgent: "test"}
}

func TestInstallUnpacksAndVerifies(t *testing.T) {
	z := makeZip(t, map[string]string{
		"ffmpeg.exe":     "FFMPEG",
		"ffprobe.exe":    "FFPROBE",
		"LICENSE.txt":    "GPL",
		"SOURCE.txt":     "where it came from",
		"notes/junk.dll": "must not be extracted",
		"evil/../x.exe":  "must not be extracted",
	})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, sumOf(z)+"  ffmpeg-windows-amd64.zip\n"+strings.Repeat("0", 64)+"  ffmpeg-linux-amd64.zip\n")
	dir := filepath.Join(t.TempDir(), "bin")
	in := newInstaller(srv, dir)

	var mu sync.Mutex
	var steps []string
	var last int
	in.OnStep = func(s string) { mu.Lock(); steps = append(steps, s); mu.Unlock() }
	in.OnPercent = func(p int) { mu.Lock(); last = p; mu.Unlock() }

	if err := in.Install(context.Background()); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"ffmpeg.exe": "FFMPEG", "ffprobe.exe": "FFPROBE", "LICENSE.txt": "GPL"} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", name, got, err, want)
		}
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		switch e.Name() {
		case "ffmpeg.exe", "ffprobe.exe", "LICENSE.txt", "SOURCE.txt":
		default:
			t.Errorf("unexpected file %q in the tools folder (leftover or not on the allow list)", e.Name())
		}
	}
	if last != 100 {
		t.Errorf("download progress ended at %d, want 100", last)
	}
	want := []string{"Checking the download", "Downloading FFmpeg", "Unpacking"}
	if strings.Join(steps, "|") != strings.Join(want, "|") {
		t.Errorf("steps = %v, want %v", steps, want)
	}
}

func TestInstallRejectsWrongChecksum(t *testing.T) {
	z := makeZip(t, map[string]string{"ffmpeg.exe": "FFMPEG"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, strings.Repeat("a", 64)+"  ffmpeg-windows-amd64.zip\n")
	dir := t.TempDir()
	err := newInstaller(srv, dir).Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("a tampered download must be rejected, got %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("nothing should be left behind after a failed check, found %d files", len(entries))
	}
}

func TestInstallNeedsAnEntryForThisComputer(t *testing.T) {
	z := makeZip(t, map[string]string{"ffmpeg.exe": "x"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, sumOf(z)+"  ffmpeg-linux-amd64.zip\n")
	err := newInstaller(srv, t.TempDir()).Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no entry") {
		t.Errorf("want a 'no entry' error, got %v", err)
	}
}

func TestInstallAcceptsBinaryModeChecksumLines(t *testing.T) {
	z := makeZip(t, map[string]string{"ffmpeg.exe": "x"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, strings.ToUpper(sumOf(z))+" *ffmpeg-windows-amd64.zip\n")
	if err := newInstaller(srv, t.TempDir()).Install(context.Background()); err != nil {
		t.Errorf("sha256sum's '*name' form and upper case hex should work: %v", err)
	}
}

func TestInstallZipWithoutFfmpeg(t *testing.T) {
	z := makeZip(t, map[string]string{"readme.md": "nothing useful"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, sumOf(z)+"  ffmpeg-windows-amd64.zip\n")
	err := newInstaller(srv, t.TempDir()).Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "did not contain ffmpeg") {
		t.Errorf("got %v", err)
	}
}

func TestInstallServerErrors(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	err := newInstaller(srv, t.TempDir()).Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("want the status code in the error, got %v", err)
	}
}

func TestInstallReplacesExistingFiles(t *testing.T) {
	z := makeZip(t, map[string]string{"ffmpeg.exe": "NEW"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, sumOf(z)+"  ffmpeg-windows-amd64.zip\n")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ffmpeg.exe"), []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := newInstaller(srv, dir).Install(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "ffmpeg.exe")); string(got) != "NEW" {
		t.Errorf("the old copy should be replaced, got %q", got)
	}
}

func TestInstallCancelled(t *testing.T) {
	z := makeZip(t, map[string]string{"ffmpeg.exe": "x"})
	srv := release(t, "ffmpeg-windows-amd64.zip", z, sumOf(z)+"  ffmpeg-windows-amd64.zip\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := newInstaller(srv, t.TempDir()).Install(ctx); err == nil {
		t.Error("a cancelled context should stop the install")
	}
}

func TestSupported(t *testing.T) {
	cases := []struct {
		name string
		in   Installer
		want bool
	}{
		{"windows amd64", Installer{Repo: "me/cinema", Bundle: "2026.10", GOOS: "windows", GOARCH: "amd64"}, true},
		{"linux arm64", Installer{Repo: "me/cinema", Bundle: "2026.10", GOOS: "linux", GOARCH: "arm64"}, true},
		{"macOS has no download", Installer{Repo: "me/cinema", Bundle: "2026.10", GOOS: "darwin", GOARCH: "arm64"}, false},
		{"windows 32 bit", Installer{Repo: "me/cinema", Bundle: "2026.10", GOOS: "windows", GOARCH: "386"}, false},
		{"development build has no repo", Installer{Bundle: "2026.10", GOOS: "windows", GOARCH: "amd64"}, false},
		{"no bundle", Installer{Repo: "me/cinema", GOOS: "windows", GOARCH: "amd64"}, false},
	}
	for _, c := range cases {
		if got := c.in.Supported(); got != c.want {
			t.Errorf("%s: Supported() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestInstallUnsupportedBuild(t *testing.T) {
	in := &Installer{Dir: t.TempDir(), GOOS: "darwin", GOARCH: "arm64"}
	if err := in.Install(context.Background()); err == nil {
		t.Error("a build without a download should explain itself, not try the network")
	}
}

func TestAssetAndBaseURL(t *testing.T) {
	in := &Installer{Repo: "me/cinema", Bundle: "2026.10", GOOS: "linux", GOARCH: "arm64"}
	if got, want := in.Asset(), "ffmpeg-linux-arm64.zip"; got != want {
		t.Errorf("Asset = %q, want %q", got, want)
	}
	if got, want := in.base(), "https://github.com/me/cinema/releases/download/ffmpeg-2026.10/"; got != want {
		t.Errorf("base = %q, want %q", got, want)
	}
}

func TestExpectedHash(t *testing.T) {
	sums := "ABC123  ffmpeg-windows-amd64.zip\nbad line\n" + "def456 *ffmpeg-linux-amd64.zip\n"
	if got, err := ExpectedHash(strings.NewReader(sums), "ffmpeg-windows-amd64.zip"); err != nil || got != "abc123" {
		t.Errorf("got %q, %v", got, err)
	}
	if got, err := ExpectedHash(strings.NewReader(sums), "ffmpeg-linux-amd64.zip"); err != nil || got != "def456" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, err := ExpectedHash(strings.NewReader(sums), "ffmpeg-darwin-arm64.zip"); err == nil {
		t.Error("a missing entry should be an error")
	}
}
