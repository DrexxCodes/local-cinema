package tmdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func testClient(srv *httptest.Server) *Client {
	c := New()
	c.HTTP = srv.Client()
	c.APIBase = srv.URL + "/3"
	c.PosterBase = srv.URL + "/img/w500"
	c.BackdropBase = srv.URL + "/img/w1280"
	return c
}

const searchBody = `{"results": [
  {"id": 603, "overview": "A hacker learns the truth.", "poster_path": "/poster.jpg", "backdrop_path": "/back.jpg",
   "genre_ids": [28, 878, 99999], "vote_average": 8.2, "release_date": "1999-03-30"},
  {"id": 604, "overview": "The sequel.", "genre_ids": [28], "vote_average": 7.0, "release_date": "2003-05-15"}
]}`

func TestSearchPicksFirstResult(t *testing.T) {
	var got struct{ key, query, year, path string }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		got.key, got.query, got.year = r.URL.Query().Get("api_key"), r.URL.Query().Get("query"), r.URL.Query().Get("year")
		_, _ = w.Write([]byte(searchBody))
	}))
	defer srv.Close()

	m, ok := testClient(srv).Search("KEY", "The Matrix", 1999)
	if !ok {
		t.Fatal("expected a match")
	}
	if got.path != "/3/search/movie" || got.key != "KEY" || got.query != "The Matrix" || got.year != "1999" {
		t.Errorf("request was %+v", got)
	}
	if m.ID != 603 || m.Overview != "A hacker learns the truth." || m.Rating != 8.2 || m.Year != 1999 {
		t.Errorf("match = %+v", m)
	}
	if !slices.Equal(m.Genres, []string{"Action", "Science Fiction"}) {
		t.Errorf("genres = %v (unknown ids are dropped)", m.Genres)
	}
	if m.PosterPath != "/poster.jpg" || m.BackdropPath != "/back.jpg" {
		t.Errorf("paths = %q, %q", m.PosterPath, m.BackdropPath)
	}
}

func TestSearchOmitsYearWhenUnknown(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte(searchBody))
	}))
	defer srv.Close()
	if _, ok := testClient(srv).Search("K", "Heat", 0); !ok {
		t.Fatal("expected a match")
	}
	if want := "api_key=K&query=Heat"; query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
}

func TestSearchFailures(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"bad key":      func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
		"server error": func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
		"not json":     func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("<html>")) },
		"no results":   func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"results": []}`)) },
	}
	for name, h := range cases {
		srv := httptest.NewServer(h)
		if _, ok := testClient(srv).Search("K", "Anything", 0); ok {
			t.Errorf("%s: should not report a match", name)
		}
		srv.Close()
	}
}

func TestSearchOffline(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	c := testClient(srv)
	srv.Close() // nothing is listening any more
	if _, ok := c.Search("K", "Anything", 0); ok {
		t.Error("an unreachable server should not report a match")
	}
}

func TestSearchYearParsing(t *testing.T) {
	cases := map[string]int{
		`"1999-03-30"`: 1999,
		`"2003"`:       2003,
		`""`:           0,
		`"abc"`:        0,
		`"99"`:         0,
	}
	for date, want := range cases {
		body := `{"results":[{"id":1,"release_date":` + date + `}]}`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
		m, ok := testClient(srv).Search("K", "x", 0)
		srv.Close()
		if !ok || m.Year != want {
			t.Errorf("release_date %s: year = %d (ok=%v), want %d", date, m.Year, ok, want)
		}
	}
}

func TestGenreNames(t *testing.T) {
	if got := GenreNames([]int{18, 27, 1, 10749}); !slices.Equal(got, []string{"Drama", "Horror", "Romance"}) {
		t.Errorf("GenreNames = %v", got)
	}
	if got := GenreNames(nil); got == nil || len(got) != 0 {
		t.Errorf("no ids should give an empty, non-nil list (it is stored as JSON), got %#v", got)
	}
}

func TestImageURLs(t *testing.T) {
	c := New()
	m := Match{PosterPath: "/p.jpg", BackdropPath: "/b.jpg"}
	if got, want := c.PosterURL(m), "https://image.tmdb.org/t/p/w500/p.jpg"; got != want {
		t.Errorf("PosterURL = %q, want %q", got, want)
	}
	if got, want := c.BackdropURL(m), "https://image.tmdb.org/t/p/w1280/b.jpg"; got != want {
		t.Errorf("BackdropURL = %q, want %q", got, want)
	}
}

func TestDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok.jpg" {
			_, _ = w.Write([]byte("JPEGDATA"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := testClient(srv)
	dir := t.TempDir()

	dest := filepath.Join(dir, "a.jpg")
	if !c.Download(context.Background(), srv.URL+"/ok.jpg", dest) {
		t.Fatal("download should succeed")
	}
	if b, _ := os.ReadFile(dest); string(b) != "JPEGDATA" {
		t.Errorf("file = %q", b)
	}
	if _, err := os.Stat(dest + ".download"); !os.IsNotExist(err) {
		t.Error("the temporary file should be gone")
	}

	missing := filepath.Join(dir, "b.jpg")
	if c.Download(context.Background(), srv.URL+"/nope.jpg", missing) {
		t.Error("a 404 must not count as a download")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Error("a failed download must not leave a file")
	}

	if c.Download(context.Background(), srv.URL+"/ok.jpg", filepath.Join(dir, "no-such-dir", "c.jpg")) {
		t.Error("an unwritable destination must fail")
	}
}

func TestDownloadDoesNotDamageExistingImage(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	dest := filepath.Join(t.TempDir(), "keep.jpg")
	if err := os.WriteFile(dest, []byte("FRAME FROM THE VIDEO"), 0o644); err != nil {
		t.Fatal(err)
	}
	if testClient(srv).Download(context.Background(), srv.URL+"/x.jpg", dest) {
		t.Fatal("download should fail")
	}
	if b, _ := os.ReadFile(dest); string(b) != "FRAME FROM THE VIDEO" {
		t.Errorf("a failed download must keep the existing picture, got %q", b)
	}
}
