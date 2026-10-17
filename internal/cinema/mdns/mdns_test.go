package mdns

import (
	"slices"
	"testing"
)

func TestHostName(t *testing.T) {
	if HostName("") != "cinema" {
		t.Error("an empty setting falls back to cinema")
	}
	if HostName("drexx-movies") != "drexx-movies" {
		t.Error("a set name is used as it is")
	}
}

func TestTXT(t *testing.T) {
	got := TXT("Drexx Cinema", true, 42)
	want := []string{"name=Drexx Cinema", "version=1.0", "setup=true", "movies=42"}
	if !slices.Equal(got, want) {
		t.Errorf("TXT = %v, want %v", got, want)
	}
	if got := TXT("x", false, 0); got[2] != "setup=false" || got[3] != "movies=0" {
		t.Errorf("TXT = %v", got)
	}
}
