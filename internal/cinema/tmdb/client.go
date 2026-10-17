// Package tmdb looks movies up on The Movie Database (themoviedb.org) for posters,
// synopses and ratings. It is optional: without an API key nothing here is called.
package tmdb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var genres = map[int]string{
	28: "Action", 12: "Adventure", 16: "Animation", 35: "Comedy", 80: "Crime", 99: "Documentary",
	18: "Drama", 10751: "Family", 14: "Fantasy", 36: "History", 27: "Horror", 10402: "Music",
	9648: "Mystery", 10749: "Romance", 878: "Science Fiction", 10770: "TV Movie", 53: "Thriller",
	10752: "War", 37: "Western",
}

// GenreNames turns TMDB genre ids into names. Unknown ids are dropped.
func GenreNames(ids []int) []string {
	names := []string{}
	for _, id := range ids {
		if g, ok := genres[id]; ok {
			names = append(names, g)
		}
	}
	return names
}

// Client talks to the TMDB API. The base addresses can be replaced, which is how the
// tests point it at a local server.
type Client struct {
	HTTP         *http.Client
	APIBase      string // https://api.themoviedb.org/3
	PosterBase   string // https://image.tmdb.org/t/p/w500
	BackdropBase string // https://image.tmdb.org/t/p/w1280
}

// New returns a client for the real service with a 15 second timeout.
func New() *Client {
	return &Client{
		HTTP:         &http.Client{Timeout: 15 * time.Second},
		APIBase:      "https://api.themoviedb.org/3",
		PosterBase:   "https://image.tmdb.org/t/p/w500",
		BackdropBase: "https://image.tmdb.org/t/p/w1280",
	}
}

// Match is the best search result for a title.
type Match struct {
	ID           int
	Overview     string
	PosterPath   string // empty when TMDB has no poster
	BackdropPath string
	Genres       []string
	Rating       float64
	Year         int // from the release date, 0 when unknown
}

// PosterURL is the download address of the poster.
func (c *Client) PosterURL(m Match) string { return c.PosterBase + m.PosterPath }

// BackdropURL is the download address of the backdrop.
func (c *Client) BackdropURL(m Match) string { return c.BackdropBase + m.BackdropPath }

// Search finds the best match for a title (and year, when known). ok is false when the
// internet is unreachable, the key is refused, or nothing matches. Callers just try
// again on the next scan.
func (c *Client) Search(key, title string, year int) (m Match, ok bool) {
	q := url.Values{"api_key": {key}, "query": {title}}
	if year > 0 {
		q.Set("year", strconv.Itoa(year))
	}
	resp, err := c.HTTP.Get(c.APIBase + "/search/movie?" + q.Encode())
	if err != nil {
		return Match{}, false
	}
	defer resp.Body.Close()
	var res struct {
		Results []struct {
			ID           int     `json:"id"`
			Overview     string  `json:"overview"`
			PosterPath   string  `json:"poster_path"`
			BackdropPath string  `json:"backdrop_path"`
			GenreIDs     []int   `json:"genre_ids"`
			VoteAverage  float64 `json:"vote_average"`
			ReleaseDate  string  `json:"release_date"`
		} `json:"results"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&res) != nil || len(res.Results) == 0 {
		return Match{}, false
	}
	r := res.Results[0]
	m = Match{
		ID: r.ID, Overview: r.Overview, PosterPath: r.PosterPath, BackdropPath: r.BackdropPath,
		Genres: GenreNames(r.GenreIDs), Rating: r.VoteAverage,
	}
	if len(r.ReleaseDate) >= 4 {
		m.Year, _ = strconv.Atoi(r.ReleaseDate[:4])
	}
	return m, true
}

// Download saves an image to dest. It reports whether the file was written completely.
// The picture is written beside dest first, so a failed download never leaves a broken file.
func (c *Client) Download(ctx context.Context, src, dest string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return false
	}
	resp, err := c.HTTP.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return false
	}
	defer resp.Body.Close()
	tmp := dest + ".download"
	f, err := os.Create(tmp)
	if err != nil {
		return false
	}
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return false
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return false
	}
	return true
}
