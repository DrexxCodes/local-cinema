package media

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseName(t *testing.T) {
	cases := []struct {
		file  string
		title string
		year  int
	}{
		{"Inception (2010).mp4", "Inception", 2010},
		{"The.Matrix.1999.1080p.BluRay.x264.mkv", "The Matrix", 1999},
		{"Blade_Runner_2049_2017_2160p.mkv", "Blade Runner 2049", 2017},
		{"Alien [1979].avi", "Alien", 1979},
		{"Heat.1995.mp4", "Heat", 1995},
		{"Casablanca.mp4", "Casablanca", 0},
		{"Movie.Name.720p.WEB-DL.x265.mkv", "Movie Name", 0},
		{"  Spaced Out (2001).mp4", "Spaced Out", 2001},
		{"/library/Action/Die Hard (1988).mkv", "Die Hard", 1988},
		{"2012.2009.1080p.mkv", "2012", 2009}, // a year in the title is not the release year
		{"1917 (2019).mp4", "1917", 2019},
		{"1917.mp4", "1917", 0},
		{"Movie 1080p.mkv", "Movie", 0},
		{"Show.S01E02.HDTV.mkv", "Show S01E02 HDTV", 0},
		{".mp4", ".mp4", 0},
		{"Wonder Woman 1984 (2020).mkv", "Wonder Woman 1984", 2020},
		{"Blade.Runner.2049.2017.1080p.BluRay.mkv", "Blade Runner 2049", 2017},
		{"Movie Title 2019 2160p.mkv", "Movie Title", 2019},
	}
	for _, c := range cases {
		title, year := ParseName(c.file)
		if title != c.title || year != c.year {
			t.Errorf("ParseName(%q) = (%q, %d), want (%q, %d)", c.file, title, year, c.title, c.year)
		}
	}
}

func TestIsVideo(t *testing.T) {
	yes := []string{"a.mp4", "a.MKV", "a.Mov", "dir/a.webm", "a.m4v", "a.wmv", "a.flv", "a.mpg", "a.mpeg", "a.ts", "a.m2ts", "a.3gp", "a.ogv", "a.avi"}
	no := []string{"a.srt", "a.txt", "a", "a.mp4.part", "a.jpg", "a.vtt", ".mp4x"}
	for _, f := range yes {
		if !IsVideo(f) {
			t.Errorf("%q should be a video", f)
		}
	}
	for _, f := range no {
		if IsVideo(f) {
			t.Errorf("%q should not be a video", f)
		}
	}
	if !IsVideoExt(".MP4") || IsVideoExt("mp4") {
		t.Error("IsVideoExt wants a leading dot and ignores case")
	}
}

func TestContentType(t *testing.T) {
	if got := ContentType("mp4"); got != "video/mp4" {
		t.Errorf("mp4 = %q", got)
	}
	if got := ContentType("mkv"); got != "video/x-matroska" {
		t.Errorf("mkv = %q", got)
	}
	if got := ContentType("xyz"); got != "application/octet-stream" {
		t.Errorf("unknown = %q", got)
	}
	if got := ContentType(""); got != "application/octet-stream" {
		t.Errorf("empty = %q", got)
	}
}

func write(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindVideos(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	write(t, filepath.Join(a, "one.mp4"), 1)
	write(t, filepath.Join(a, "sub", "deeper", "two.MKV"), 1)
	write(t, filepath.Join(a, "notes.txt"), 1)
	write(t, filepath.Join(a, "one.srt"), 1)
	write(t, filepath.Join(b, "three.avi"), 1)

	got := FindVideos([]string{a, b, filepath.Join(a, "missing")})
	want := []string{
		filepath.Join(a, "one.mp4"),
		filepath.Join(a, "sub", "deeper", "two.MKV"),
		filepath.Join(b, "three.avi"),
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("FindVideos = %v, want %v", got, want)
	}
	if got := FindVideos(nil); len(got) != 0 {
		t.Errorf("no folders, no videos, got %v", got)
	}
}

func TestValidateFolder(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.mp4"), 100)
	write(t, filepath.Join(dir, "sub", "b.mkv"), 250)
	write(t, filepath.Join(dir, "readme.txt"), 999)

	r := ValidateFolder(dir)
	if !r.Valid || r.VideoCount != 2 || r.TotalBytes != 350 || len(r.Warnings) != 0 {
		t.Errorf("unexpected report: %+v", r)
	}
	if r.Warnings == nil {
		t.Error("Warnings must be an empty list, not null, so the UI can map over it")
	}
}

func TestValidateFolderProblems(t *testing.T) {
	empty := t.TempDir()
	r := ValidateFolder(empty)
	if !r.Valid || r.VideoCount != 0 || len(r.Warnings) != 1 {
		t.Errorf("an empty folder is valid but warns: %+v", r)
	}

	r = ValidateFolder(filepath.Join(empty, "nope"))
	if r.Valid || len(r.Warnings) != 1 || r.Warnings[0] == "" {
		t.Errorf("a missing folder is invalid: %+v", r)
	}

	file := filepath.Join(empty, "file.mp4")
	write(t, file, 1)
	if r := ValidateFolder(file); r.Valid {
		t.Error("a file is not a folder")
	}
}

func TestInsideRoots(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	inside := filepath.Join(root, "movies", "a.mp4")
	write(t, inside, 1)
	write(t, filepath.Join(outside, "secret.mp4"), 1)
	write(t, filepath.Join(root, "..hidden", "b.mp4"), 1)

	if !InsideRoots(inside, []string{root}) {
		t.Error("a file inside the root should be accepted")
	}
	if !InsideRoots(filepath.Join(root, "..hidden", "b.mp4"), []string{root}) {
		t.Error("a folder whose name merely starts with two dots is still inside")
	}
	if InsideRoots(filepath.Join(outside, "secret.mp4"), []string{root}) {
		t.Error("a file in another folder must be refused")
	}
	if InsideRoots(filepath.Join(root, "missing.mp4"), []string{root}) {
		t.Error("a file that does not exist must be refused")
	}
	if InsideRoots(inside, nil) {
		t.Error("no library folders means nothing is inside")
	}
	if InsideRoots(inside, []string{filepath.Join(root, "nope")}) {
		t.Error("an unreadable root contributes nothing")
	}
	if !InsideRoots(inside, []string{outside, root}) {
		t.Error("any matching root is enough")
	}
}

func TestInsideRootsRefusesSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.mp4")
	write(t, secret, 1)

	link := filepath.Join(root, "link.mp4")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}
	if InsideRoots(link, []string{root}) {
		t.Error("a symlink that points outside the library must be refused")
	}

	dirLink := filepath.Join(root, "linkdir")
	if err := os.Symlink(outside, dirLink); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}
	if InsideRoots(filepath.Join(dirLink, "secret.mp4"), []string{root}) {
		t.Error("a folder symlink that points outside the library must be refused")
	}
}

func TestInsideRootsAllowsSymlinkedRoot(t *testing.T) {
	real := t.TempDir()
	write(t, filepath.Join(real, "a.mp4"), 1)
	linkRoot := filepath.Join(t.TempDir(), "library")
	if err := os.Symlink(real, linkRoot); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}
	if !InsideRoots(filepath.Join(linkRoot, "a.mp4"), []string{linkRoot}) {
		t.Error("a library folder that is itself a link should work")
	}
}
