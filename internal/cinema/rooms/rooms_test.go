package rooms

import (
	"regexp"
	"testing"
)

func TestUsername(t *testing.T) {
	cases := map[string]string{
		"room204@cinema.local": "room204",
		"a@b@c":                "a",
		"noat":                 "noat",
		"":                     "",
	}
	for in, want := range cases {
		if got := Username(in); got != want {
			t.Errorf("Username(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRoomEmail(t *testing.T) {
	if got := RoomEmail("room", 204); got != "room204@cinema.local" {
		t.Errorf("got %q", got)
	}
	if got := RoomEmail("suite-", 7); got != "suite-7@cinema.local" {
		t.Errorf("got %q", got)
	}
}

func TestValidRange(t *testing.T) {
	cases := []struct {
		from, to int
		want     bool
	}{
		{101, 120, true}, {5, 5, true}, {0, 1000, true},
		{120, 101, false}, {0, 1001, false}, {1, 2000, false},
	}
	for _, c := range cases {
		if got := ValidRange(c.from, c.to); got != c.want {
			t.Errorf("ValidRange(%d, %d) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestBulkDefaults(t *testing.T) {
	cases := []struct {
		prefix     string
		max        int
		wantPrefix string
		wantMax    int
	}{
		{"", 0, "profile", 2}, {"suite", 3, "suite", 3}, {"", -1, "profile", 2}, {"x", 1, "x", 1},
	}
	for _, c := range cases {
		p, m := BulkDefaults(c.prefix, c.max)
		if p != c.wantPrefix || m != c.wantMax {
			t.Errorf("BulkDefaults(%q, %d) = (%q, %d), want (%q, %d)", c.prefix, c.max, p, m, c.wantPrefix, c.wantMax)
		}
	}
}

func TestNewPassword(t *testing.T) {
	hex8 := regexp.MustCompile(`^[0-9a-f]{8}$`)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		p := NewPassword()
		if !hex8.MatchString(p) {
			t.Fatalf("password %q should be 8 hex characters", p)
		}
		seen[p] = true
	}
	if len(seen) < 99 {
		t.Errorf("passwords should be random, only %d different out of 100", len(seen))
	}
}

func TestClampDevices(t *testing.T) {
	for in, want := range map[int]int{-1: 1, 0: 1, 1: 1, 7: 7, 50: 50, 51: 50} {
		if got := ClampDevices(in); got != want {
			t.Errorf("ClampDevices(%d) = %d, want %d", in, got, want)
		}
	}
}
