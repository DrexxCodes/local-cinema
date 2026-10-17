package ffmpeg

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// The tests run this test binary itself as a stand-in for ffmpeg: when
// CINEMA_FAKE_FFMPEG is set, the program behaves like ffmpeg instead of running tests.
// This works the same on Windows, Linux and macOS and needs no shell scripts.
func TestMain(m *testing.M) {
	if mode := os.Getenv("CINEMA_FAKE_FFMPEG"); mode != "" {
		os.Exit(fakeFFmpeg(mode, os.Args[1:]))
	}
	os.Exit(m.Run())
}

func argAfter(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func fakeFFmpeg(mode string, args []string) int {
	last := ""
	if len(args) > 0 {
		last = args[len(args)-1]
	}
	switch mode {
	case "version":
		fmt.Println("ffmpeg version 9.9-test Copyright (c) the test authors")
		fmt.Println("built with a fake compiler")
		return 0
	case "convert-ok":
		fmt.Println("out_time_us=5000000")
		fmt.Println("out_time_us=N/A")
		fmt.Println("out_time_us=10000000")
		fmt.Println("progress=end")
		_ = os.WriteFile(last, []byte("fake mp4"), 0o644)
		return 0
	case "convert-fail":
		fmt.Fprint(os.Stderr, strings.Repeat("x", 500)+"THE-END")
		return 1
	case "convert-empty":
		return 0
	case "convert-hang":
		time.Sleep(30 * time.Second)
		return 0
	case "thumbs-ok", "thumbs-fail-late":
		if mode == "thumbs-fail-late" && argAfter(args, "-ss") != "1.00" {
			fmt.Fprint(os.Stderr, "seek past the end")
			return 1
		}
		for i, a := range args {
			if a == "[bd]" && i+5 < len(args) {
				_ = os.WriteFile(args[i+5], []byte("backdrop"), 0o644)
			}
		}
		_ = os.WriteFile(last, []byte("poster"), 0o644)
		return 0
	case "thumbs-fail":
		fmt.Fprint(os.Stderr, "no such file")
		return 1
	case "thumbs-empty":
		return 0
	}
	return 2
}

// useFake makes every ffmpeg run started by the current test behave as the given mode.
func useFake(t *testing.T, mode string) string {
	t.Helper()
	t.Setenv("CINEMA_FAKE_FFMPEG", mode)
	return os.Args[0]
}
