package subtitles

import (
	"strings"
	"testing"
)

func TestSRTToVTT(t *testing.T) {
	got := SRTToVTT("1\n00:00:01,000 --> 00:00:02,000\nHi\n")
	want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:02.000\nHi\n"
	if got != want {
		t.Errorf("SRTToVTT = %q, want %q", got, want)
	}
}

func TestSRTToVTTCleansUp(t *testing.T) {
	got := SRTToVTT("\xEF\xBB\xBF1\r\n00:00:01,500 --> 00:00:02,250\r\nHello, world\r\n")
	if strings.Contains(got, "\r") || strings.Contains(got, "\xEF\xBB\xBF") {
		t.Errorf("a BOM and Windows line ends should be removed, got %q", got)
	}
	if !strings.Contains(got, "00:00:01.500 --> 00:00:02.250") {
		t.Errorf("times should use dots, got %q", got)
	}
	if !strings.Contains(got, "Hello, world") {
		t.Errorf("commas in the text must be kept, got %q", got)
	}
}

func TestShiftVTT(t *testing.T) {
	in := "WEBVTT\n\n00:00:01.000 --> 00:00:02.500\nHello\n"
	want := "WEBVTT\n\n00:00:01.500 --> 00:00:03.000\nHello\n"
	if got := ShiftVTT(in, 500); got != want {
		t.Errorf("ShiftVTT = %q, want %q", got, want)
	}
}

func TestShiftVTTClampsAtZero(t *testing.T) {
	if got := ShiftVTT("00:00:00.200 --> 00:00:01.000", -500); got != "00:00:00.000 --> 00:00:00.500" {
		t.Errorf("negative shift clamps at zero, got %q", got)
	}
}

func TestShiftVTTCarriesIntoHours(t *testing.T) {
	got := ShiftVTT("00:59:59.900 --> 01:00:00.000", 200)
	if got != "01:00:00.100 --> 01:00:00.200" {
		t.Errorf("got %q", got)
	}
}

func TestShiftVTTShortTimes(t *testing.T) {
	// WebVTT allows mm:ss.mmm without hours. They are written back as hh:mm:ss.mmm.
	got := ShiftVTT("WEBVTT\n\n01:02.000 --> 01:03.000\nline\n", 1000)
	if !strings.Contains(got, "00:01:03.000 --> 00:01:04.000") {
		t.Errorf("got %q", got)
	}
}

func TestShiftVTTLeavesNonCueLinesAlone(t *testing.T) {
	in := "WEBVTT\n\nNOTE this is 00:00:05.000 in a note\n\n00:00:05.000 --> 00:00:06.000\nline\n"
	got := ShiftVTT(in, 1000)
	if !strings.Contains(got, "NOTE this is 00:00:05.000 in a note") {
		t.Errorf("only cue timing lines may change, got %q", got)
	}
}

func TestToVTT(t *testing.T) {
	t.Run("vtt is kept", func(t *testing.T) {
		out, err := ToVTT([]byte("WEBVTT\r\n\r\n00:00:01.000 --> 00:00:02.000\r\nHi\r\n"), ".vtt", "")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "\r") || !strings.HasPrefix(string(out), "WEBVTT") {
			t.Errorf("got %q", out)
		}
	})
	t.Run("vtt without header is refused", func(t *testing.T) {
		if _, err := ToVTT([]byte("not a subtitle file"), ".vtt", ""); err == nil {
			t.Error("a .vtt file must start with WEBVTT")
		}
	})
	t.Run("vtt with a BOM", func(t *testing.T) {
		if _, err := ToVTT([]byte("\xEF\xBB\xBFWEBVTT\n\n"), ".vtt", ""); err != nil {
			t.Errorf("a BOM before WEBVTT is normal: %v", err)
		}
	})
	t.Run("srt is converted", func(t *testing.T) {
		out, err := ToVTT([]byte("1\n00:00:01,000 --> 00:00:02,000\nHi\n"), ".srt", "")
		if err != nil || !strings.HasPrefix(string(out), "WEBVTT\n\n") {
			t.Errorf("got %q, %v", out, err)
		}
	})
	t.Run("windows-1252 text becomes UTF-8", func(t *testing.T) {
		// "caf\xe9" is "café" in Windows-1252 and is not valid UTF-8.
		out, err := ToVTT([]byte("1\n00:00:01,000 --> 00:00:02,000\ncaf\xe9\n"), ".srt", "")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), "café") {
			t.Errorf("accents should survive, got %q", out)
		}
	})
	t.Run("ass needs ffmpeg", func(t *testing.T) {
		for _, ext := range []string{".ass", ".ssa", ".sub"} {
			_, err := ToVTT([]byte("[Script Info]"), ext, "")
			if err == nil || !strings.Contains(err.Error(), "ffmpeg is needed") {
				t.Errorf("%s without ffmpeg: %v", ext, err)
			}
		}
	})
	t.Run("other formats are refused", func(t *testing.T) {
		for _, ext := range []string{".txt", ".docx", "", ".SRT"} { // the caller lower-cases the extension
			if _, err := ToVTT([]byte("x"), ext, ""); err == nil {
				t.Errorf("extension %q should be refused", ext)
			}
		}
	})
}
