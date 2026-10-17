package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewUsesCacheFolder(t *testing.T) {
	l := New(filepath.Join("data"))
	if want := filepath.Join("data", "cache"); l.Root != want {
		t.Errorf("Root = %q, want %q", l.Root, want)
	}
}

func TestPaths(t *testing.T) {
	l := Layout{Root: "root"}
	cases := map[string]string{
		l.Image(Posters, "abc"):   filepath.Join("root", "posters", "abc.jpg"),
		l.Image(Backdrops, "abc"): filepath.Join("root", "backdrops", "abc.jpg"),
		l.Compat("abc"):           filepath.Join("root", "compat", "abc.mp4"),
		l.SubsDir("abc"):          filepath.Join("root", "subs", "abc"),
		l.ImageDir(Posters):       filepath.Join("root", "posters"),
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestIDsCannotEscapeTheCache(t *testing.T) {
	l := Layout{Root: "root"}
	for _, evil := range []string{"../../etc/passwd", "..", "a/../../b", `..\..\x`} {
		for _, p := range []string{l.Image(Posters, evil), l.Compat(evil), l.SubsDir(evil)} {
			rel, err := filepath.Rel("root", p)
			if err != nil || rel == ".." || len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
				t.Errorf("id %q produced %q, which leaves the cache", evil, p)
			}
		}
	}
}

func TestEnsureCreatesFolders(t *testing.T) {
	l := Layout{Root: filepath.Join(t.TempDir(), "cache")}
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"posters", "backdrops", "compat"} {
		if st, err := os.Stat(filepath.Join(l.Root, d)); err != nil || !st.IsDir() {
			t.Errorf("%s folder was not created", d)
		}
	}
}

func TestCleanPartialConversions(t *testing.T) {
	l := Layout{Root: t.TempDir()}
	if err := l.Ensure(); err != nil {
		t.Fatal(err)
	}
	done := l.Compat("movie1")
	part := filepath.Join(l.Root, "compat", "movie2.part.mp4")
	for _, p := range []string{done, part} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	l.CleanPartialConversions()
	if _, err := os.Stat(part); !os.IsNotExist(err) {
		t.Error("the half written file should be removed")
	}
	if _, err := os.Stat(done); err != nil {
		t.Error("a finished conversion must be kept")
	}
}
