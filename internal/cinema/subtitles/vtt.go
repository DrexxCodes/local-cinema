package subtitles

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	srtTimeRe = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)
	cueTimeRe = regexp.MustCompile(`(?:(\d{2,}):)?(\d{2}):(\d{2})\.(\d{3})`)
)

func lower(s string) string { return strings.ToLower(s) }

// SRTToVTT turns .srt text into WebVTT: a header line, and commas in the times become dots.
func SRTToVTT(txt string) string {
	txt = strings.TrimPrefix(txt, "\ufeff")
	txt = strings.ReplaceAll(txt, "\r\n", "\n")
	return "WEBVTT\n\n" + srtTimeRe.ReplaceAllString(txt, "$1.$2")
}

// ShiftVTT moves every cue by ms milliseconds (negative means earlier). Times never go below zero.
func ShiftVTT(txt string, ms int) string {
	lines := strings.Split(txt, "\n")
	for i, ln := range lines {
		if !strings.Contains(ln, "-->") {
			continue
		}
		lines[i] = cueTimeRe.ReplaceAllStringFunc(ln, func(t string) string {
			m := cueTimeRe.FindStringSubmatch(t)
			h, _ := strconv.Atoi(m[1])
			mi, _ := strconv.Atoi(m[2])
			se, _ := strconv.Atoi(m[3])
			mm, _ := strconv.Atoi(m[4])
			total := ((h*60+mi)*60+se)*1000 + mm + ms
			if total < 0 {
				total = 0
			}
			return fmt.Sprintf("%02d:%02d:%02d.%03d", total/3600000, total/60000%60, total/1000%60, total%1000)
		})
	}
	return strings.Join(lines, "\n")
}

// latin1ToUTF8 reads bytes as Latin-1 (close enough to Windows-1252 for subtitles).
func latin1ToUTF8(b []byte) []byte {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return []byte(string(r))
}

// ToVTT converts an uploaded subtitle file to WebVTT. .vtt and .srt are handled here;
// .ass, .ssa and .sub need ffmpeg (ffmpegBin, "" when it is not installed).
// ext is the lower case extension with its dot. Text that is not UTF-8 is read as Latin-1.
func ToVTT(raw []byte, ext, ffmpegBin string) ([]byte, error) {
	if !utf8.Valid(raw) {
		raw = latin1ToUTF8(raw) // old files are often Windows-1252
	}
	txt := strings.TrimPrefix(string(raw), "\ufeff")
	switch ext {
	case ".vtt":
		if !strings.HasPrefix(strings.TrimSpace(txt), "WEBVTT") {
			return nil, fmt.Errorf("that .vtt file does not start with WEBVTT")
		}
		return []byte(strings.ReplaceAll(txt, "\r\n", "\n")), nil
	case ".srt":
		return []byte(SRTToVTT(txt)), nil
	case ".ass", ".ssa", ".sub":
		if ffmpegBin == "" {
			return nil, fmt.Errorf("ffmpeg is needed to read %s files. Upload .srt or .vtt instead", ext)
		}
		return viaFFmpeg(ffmpegBin, txt, ext)
	}
	return nil, fmt.Errorf("use a .srt, .vtt, .ass or .ssa file")
}

func viaFFmpeg(bin, txt, ext string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "subs")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	in, out := filepath.Join(dir, "in"+ext), filepath.Join(dir, "out.vtt")
	if err := os.WriteFile(in, []byte(txt), 0o644); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, bin, "-v", "error", "-y", "-i", in, "-f", "webvtt", out).Run(); err != nil {
		return nil, fmt.Errorf("could not read that subtitle file")
	}
	return os.ReadFile(out)
}
