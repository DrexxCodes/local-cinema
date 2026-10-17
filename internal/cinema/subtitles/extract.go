package subtitles

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/ffmpeg"
	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/util"
)

// extractJob is one subtitle track to read out of a video.
type extractJob struct {
	stream           int    // stream index inside the file
	key, lang, label string // how it is stored
	tmp              string // where ffmpeg writes it first
}

// planExtraction decides which tracks of a video can be extracted. Picture based tracks
// (DVD, Blu-ray) cannot become text and are only counted in skipped. dir is where the
// temporary files go.
func planExtraction(streams []ffmpeg.SubStream, dir string) (jobs []extractJob, skipped int) {
	for _, st := range streams {
		if !textCodecs[st.Codec] {
			skipped++
			continue
		}
		lang, name := languageFor(st.Lang)
		label := name
		if t := strings.TrimSpace(st.Title); t != "" && len(t) <= 40 {
			label = t
		}
		if label == "" {
			label = "Track " + strconv.Itoa(len(jobs)+1)
		}
		key := "e" + strconv.Itoa(st.Index)
		jobs = append(jobs, extractJob{stream: st.Index, key: key, lang: lang, label: label, tmp: filepath.Join(dir, key+".part.vtt")})
	}
	return jobs, skipped
}

// extractArgs builds the ffmpeg arguments that write every planned track as WebVTT.
func extractArgs(src string, jobs []extractJob) []string {
	args := []string{"-v", "error", "-nostdin", "-y", "-i", src}
	for _, j := range jobs {
		args = append(args, "-map", "0:"+strconv.Itoa(j.stream), "-f", "webvtt", j.tmp)
	}
	return args
}

// extractSummary is the message shown after an extraction.
func extractSummary(done, skipped int) string {
	plural := "s"
	if done == 1 {
		plural = ""
	}
	msg := fmt.Sprintf("Extracted %d subtitle track%s.", done, plural)
	if skipped > 0 {
		msg += fmt.Sprintf(" %d picture based track(s) were skipped.", skipped)
	}
	return msg
}

// handleAdminExtract starts reading the text subtitle tracks out of the video file.
func (s *Service) handleAdminExtract(e *core.RequestEvent) error {
	m, err := s.movieFor(e)
	if m == nil {
		return err
	}
	if s.st.Tools.FFmpeg() == "" {
		return httpx.Err(e, 400, "no_ffmpeg", "ffmpeg is not installed on the host.")
	}
	if s.Extract.Running() {
		return httpx.Err(e, 409, "busy", "Another extraction is running.")
	}
	go s.runExtract(m)
	return e.JSON(202, map[string]any{"ok": true})
}

func (s *Service) runExtract(m *core.Record) {
	if !s.Extract.Begin() {
		return
	}
	defer s.Extract.End()
	s.setNote("")
	src := m.GetString("file_path")
	if !s.st.InsideLibrary(src) {
		s.setNote("The file is outside the library folders.")
		return
	}
	probeBin := s.st.Tools.FFprobe()
	if probeBin == "" {
		s.setNote("ffprobe is not installed, so the tracks inside the video can't be read.")
		return
	}
	streams, err := ffmpeg.ProbeSubtitles(probeBin, src)
	if err != nil {
		s.setNote("Could not read the file's tracks (is ffprobe installed?).")
		return
	}
	if len(streams) == 0 {
		s.setNote("This video has no subtitle tracks inside it.")
		return
	}

	dir := s.st.Cache.SubsDir(m.Id)
	_ = os.MkdirAll(dir, 0o755)
	jobs, skipped := planExtraction(streams, dir)
	if len(jobs) == 0 {
		s.setNote(fmt.Sprintf("The %d subtitle track(s) in this video are picture based and cannot be converted to text.", skipped))
		return
	}
	s.Extract.Set(len(jobs), 0, m.GetString("title"))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, s.st.Tools.FFmpeg(), extractArgs(src, jobs)...)
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.loadIndex(m.Id)
	done := 0
	for _, j := range jobs {
		if !util.FileExists(j.tmp) {
			continue
		}
		if err := os.Rename(j.tmp, s.trackFile(m.Id, j.key)); err != nil {
			continue
		}
		idx.upsert(Entry{Key: j.key, Lang: j.lang, Label: j.label, Source: SourceEmbedded})
		done++
	}
	for _, j := range jobs {
		_ = os.Remove(j.tmp)
	}
	_ = s.saveIndex(m.Id, idx)
	msg := extractSummary(done, skipped)
	if runErr != nil && done == 0 {
		msg = "Extraction failed: " + strings.TrimSpace(stderr.String())
	}
	s.setNote(msg)
}
