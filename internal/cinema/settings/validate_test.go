package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCinemaName(t *testing.T) {
	if n, err := CinemaName("  Drexx Cinema "); err != nil || n != "Drexx Cinema" {
		t.Errorf("got %q, %v", n, err)
	}
	if _, err := CinemaName("   "); err == nil {
		t.Error("a blank name is refused")
	}
	if _, err := CinemaName(strings.Repeat("x", 51)); err == nil {
		t.Error("51 characters is too long")
	}
	if _, err := CinemaName(strings.Repeat("x", 50)); err != nil {
		t.Error("50 characters is fine")
	}
}

func TestTagline(t *testing.T) {
	if s, err := Tagline("  "); err != nil || s != "" {
		t.Errorf("an empty tagline is allowed: %q, %v", s, err)
	}
	if _, err := Tagline(strings.Repeat("x", 121)); err == nil {
		t.Error("121 characters is too long")
	}
	if _, err := Tagline(strings.Repeat("x", 120)); err != nil {
		t.Error("120 characters is fine")
	}
}

func TestAccentColor(t *testing.T) {
	for _, ok := range []string{"#7C3AED", "#000000", "#abcdef", "#ABCDEF"} {
		if _, err := AccentColor(ok); err != nil {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"7C3AED", "#7C3AE", "#7C3AEDD", "#GGGGGG", "red", "", "#7C3AED "} {
		if _, err := AccentColor(bad); err == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
}

func TestHostName(t *testing.T) {
	if h, err := HostName("  My-Cinema "); err != nil || h != "my-cinema" {
		t.Errorf("got %q, %v", h, err)
	}
	for _, bad := range []string{"", "has space", "under_score", "dot.name", "é", strings.Repeat("a", 64)} {
		if _, err := HostName(bad); err == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
	if _, err := HostName(strings.Repeat("a", 63)); err != nil {
		t.Error("63 characters is fine")
	}
}

func TestMovieFolders(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	file := filepath.Join(a, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := MovieFolders([]string{" " + a + " ", "", b})
	if err != nil || len(got) != 2 || got[0] != a || got[1] != b {
		t.Errorf("got %v, %v", got, err)
	}
	if _, err := MovieFolders([]string{a, filepath.Join(a, "missing")}); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Errorf("a missing folder is refused and named, got %v", err)
	}
	if _, err := MovieFolders([]string{file}); err == nil {
		t.Error("a file is not a folder")
	}
	if _, err := MovieFolders(nil); err == nil {
		t.Error("at least one folder is required")
	}
	if _, err := MovieFolders([]string{"", "  "}); err == nil {
		t.Error("blank entries do not count as a folder")
	}
}

func TestMaxStreams(t *testing.T) {
	for in, want := range map[int]int{-5: 1, 0: 1, 1: 1, 10: 10, 1000: 1000, 5000: 1000} {
		if got := MaxStreams(in); got != want {
			t.Errorf("MaxStreams(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestTMDBKey(t *testing.T) {
	cases := []struct {
		key    string
		clear  bool
		value  string
		change bool
	}{
		{"abc123", false, "abc123", true},
		{"  abc  ", false, "abc", true},
		{"", false, "", false}, // blank leaves the stored key alone
		{"   ", false, "", false},
		{"abc", true, "", true}, // clear wins
		{"", true, "", true},
	}
	for _, c := range cases {
		v, ch := TMDBKey(c.key, c.clear)
		if v != c.value || ch != c.change {
			t.Errorf("TMDBKey(%q, %v) = (%q, %v), want (%q, %v)", c.key, c.clear, v, ch, c.value, c.change)
		}
	}
}

func TestEmail(t *testing.T) {
	if em, err := Email("  Admin@Example.COM "); err != nil || em != "admin@example.com" {
		t.Errorf("got %q, %v", em, err)
	}
	if em, err := Email(""); err != nil || em != "" {
		t.Errorf("an empty email means no change: %q, %v", em, err)
	}
	if _, err := Email("not-an-email"); err == nil {
		t.Error("an email needs an @")
	}
}

func TestPassword(t *testing.T) {
	if Password("1234567") == nil {
		t.Error("7 characters is too short")
	}
	if err := Password("12345678"); err != nil {
		t.Error("8 characters is fine")
	}
}

func TestMessagesAreSentcaseWithAPeriod(t *testing.T) {
	// These texts are shown to the admin as they are.
	_, e1 := CinemaName("")
	_, e2 := AccentColor("x")
	_, e3 := HostName("")
	_, e4 := MovieFolders(nil)
	for _, err := range []error{e1, e2, e3, e4, Password("x")} {
		msg := err.Error()
		if !strings.HasSuffix(msg, ".") || msg[0] < 'A' || msg[0] > 'Z' {
			t.Errorf("message %q should start with a capital letter and end with a period", msg)
		}
	}
}
