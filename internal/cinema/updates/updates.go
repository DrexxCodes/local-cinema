// Package updates tells the admin when a newer release exists on GitHub. The server asks
// GitHub, never a guest's browser.
package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
)

// Info is what the admin Settings page shows about updates.
type Info struct {
	Enabled     bool   `json:"enabled"` // false for development builds
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	Available   bool   `json:"available"`
	URL         string `json:"url"` // release page
	Notes       string `json:"notes"`
	PublishedAt string `json:"published_at"`
	AssetName   string `json:"asset_name"` // the download that fits this computer
	AssetURL    string `json:"asset_url"`
	CheckedAt   string `json:"checked_at"`
	Error       string `json:"error"`
}

type release struct {
	Tag         string  `json:"tag_name"`
	Body        string  `json:"body"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Service checks for updates.
type Service struct {
	Version string // running version, "dev" for development builds
	Repo    string // GitHub "owner/name"; empty disables checking
	APIBase string // defaults to https://api.github.com
	Client  *http.Client
	GOOS    string
	GOARCH  string

	mu   sync.Mutex
	info Info
}

// New returns a checker for the running version.
func New(version, repo string) *Service {
	return &Service{Version: version, Repo: repo}
}

func (s *Service) enabled() bool { return s.Repo != "" && s.Version != "dev" }

func (s *Service) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

func (s *Service) apiBase() string {
	if s.APIBase != "" {
		return strings.TrimRight(s.APIBase, "/")
	}
	return "https://api.github.com"
}

func (s *Service) goos() string {
	if s.GOOS != "" {
		return s.GOOS
	}
	return runtime.GOOS
}

func (s *Service) goarch() string {
	if s.GOARCH != "" {
		return s.GOARCH
	}
	return runtime.GOARCH
}

// Info returns the result of the last check.
func (s *Service) Info() Info {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.info
	u.Current = s.Version
	u.Enabled = s.enabled()
	return u
}

// Available is the latest version when it is newer than the running one, else "".
func (s *Service) Available() string {
	if u := s.Info(); u.Available {
		return u.Latest
	}
	return ""
}

// Check asks GitHub for the latest release. Failures (offline, rate limit) are stored as a
// message, not raised.
func (s *Service) Check() {
	u := Info{Current: s.Version, Enabled: s.enabled()}
	u.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	defer func() { s.mu.Lock(); s.info = u; s.mu.Unlock() }()
	if !u.Enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.apiBase()+"/repos/"+s.Repo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "local-cinema/"+s.Version)
	res, err := s.client().Do(req)
	if err != nil {
		u.Error = "Could not reach GitHub. Check the internet connection."
		return
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusNotFound:
		u.Error = "No release has been published yet."
		return
	case res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusTooManyRequests:
		u.Error = "GitHub is limiting requests right now. Try again later."
		return
	case res.StatusCode != http.StatusOK:
		u.Error = fmt.Sprintf("GitHub answered with status %d.", res.StatusCode)
		return
	}
	var rel release
	if err := json.NewDecoder(res.Body).Decode(&rel); err != nil || rel.Tag == "" {
		u.Error = "GitHub's answer could not be read."
		return
	}
	u.Latest = strings.TrimPrefix(rel.Tag, "v")
	u.Available = IsNewer(u.Latest, u.Current)
	u.URL, u.PublishedAt = rel.HTMLURL, rel.PublishedAt
	u.Notes = rel.Body
	if len(u.Notes) > 4000 {
		u.Notes = u.Notes[:4000] + "..."
	}
	u.AssetName, u.AssetURL = PickAsset(rel.Assets, s.goos(), s.goarch())
}

// Loop checks once when the app starts, then once a day. It never returns.
func (s *Service) Loop() {
	time.Sleep(5 * time.Second) // let the server finish starting first
	s.Check()
	for range time.Tick(24 * time.Hour) {
		s.Check()
	}
}

// Mount registers the admin endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.GET("/update", func(e *core.RequestEvent) error { return e.JSON(200, s.Info()) })
	rt.Admin.POST("/update/check", func(e *core.RequestEvent) error {
		s.Check()
		return e.JSON(200, s.Info())
	})
}
