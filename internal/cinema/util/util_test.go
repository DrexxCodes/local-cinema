package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClampInt(t *testing.T) {
	cases := []struct{ v, lo, hi, want int }{
		{5, 1, 10, 5}, {0, 1, 10, 1}, {11, 1, 10, 10}, {1, 1, 10, 1}, {10, 1, 10, 10}, {-3, -5, -1, -3},
	}
	for _, c := range cases {
		if got := ClampInt(c.v, c.lo, c.hi); got != c.want {
			t.Errorf("ClampInt(%d, %d, %d) = %d, want %d", c.v, c.lo, c.hi, got, c.want)
		}
	}
}

func TestClampFloat(t *testing.T) {
	cases := []struct{ v, lo, hi, want float64 }{
		{5.5, 0, 10, 5.5}, {-0.1, 0, 10, 0}, {10.1, 0, 10, 10}, {0, 0, 10, 0},
	}
	for _, c := range cases {
		if got := ClampFloat(c.v, c.lo, c.hi); got != c.want {
			t.Errorf("ClampFloat(%v, %v, %v) = %v, want %v", c.v, c.lo, c.hi, got, c.want)
		}
	}
}

func TestQueryInt(t *testing.T) {
	cases := []struct {
		raw               string
		def, lo, hi, want int
	}{
		{"50", 20, 1, 100, 50},
		{"", 20, 1, 100, 20},
		{"abc", 20, 1, 100, 20},
		{"1000", 20, 1, 100, 100},
		{"0", 20, 1, 100, 1},
		{"-5", 20, 0, 100, 0},
		{"1.5", 20, 1, 100, 20},
	}
	for _, c := range cases {
		if got := QueryInt(c.raw, c.def, c.lo, c.hi); got != c.want {
			t.Errorf("QueryInt(%q, %d, %d, %d) = %d, want %d", c.raw, c.def, c.lo, c.hi, got, c.want)
		}
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	full := filepath.Join(dir, "full.bin")
	empty := filepath.Join(dir, "empty.bin")
	if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(full) {
		t.Error("a file with content should exist")
	}
	if FileExists(empty) {
		t.Error("an empty file should count as missing")
	}
	if FileExists(dir) {
		t.Error("a directory is not a file")
	}
	if FileExists(filepath.Join(dir, "nope")) {
		t.Error("a missing path should not exist")
	}
}
