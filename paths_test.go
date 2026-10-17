package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDataDirEnvWins(t *testing.T) {
	t.Setenv("CINEMA_DATA_DIR", "/somewhere/data")
	dir, installed := defaultDataDir()
	if dir != "/somewhere/data" || installed {
		t.Errorf("got %q, %v", dir, installed)
	}
}

func TestDefaultDataDirPortable(t *testing.T) {
	t.Setenv("CINEMA_DATA_DIR", "")
	// The test program has no "installed" marker beside it, so data stays next to the program.
	dir, installed := defaultDataDir()
	if dir != "" || installed {
		t.Errorf("a portable copy returns (\"\", false), got %q, %v", dir, installed)
	}
}

func TestSetupFileLoggingWritesAndRotates(t *testing.T) {
	oldSlog, oldLog := slog.Default(), log.Writer()
	t.Cleanup(func() { slog.SetDefault(oldSlog); log.SetOutput(oldLog) })

	root := t.TempDir()
	dataDir := filepath.Join(root, "data")
	logs := filepath.Join(root, "logs")
	if err := os.MkdirAll(logs, 0o755); err != nil {
		t.Fatal(err)
	}
	big := filepath.Join(logs, "local-cinema.log")
	if err := os.WriteFile(big, make([]byte, 5<<20+1), 0o644); err != nil {
		t.Fatal(err)
	}

	setupFileLogging(dataDir)
	slog.Info("hello from the test")

	if _, err := os.Stat(big + ".old"); err != nil {
		t.Error("a log over 5 MB must be moved aside")
	}
	b, err := os.ReadFile(big)
	if err != nil || len(b) == 0 || len(b) > 1<<20 {
		t.Errorf("a fresh log should hold only the new line, got %d bytes, err %v", len(b), err)
	}
}
