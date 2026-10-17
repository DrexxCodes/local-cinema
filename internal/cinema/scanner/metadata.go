package scanner

import (
	"context"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/cache"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/cinema/tmdb"
)

// Metadata fills in a movie's details and pictures from TMDB.
type Metadata struct {
	st     *store.Store
	Client *tmdb.Client
}

// NewMetadata returns a TMDB helper that talks to the real service.
func NewMetadata(st *store.Store) *Metadata { return &Metadata{st: st, Client: tmdb.New()} }

// Key is the TMDB API key: CINEMA_TMDB_KEY wins over the one in Settings. "" means no lookups.
func (m *Metadata) Key() string {
	if k := os.Getenv("CINEMA_TMDB_KEY"); k != "" {
		return k
	}
	return m.st.Settings().GetString("tmdb_api_key")
}

// Fetch looks the movie up by its title and year and stores what it finds on rec.
// The caller saves rec. Nothing changes when TMDB is unreachable or has no match.
func (m *Metadata) Fetch(rec *core.Record, key string) {
	match, ok := m.Client.Search(key, rec.GetString("title"), int(rec.GetFloat("year")))
	if !ok {
		return
	}
	rec.Set("tmdb_id", match.ID)
	rec.Set("synopsis", match.Overview)
	rec.Set("genres", match.Genres)
	rec.Set("rating", match.Rating)
	if rec.GetFloat("year") == 0 && match.Year > 0 {
		rec.Set("year", match.Year)
	}
	// Only flag an image when the download worked, so a failed fetch never hides a frame
	// already extracted from the video.
	ctx := context.Background()
	if match.PosterPath != "" && m.Client.Download(ctx, m.Client.PosterURL(match), m.st.Cache.Image(cache.Posters, rec.Id)) {
		rec.Set("has_poster", true)
		rec.Set("thumb_version", time.Now().Unix())
	}
	if match.BackdropPath != "" && m.Client.Download(ctx, m.Client.BackdropURL(match), m.st.Cache.Image(cache.Backdrops, rec.Id)) {
		rec.Set("has_backdrop", true)
		rec.Set("thumb_version", time.Now().Unix())
	}
}
