package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// PickSeek chooses a frame position for a thumbnail: about 12% in (past logos and fades),
// never later than 5 minutes, and the middle of very short clips.
func PickSeek(duration float64) float64 {
	if duration <= 0 {
		return 5
	}
	t := duration * 0.12
	if t < 1 {
		t = duration / 2
	}
	if t > 300 {
		t = 300
	}
	return t
}

// thumbGraph decodes once and makes two pictures. "thumbnail" picks a representative frame so
// black or faded frames are skipped. The poster is a centre crop to 2:3 that also works for
// portrait phone videos.
const thumbGraph = "[0:v]thumbnail=48,split=2[a][b];" +
	"[a]scale='min(1280,iw)':-2[bd];" +
	"[b]crop='min(iw,ih*2/3)':'min(ih,iw*3/2)',scale=400:600[po]"

// ThumbArgs builds the ffmpeg arguments that write a wide backdrop and a 2:3 poster
// from the frame at pos seconds.
func ThumbArgs(src string, pos float64, backdropOut, posterOut string) []string {
	return []string{"-v", "error", "-y",
		"-ss", fmt.Sprintf("%.2f", pos), "-i", src,
		"-filter_complex", thumbGraph,
		"-map", "[bd]", "-frames:v", "1", "-q:v", "3", backdropOut,
		"-map", "[po]", "-frames:v", "1", "-q:v", "3", posterOut}
}

// ExtractThumbs writes a backdrop and a poster JPEG taken at "at" seconds. If that position
// fails (for example it is past the end), it tries again at 1 second.
func ExtractThumbs(bin, src string, at float64, backdropOut, posterOut string) error {
	if bin == "" {
		return errors.New("ffmpeg is not installed")
	}
	run := func(pos float64) error {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bin, ThumbArgs(src, pos, backdropOut, posterOut)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ffmpeg: %w: %s", err, string(out))
		}
		for _, p := range []string{posterOut, backdropOut} {
			if st, err := os.Stat(p); err != nil || st.Size() == 0 {
				return errors.New("ffmpeg produced no frame (position past the end?)")
			}
		}
		return nil
	}
	err := run(at)
	if err != nil && at > 1 {
		err = run(1)
	}
	return err
}
