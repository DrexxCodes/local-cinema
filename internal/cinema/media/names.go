// Package media holds the pure rules about movie files: which files count as videos, how a
// file name becomes a title and year, which folders are safe to read from.
package media

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	yearRe = regexp.MustCompile(`[\(\[\.\s_-]((?:19|20)\d{2})(?:[\)\]\.\s_-]|$)`)
	junkRe = regexp.MustCompile(`(?i)[\.\s_\-\(\[](2160p|1080p|720p|480p|4k|bluray|brrip|bdrip|webrip|web-dl|hdrip|dvdrip|x264|x265|h264|h265|hevc|hdr|remux)\b`)
	sepRe  = regexp.MustCompile(`[._]+`)
)

// lastYear finds the release year in a name: the last year-like number that follows a
// separator. Taking the last one (not the first) keeps a number that belongs to the title,
// as in "Blade_Runner_2049_2017", out of the year. cut is where the title ends (the
// separator before the year). ok is false when there is no year.
func lastYear(base string) (year, cut int, ok bool) {
	for start := 0; start < len(base); {
		m := yearRe.FindStringSubmatchIndex(base[start:])
		if m == nil {
			break
		}
		// Resume at the end of the digits, so the separator that closes this year can open the next.
		if m[0]+start > 0 {
			year, _ = strconv.Atoi(base[start+m[2] : start+m[3]])
			cut, ok = start+m[0], true
		}
		start += m[3]
	}
	return year, cut, ok
}

// ParseName turns a file name into a title and a year (0 when unknown).
// "The.Matrix.1999.1080p.BluRay.x264.mkv" gives ("The Matrix", 1999).
func ParseName(file string) (string, int) {
	base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	cut, year := len(base), 0
	if y, c, ok := lastYear(base); ok {
		year, cut = y, c
	}
	if m := junkRe.FindStringIndex(base); m != nil && m[0] < cut && m[0] > 0 {
		cut = m[0]
	}
	title := strings.TrimSpace(sepRe.ReplaceAllString(base[:cut], " "))
	if title == "" {
		title = filepath.Base(file) // a file called just ".mp4" still gets a name
	}
	return title, year
}
