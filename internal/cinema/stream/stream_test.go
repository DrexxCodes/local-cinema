package stream

import "testing"

func TestIsStreamStart(t *testing.T) {
	cases := map[string]bool{
		"":                true,
		"bytes=0-":        true,
		"bytes=0-1023":    true,
		"bytes=1024-":     false,
		"bytes=500000-":   false,
		"bytes=-500":      false,
		"bytes=0-5,10-20": true, // starts at the first byte
	}
	for h, want := range cases {
		if got := IsStreamStart(h); got != want {
			t.Errorf("IsStreamStart(%q) = %v, want %v", h, got, want)
		}
	}
}

func TestSource(t *testing.T) {
	p, c, used := Source("/m/a.mkv", "mkv", "/c/a.mp4", false)
	if p != "/c/a.mp4" || c != "mp4" || !used {
		t.Errorf("a browser gets the converted copy: %q %q %v", p, c, used)
	}
	p, c, used = Source("/m/a.mkv", "mkv", "/c/a.mp4", true)
	if p != "/m/a.mkv" || c != "mkv" || used {
		t.Errorf("an external player gets the original: %q %q %v", p, c, used)
	}
	p, c, used = Source("/m/a.mkv", "mkv", "", false)
	if p != "/m/a.mkv" || c != "mkv" || used {
		t.Errorf("without a converted copy the original is sent: %q %q %v", p, c, used)
	}
}
