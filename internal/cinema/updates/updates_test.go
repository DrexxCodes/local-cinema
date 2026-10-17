package updates

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.2.0", "1.1.9", true},
		{"1.10.0", "1.9.9", true},
		{"v2.0.0", "1.9.9", true},
		{"1.2.0", "1.2.0", false},
		{"1.1.0", "1.2.0", false},
		{"1.2.0", "1.2.0-beta.1", true},
		{"1.2.0-beta.1", "1.2.0", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.a, c.b); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	n, pre := ParseVersion(" v1.2.3-rc.1 ")
	if n != [3]int{1, 2, 3} || !pre {
		t.Errorf("got %v %v", n, pre)
	}
	n, pre = ParseVersion("2.0+build5")
	if n != [3]int{2, 0, 0} || pre {
		t.Errorf("build metadata is not a pre-release: got %v %v", n, pre)
	}
}

func TestPickAsset(t *testing.T) {
	assets := []Asset{
		{"LocalCinema-Setup-1.2.0.exe", "u/win"},
		{"local-cinema_1.2.0_linux_amd64.tar.gz", "u/lin"},
		{"local-cinema_1.2.0_darwin_arm64.tar.gz", "u/mac"},
	}
	cases := []struct{ os, arch, name string }{
		{"windows", "amd64", "LocalCinema-Setup-1.2.0.exe"},
		{"linux", "amd64", "local-cinema_1.2.0_linux_amd64.tar.gz"},
		{"darwin", "arm64", "local-cinema_1.2.0_darwin_arm64.tar.gz"},
		{"linux", "arm64", ""},
	}
	for _, c := range cases {
		if n, _ := PickAsset(assets, c.os, c.arch); n != c.name {
			t.Errorf("%s/%s picked %q, want %q", c.os, c.arch, n, c.name)
		}
	}
}

func serve(t *testing.T, status int, body string) *Service {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/releases/latest" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "local-cinema/") {
			t.Errorf("GitHub needs a User-Agent, got %q", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	s := New("1.0.0", "o/r")
	s.APIBase = ts.URL
	s.GOOS, s.GOARCH = "windows", "amd64"
	return s
}

func TestCheckFindsNewRelease(t *testing.T) {
	s := serve(t, 200, `{"tag_name":"v1.1.0","body":"notes","html_url":"http://x/r","published_at":"2026-01-01T00:00:00Z",
		"assets":[{"name":"LocalCinema-Setup-1.1.0.exe","browser_download_url":"http://x/dl"}]}`)
	s.Check()
	u := s.Info()
	if !u.Enabled || !u.Available || u.Latest != "1.1.0" || u.Current != "1.0.0" {
		t.Errorf("unexpected info %+v", u)
	}
	if u.AssetName != "LocalCinema-Setup-1.1.0.exe" || u.AssetURL != "http://x/dl" || u.Error != "" {
		t.Errorf("unexpected asset or error %+v", u)
	}
	if s.Available() != "1.1.0" {
		t.Errorf("Available = %q", s.Available())
	}
}

func TestCheckUpToDate(t *testing.T) {
	s := serve(t, 200, `{"tag_name":"v1.0.0"}`)
	s.Check()
	if s.Info().Available || s.Available() != "" {
		t.Error("the same version is not an update")
	}
}

func TestCheckTruncatesLongNotes(t *testing.T) {
	s := serve(t, 200, `{"tag_name":"v2.0.0","body":"`+strings.Repeat("a", 5000)+`"}`)
	s.Check()
	if n := len(s.Info().Notes); n != 4003 {
		t.Errorf("notes length %d, want 4003", n)
	}
}

func TestCheckErrors(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{404, "", "No release"},
		{403, "", "limiting"},
		{429, "", "limiting"},
		{500, "", "status 500"},
		{200, "not json", "could not be read"},
		{200, `{"tag_name":""}`, "could not be read"},
	}
	for _, c := range cases {
		s := serve(t, c.status, c.body)
		s.Check()
		u := s.Info()
		if !strings.Contains(u.Error, c.want) || u.Available {
			t.Errorf("status %d: error %q, want it to contain %q", c.status, u.Error, c.want)
		}
	}
}

func TestCheckUnreachable(t *testing.T) {
	s := New("1.0.0", "o/r")
	s.APIBase = "http://127.0.0.1:1"
	s.Check()
	if !strings.Contains(s.Info().Error, "Could not reach") {
		t.Errorf("got %q", s.Info().Error)
	}
}

func TestDisabledForDevAndMissingRepo(t *testing.T) {
	for _, s := range []*Service{New("dev", "o/r"), New("1.0.0", "")} {
		s.APIBase = "http://127.0.0.1:1" // must never be called
		s.Check()
		u := s.Info()
		if u.Enabled || u.Error != "" || u.CheckedAt == "" {
			t.Errorf("%+v: expected disabled and quiet, but checked", u)
		}
	}
}
