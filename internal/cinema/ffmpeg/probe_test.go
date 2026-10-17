package ffmpeg

import "testing"

const sampleProbe = `{
  "streams": [
    {"codec_type": "audio", "codec_name": "ac3"},
    {"codec_type": "video", "codec_name": "h264", "width": 1920, "height": 1080},
    {"codec_type": "video", "codec_name": "mjpeg", "width": 600, "height": 900},
    {"codec_type": "audio", "codec_name": "aac"},
    {"codec_type": "subtitle", "codec_name": "subrip"}
  ],
  "format": {"duration": "7265.432000"}
}`

func TestParseProbe(t *testing.T) {
	got, err := ParseProbe([]byte(sampleProbe))
	if err != nil {
		t.Fatal(err)
	}
	want := Info{Duration: 7265.432, VCodec: "h264", ACodec: "ac3", Resolution: "1920x1080"}
	if got != want {
		t.Errorf("ParseProbe = %+v, want %+v (first video and first audio stream win)", got, want)
	}
}

func TestParseProbeBadInput(t *testing.T) {
	if _, err := ParseProbe([]byte("not json")); err == nil {
		t.Error("invalid JSON should be an error")
	}
	got, err := ParseProbe([]byte(`{"streams": [], "format": {}}`))
	if err != nil || got != (Info{}) {
		t.Errorf("an empty file gives empty info, got %+v, %v", got, err)
	}
	got, _ = ParseProbe([]byte(`{"format": {"duration": "N/A"}}`))
	if got.Duration != 0 {
		t.Errorf("an unreadable duration should be 0, got %v", got.Duration)
	}
}

func TestProbeWithoutBinary(t *testing.T) {
	if got := Probe("", "movie.mp4"); got != (Info{}) {
		t.Errorf("no ffprobe should give empty info, got %+v", got)
	}
}

func TestProbeRunsTheProgram(t *testing.T) {
	// The stand-in prints a version banner, which is not JSON: Probe must not fail loudly.
	bin := useFake(t, "version")
	if got := Probe(bin, "movie.mp4"); got != (Info{}) {
		t.Errorf("unparseable output should give empty info, got %+v", got)
	}
}

func TestParseSubtitleStreams(t *testing.T) {
	raw := `{"streams": [
	  {"index": 2, "codec_name": "subrip", "tags": {"language": "eng", "title": "English SDH"}},
	  {"index": 3, "codec_name": "hdmv_pgs_subtitle", "tags": {"language": "fra"}},
	  {"index": 4, "codec_name": "ass"}
	]}`
	got, err := ParseSubtitleStreams([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	want := []SubStream{
		{Index: 2, Codec: "subrip", Lang: "eng", Title: "English SDH"},
		{Index: 3, Codec: "hdmv_pgs_subtitle", Lang: "fra"},
		{Index: 4, Codec: "ass"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d streams, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stream %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if _, err := ParseSubtitleStreams([]byte("{")); err == nil {
		t.Error("invalid JSON should be an error")
	}
	if got, _ := ParseSubtitleStreams([]byte(`{}`)); len(got) != 0 {
		t.Errorf("no streams expected, got %v", got)
	}
}
