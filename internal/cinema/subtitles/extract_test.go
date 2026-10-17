package subtitles

import (
	"path/filepath"
	"slices"
	"testing"

	"local-cinema/internal/cinema/ffmpeg"
)

func TestLanguageFor(t *testing.T) {
	cases := []struct{ tag, code, name string }{
		{"eng", "en", "English"},
		{"ENG", "en", "English"},
		{"fre", "fr", "French"},
		{"fra", "fr", "French"},
		{"yor", "yo", "Yoruba"},
		{"zzz", "und", ""},
		{"", "und", ""},
	}
	for _, c := range cases {
		if code, name := languageFor(c.tag); code != c.code || name != c.name {
			t.Errorf("languageFor(%q) = (%q, %q), want (%q, %q)", c.tag, code, name, c.code, c.name)
		}
	}
}

func TestPlanExtraction(t *testing.T) {
	streams := []ffmpeg.SubStream{
		{Index: 2, Codec: "subrip", Lang: "eng"},
		{Index: 3, Codec: "hdmv_pgs_subtitle", Lang: "eng"}, // picture based
		{Index: 4, Codec: "ass", Lang: "fre", Title: "  Forced  "},
		{Index: 5, Codec: "mov_text"},                  // unknown language, no title
		{Index: 6, Codec: "dvd_subtitle", Lang: "spa"}, // picture based
		{Index: 7, Codec: "webvtt", Lang: "xxx", Title: "A very long title that is over forty characters"},
	}
	jobs, skipped := planExtraction(streams, "dir")
	if skipped != 2 {
		t.Errorf("skipped = %d, want 2", skipped)
	}
	type view struct{ key, lang, label string }
	var got []view
	for _, j := range jobs {
		got = append(got, view{j.key, j.lang, j.label})
	}
	want := []view{
		{"e2", "en", "English"},
		{"e4", "fr", "Forced"}, // the track's own title wins
		{"e5", "und", "Track 3"},
		{"e7", "und", "Track 4"}, // a title over 40 characters is ignored, and the language is unknown
	}
	if !slices.Equal(got, want) {
		t.Errorf("jobs = %+v, want %+v", got, want)
	}
	for _, j := range jobs {
		if want := filepath.Join("dir", j.key+".part.vtt"); j.tmp != want {
			t.Errorf("tmp for %s = %q, want %q", j.key, j.tmp, want)
		}
	}
}

func TestPlanExtractionNothingUsable(t *testing.T) {
	jobs, skipped := planExtraction([]ffmpeg.SubStream{{Index: 1, Codec: "hdmv_pgs_subtitle"}}, "d")
	if len(jobs) != 0 || skipped != 1 {
		t.Errorf("jobs=%v skipped=%d", jobs, skipped)
	}
	jobs, skipped = planExtraction(nil, "d")
	if len(jobs) != 0 || skipped != 0 {
		t.Errorf("jobs=%v skipped=%d", jobs, skipped)
	}
}

func TestExtractArgs(t *testing.T) {
	jobs := []extractJob{{stream: 2, tmp: "a.part.vtt"}, {stream: 5, tmp: "b.part.vtt"}}
	got := extractArgs("movie.mkv", jobs)
	want := []string{"-v", "error", "-nostdin", "-y", "-i", "movie.mkv",
		"-map", "0:2", "-f", "webvtt", "a.part.vtt",
		"-map", "0:5", "-f", "webvtt", "b.part.vtt"}
	if !slices.Equal(got, want) {
		t.Errorf("args = %v\nwant   %v", got, want)
	}
}

func TestExtractSummary(t *testing.T) {
	cases := []struct {
		done, skipped int
		want          string
	}{
		{1, 0, "Extracted 1 subtitle track."},
		{3, 0, "Extracted 3 subtitle tracks."},
		{0, 0, "Extracted 0 subtitle tracks."},
		{2, 1, "Extracted 2 subtitle tracks. 1 picture based track(s) were skipped."},
	}
	for _, c := range cases {
		if got := extractSummary(c.done, c.skipped); got != c.want {
			t.Errorf("extractSummary(%d, %d) = %q, want %q", c.done, c.skipped, got, c.want)
		}
	}
}
