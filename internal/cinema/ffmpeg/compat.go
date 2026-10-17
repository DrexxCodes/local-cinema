package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"local-cinema/internal/cinema/util"
)

// Compatibility statuses reported to the admin page.
const (
	StatusOK        = "ok"        // plays on every device as it is
	StatusConvert   = "convert"   // some devices (iPhones, older browsers) may not play it
	StatusConverted = "converted" // a universal MP4 copy exists and is what guests get
	StatusUnknown   = "unknown"   // no codec info (ffprobe missing), cannot tell
)

// Universal reports whether a file is H.264 video (and AAC or MP3 audio, or none) in an
// MP4 style container: the combination every browser and phone plays.
func Universal(container, vcodec, acodec string) bool {
	switch container {
	case "mp4", "m4v", "mov":
	default:
		return false
	}
	if vcodec != "h264" {
		return false
	}
	switch acodec {
	case "aac", "mp3", "":
		return true
	}
	return false
}

// Classify gives the compatibility status of a movie. converted says a converted copy exists.
func Classify(container, vcodec, acodec string, converted bool) string {
	if converted {
		return StatusConverted
	}
	if vcodec == "" {
		return StatusUnknown
	}
	if Universal(container, vcodec, acodec) {
		return StatusOK
	}
	return StatusConvert
}

// NeedsConversion is true for the movies "Make compatible" works on. Movies with unknown
// codecs are included only when the admin picked them by hand (explicit is true).
func NeedsConversion(status string, explicit bool) bool {
	return status == StatusConvert || (explicit && status == StatusUnknown)
}

// ConvertArgs builds the ffmpeg arguments that write a universal MP4 to tmp. H.264 video is
// copied untouched (fast, no quality loss). Anything else is re-encoded, which is slow.
// AAC audio is copied, other audio becomes stereo AAC.
func ConvertArgs(src, tmp, vcodec, acodec string) []string {
	args := []string{"-hide_banner", "-v", "error", "-nostdin", "-y", "-i", src,
		"-map", "0:v:0", "-map", "0:a:0?", "-sn", "-dn"}
	if vcodec == "h264" {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "21", "-pix_fmt", "yuv420p")
	}
	if acodec == "aac" {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "aac", "-b:a", "192k", "-ac", "2")
	}
	return append(args, "-movflags", "+faststart", "-progress", "pipe:1", "-nostats", tmp)
}

// ParseProgress reads one line of "-progress pipe:1" output. It returns the percent done
// (0 to 100) when the line reports the output time and the total length is known.
func ParseProgress(line string, totalSeconds float64) (int, bool) {
	if totalSeconds <= 0 || !strings.HasPrefix(line, "out_time_us=") {
		return 0, false
	}
	us, err := strconv.ParseFloat(strings.TrimPrefix(line, "out_time_us="), 64)
	if err != nil {
		return 0, false
	}
	return util.ClampInt(int(us/1e6/totalSeconds*100), 0, 100), true
}

// ConvertOptions describes one conversion.
type ConvertOptions struct {
	Src, Out       string  // source movie and the finished MP4
	VCodec, ACodec string  // codecs of the source, from ffprobe
	Duration       float64 // seconds, for the progress percent (0 when unknown)
	OnPercent      func(int)
}

// Convert writes a universal MP4 copy of Src to Out. The work happens in a ".part.mp4" file
// that is renamed when ffmpeg succeeds, so a crash never leaves a half written Out.
// A cancelled context returns context.Canceled.
func Convert(ctx context.Context, bin string, o ConvertOptions) error {
	if bin == "" {
		return errors.New("ffmpeg is not installed")
	}
	tmp := strings.TrimSuffix(o.Out, ".mp4") + ".part.mp4"
	defer os.Remove(tmp)

	cmd := exec.CommandContext(ctx, bin, ConvertArgs(o.Src, tmp, o.VCodec, o.ACodec)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		if pct, ok := ParseProgress(sc.Text(), o.Duration); ok && o.OnPercent != nil {
			o.OnPercent(pct)
		}
	}
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return context.Canceled
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 300 {
			msg = msg[len(msg)-300:]
		}
		return fmt.Errorf("ffmpeg failed: %s", msg)
	}
	if !util.FileExists(tmp) {
		return errors.New("ffmpeg produced no file")
	}
	return os.Rename(tmp, o.Out)
}
