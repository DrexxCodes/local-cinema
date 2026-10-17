package subtitles

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	return New(&store.Store{Cache: cache.New(t.TempDir())})
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindSidecars(t *testing.T) {
	dir := t.TempDir()
	movie := filepath.Join(dir, "Inception (2010).mp4")
	for _, f := range []string{
		"Inception (2010).mp4",
		"Inception (2010).srt",    // no language
		"Inception (2010).en.srt", // English
		"Inception (2010).fr.vtt", // French, WebVTT
		"Inception (2010).en.ASS", // not supported next to the movie
		"Other Movie.srt",         // belongs to another movie
		"Inception (2010) extras.srt",
		"Inception (2010).bad lang!.srt", // not a valid language code
		"notes.txt",
	} {
		write(t, filepath.Join(dir, f), "x")
	}
	if err := os.Mkdir(filepath.Join(dir, "Inception (2010).de.srt"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := FindSidecars(movie)
	byKey := map[string]Track{}
	for _, tr := range got {
		byKey[tr.Key] = tr
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if !slices.Equal(keys, []string{"en", "fr", "und"}) {
		t.Fatalf("keys = %v, want [en fr und]", keys)
	}
	if byKey["und"].Label != "Subtitles" || byKey["en"].Label != "en" {
		t.Errorf("labels: und=%q en=%q", byKey["und"].Label, byKey["en"].Label)
	}
	for _, tr := range got {
		if tr.Source != SourceFile {
			t.Errorf("%s: source = %q, want file", tr.Key, tr.Source)
		}
		if filepath.Dir(tr.path) != dir {
			t.Errorf("%s: path %q is not beside the movie", tr.Key, tr.path)
		}
	}
}

func TestFindSidecarsMissingFolder(t *testing.T) {
	if got := FindSidecars(filepath.Join(t.TempDir(), "nope", "a.mp4")); len(got) != 0 {
		t.Errorf("a missing folder has no subtitles, got %v", got)
	}
}

func TestValidLang(t *testing.T) {
	for _, ok := range []string{"en", "pt-BR", "und", "zh-Hant", "x1"} {
		if !ValidLang(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "en us", "../x", "a/b", strings.Repeat("a", 17), "é"} {
		if ValidLang(bad) {
			t.Errorf("%q should be refused", bad)
		}
	}
}

func TestCleanLabel(t *testing.T) {
	if l, ok := cleanLabel("  English  "); !ok || l != "English" {
		t.Errorf("got %q, %v", l, ok)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("x", 41)} {
		if _, ok := cleanLabel(bad); ok {
			t.Errorf("%q should be refused", bad)
		}
	}
	if _, ok := cleanLabel(strings.Repeat("x", 40)); !ok {
		t.Error("40 characters is allowed")
	}
}

func TestRandKey(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		k := randKey("u")
		if len(k) != 7 || k[0] != 'u' {
			t.Fatalf("key %q should be u + 6 hex characters", k)
		}
		for _, c := range k[1:] {
			if !strings.ContainsRune("0123456789abcdef", c) {
				t.Fatalf("key %q has a non hex character", k)
			}
		}
		seen[k] = true
	}
	if len(seen) < 40 {
		t.Errorf("keys should be mostly unique, got %d different out of 50", len(seen))
	}
}

func TestIndexOperations(t *testing.T) {
	idx := Index{Tracks: []Entry{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}, {Key: "c", Label: "C"}}}

	if idx.position("b") != 1 || idx.position("zzz") != -1 {
		t.Error("position finds a stored key, -1 otherwise")
	}

	idx.setDefault("b", true)
	if idx.Default != "b" {
		t.Error("default should be b")
	}
	idx.setDefault("a", false)
	if idx.Default != "b" {
		t.Error("turning off another track must not clear the default")
	}
	idx.setDefault("b", false)
	if idx.Default != "" {
		t.Error("turning off the default clears it")
	}

	idx.Default = "b"
	if !idx.remove("b") || idx.Default != "" {
		t.Error("removing the default track clears the default")
	}
	if idx.remove("b") {
		t.Error("removing a missing track reports false")
	}
	keys := []string{idx.Tracks[0].Key, idx.Tracks[1].Key}
	if len(idx.Tracks) != 2 || !slices.Equal(keys, []string{"a", "c"}) {
		t.Errorf("tracks after remove = %v", idx.Tracks)
	}

	idx.upsert(Entry{Key: "c", Lang: "fr", Label: "French", Source: SourceEmbedded})
	idx.upsert(Entry{Key: "d", Lang: "de", Label: "German", Source: SourceEmbedded})
	if len(idx.Tracks) != 3 || idx.Tracks[1].Lang != "fr" || idx.Tracks[1].Label != "French" || idx.Tracks[2].Key != "d" {
		t.Errorf("upsert should update c and add d, got %+v", idx.Tracks)
	}
}

func TestIndexRoundTrip(t *testing.T) {
	s := testService(t)
	if got := s.loadIndex("movie1"); got.Default != "" || len(got.Tracks) != 0 {
		t.Errorf("a movie without an index starts empty, got %+v", got)
	}
	want := Index{Default: "u1", Tracks: []Entry{{Key: "u1", Lang: "en", Label: "English", Source: SourceUpload}}}
	if err := s.saveIndex("movie1", want); err != nil {
		t.Fatal(err)
	}
	got := s.loadIndex("movie1")
	if got.Default != want.Default || len(got.Tracks) != 1 || got.Tracks[0] != want.Tracks[0] {
		t.Errorf("loaded %+v, want %+v", got, want)
	}
	if err := os.MkdirAll(s.st.Cache.SubsDir("broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(s.st.Cache.SubsDir("broken"), "index.json"), "{not json")
	if got := s.loadIndex("broken"); len(got.Tracks) != 0 {
		t.Errorf("a damaged index reads as empty, got %+v", got)
	}
}

func TestTrackFileStaysInsideTheMovieFolder(t *testing.T) {
	s := testService(t)
	p := s.trackFile("movie1", "../../escape")
	if filepath.Dir(p) != s.st.Cache.SubsDir("movie1") {
		t.Errorf("a key must not climb out of the subtitle folder: %q", p)
	}
}

func TestAddUploadAndUpdate(t *testing.T) {
	s := testService(t)
	if err := s.addUpload("m1", "en", "English", []byte("WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nHi\n")); err != nil {
		t.Fatal(err)
	}
	idx := s.loadIndex("m1")
	if len(idx.Tracks) != 1 {
		t.Fatalf("one track expected, got %+v", idx.Tracks)
	}
	e := idx.Tracks[0]
	if e.Source != SourceUpload || e.Lang != "en" || e.Label != "English" || !strings.HasPrefix(e.Key, "u") {
		t.Errorf("entry = %+v", e)
	}
	if _, err := os.Stat(s.trackFile("m1", e.Key)); err != nil {
		t.Errorf("the track file should exist: %v", err)
	}

	// Rename, make default and shift in one go.
	label, def, shift := "Français", true, 1500
	if re := s.update("m1", e.Key, updateBody{Label: &label, Default: &def, ShiftMs: &shift}); re != nil {
		t.Fatalf("update failed: %v", re)
	}
	idx = s.loadIndex("m1")
	if idx.Tracks[0].Label != "Français" || idx.Default != e.Key {
		t.Errorf("after update: %+v", idx)
	}
	raw, _ := os.ReadFile(s.trackFile("m1", e.Key))
	if !strings.Contains(string(raw), "00:00:02.500 --> 00:00:03.500") {
		t.Errorf("timing should be shifted by 1.5 s, got %q", raw)
	}
}

func TestUpdateRejections(t *testing.T) {
	s := testService(t)
	if err := s.addUpload("m1", "en", "English", []byte("WEBVTT\n\n")); err != nil {
		t.Fatal(err)
	}
	key := s.loadIndex("m1").Tracks[0].Key

	empty := "   "
	if re := s.update("m1", key, updateBody{Label: &empty}); re == nil || re.code != 400 || re.key != "invalid" {
		t.Errorf("an empty name should be refused with 400 invalid, got %v", re)
	}

	label := "x"
	if re := s.update("m1", "en", updateBody{Label: &label}); re == nil || re.key != "read_only" {
		t.Errorf("a sidecar file can't be renamed, got %v", re)
	}
	shift := 100
	if re := s.update("m1", "en", updateBody{ShiftMs: &shift}); re == nil || re.key != "read_only" {
		t.Errorf("a sidecar file can't be shifted, got %v", re)
	}

	// A sidecar can still be made the default.
	def := true
	if re := s.update("m1", "en", updateBody{Default: &def}); re != nil {
		t.Errorf("a sidecar can be the default: %v", re)
	}
	if s.loadIndex("m1").Default != "en" {
		t.Error("the default should be saved")
	}

	// A stored track whose file vanished.
	_ = os.Remove(s.trackFile("m1", key))
	if re := s.update("m1", key, updateBody{ShiftMs: &shift}); re == nil || re.code != 404 {
		t.Errorf("a missing file should give 404, got %v", re)
	}
}

func TestUpdateClampsShift(t *testing.T) {
	s := testService(t)
	if err := s.addUpload("m1", "en", "English", []byte("00:10:00.000 --> 00:10:01.000")); err != nil {
		t.Fatal(err)
	}
	key := s.loadIndex("m1").Tracks[0].Key
	huge := 99999999
	if re := s.update("m1", key, updateBody{ShiftMs: &huge}); re != nil {
		t.Fatal(re)
	}
	raw, _ := os.ReadFile(s.trackFile("m1", key))
	if !strings.HasPrefix(string(raw), "00:20:00.000") { // 10 minutes is the most one shift can move
		t.Errorf("the shift is limited to 10 minutes, got %q", raw)
	}
}
