package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// Info is what ffprobe tells us about a video file.
type Info struct {
	Duration   float64 // seconds
	VCodec     string  // first video stream, for example h264
	ACodec     string  // first audio stream, for example aac
	Resolution string  // for example 1920x1080
}

// ParseProbe reads the JSON printed by
// "ffprobe -print_format json -show_format -show_streams".
func ParseProbe(out []byte) (Info, error) {
	var d struct {
		Format  struct{ Duration string } `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	var p Info
	if err := json.Unmarshal(out, &d); err != nil {
		return p, err
	}
	p.Duration, _ = strconv.ParseFloat(d.Format.Duration, 64)
	for _, st := range d.Streams {
		if st.CodecType == "video" && p.VCodec == "" {
			p.VCodec, p.Resolution = st.CodecName, fmt.Sprintf("%dx%d", st.Width, st.Height)
		}
		if st.CodecType == "audio" && p.ACodec == "" {
			p.ACodec = st.CodecName
		}
	}
	return p, nil
}

// Probe runs ffprobe on a file. Any failure (no ffprobe, unreadable file) gives an
// empty Info, because a movie without codec details is still listed.
func Probe(bin, path string) Info {
	if bin == "" {
		return Info{}
	}
	out, err := exec.Command(bin, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path).Output()
	if err != nil {
		return Info{}
	}
	info, err := ParseProbe(out)
	if err != nil {
		return Info{}
	}
	return info
}

// SubStream is one subtitle track inside a video file.
type SubStream struct {
	Index int
	Codec string
	Lang  string // three letter code from the file, for example eng
	Title string
}

// ParseSubtitleStreams reads the JSON printed by
// "ffprobe -select_streams s -show_entries stream=index,codec_name:stream_tags=language,title".
func ParseSubtitleStreams(out []byte) ([]SubStream, error) {
	var d struct {
		Streams []struct {
			Index     int               `json:"index"`
			CodecName string            `json:"codec_name"`
			Tags      map[string]string `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &d); err != nil {
		return nil, err
	}
	var res []SubStream
	for _, s := range d.Streams {
		res = append(res, SubStream{Index: s.Index, Codec: s.CodecName, Lang: s.Tags["language"], Title: s.Tags["title"]})
	}
	return res, nil
}

// ProbeSubtitles lists the subtitle tracks inside a video file.
func ProbeSubtitles(bin, path string) ([]SubStream, error) {
	out, err := exec.Command(bin, "-v", "quiet", "-print_format", "json", "-select_streams", "s",
		"-show_entries", "stream=index,codec_name:stream_tags=language,title", path).Output()
	if err != nil {
		return nil, err
	}
	return ParseSubtitleStreams(out)
}
