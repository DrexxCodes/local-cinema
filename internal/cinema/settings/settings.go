// Package settings is the admin's Settings page: the cinema's name, colour, folders and
// network name, and the admin account.
package settings

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/media"
	"local-cinema/internal/cinema/netinfo"
	"local-cinema/internal/cinema/store"
	"local-cinema/internal/version"
)

// Service serves the settings endpoints.
type Service struct {
	st *store.Store

	// OnIdentityChanged runs in the background after the cinema name or network name changed,
	// so the network announcement (mDNS) can be repeated.
	OnIdentityChanged func()
}

// New returns the settings service.
func New(st *store.Store) *Service { return &Service{st: st} }

// Mount registers the endpoints.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Admin.GET("/settings", s.handleGet)
	rt.Admin.PUT("/settings", s.handlePut)
	rt.Admin.POST("/validate-folder", s.handleValidateFolder)
	rt.Admin.PUT("/account", s.handleAccount)
}

func (s *Service) dto(admin *core.Record) map[string]any {
	c := s.st.Settings()
	return map[string]any{
		"cinema_name":        c.GetString("cinema_name"),
		"tagline":            c.GetString("tagline"),
		"accent_color":       c.GetString("accent_color"),
		"movie_folders":      s.st.Folders(),
		"watch_folders":      c.GetBool("watch_folders"),
		"global_max_streams": int(c.GetFloat("global_max_streams")),
		"mdns_name":          c.GetString("mdns_name"),
		"has_tmdb_key":       c.GetString("tmdb_api_key") != "",
		"admin_email":        admin.Email(),
		"version":            version.Version,
		"lan_ips":            netinfo.LAN(),
	}
}

func (s *Service) handleGet(e *core.RequestEvent) error {
	return e.JSON(200, s.dto(e.Auth))
}

// handlePut edits everything the setup wizard asked for. Only the keys present in the body
// change. Folder changes trigger a rescan through the settings update hook.
func (s *Service) handlePut(e *core.RequestEvent) error {
	var in struct {
		CinemaName       *string   `json:"cinema_name"`
		Tagline          *string   `json:"tagline"`
		AccentColor      *string   `json:"accent_color"`
		MovieFolders     *[]string `json:"movie_folders"`
		WatchFolders     *bool     `json:"watch_folders"`
		GlobalMaxStreams *int      `json:"global_max_streams"`
		MdnsName         *string   `json:"mdns_name"`
		TmdbAPIKey       *string   `json:"tmdb_api_key"`
		ClearTmdbKey     bool      `json:"clear_tmdb_key"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	c := s.st.Settings()
	prevName, prevHost := c.GetString("cinema_name"), c.GetString("mdns_name")

	invalid := func(err error) error { return httpx.Err(e, 400, "invalid", err.Error()) }
	if in.CinemaName != nil {
		n, err := CinemaName(*in.CinemaName)
		if err != nil {
			return invalid(err)
		}
		c.Set("cinema_name", n)
	}
	if in.Tagline != nil {
		t, err := Tagline(*in.Tagline)
		if err != nil {
			return invalid(err)
		}
		c.Set("tagline", t)
	}
	if in.AccentColor != nil {
		a, err := AccentColor(*in.AccentColor)
		if err != nil {
			return invalid(err)
		}
		c.Set("accent_color", a)
	}
	if in.MovieFolders != nil {
		folders, err := MovieFolders(*in.MovieFolders)
		if err != nil {
			return invalid(err)
		}
		c.Set("movie_folders", folders)
	}
	if in.WatchFolders != nil {
		c.Set("watch_folders", *in.WatchFolders)
	}
	if in.GlobalMaxStreams != nil {
		c.Set("global_max_streams", MaxStreams(*in.GlobalMaxStreams))
	}
	if in.MdnsName != nil {
		h, err := HostName(*in.MdnsName)
		if err != nil {
			return invalid(err)
		}
		c.Set("mdns_name", h)
	}
	key := ""
	if in.TmdbAPIKey != nil {
		key = *in.TmdbAPIKey
	}
	if v, change := TMDBKey(key, in.ClearTmdbKey); change {
		c.Set("tmdb_api_key", v)
	}

	if err := s.st.App.Save(c); err != nil {
		return httpx.Err(e, 400, "invalid", "Could not save: "+err.Error())
	}
	if s.OnIdentityChanged != nil && (c.GetString("cinema_name") != prevName || c.GetString("mdns_name") != prevHost) {
		go s.OnIdentityChanged()
	}
	s.st.Audit("settings_updated", "", "", "", nil)
	return e.JSON(200, s.dto(e.Auth))
}

func (s *Service) handleValidateFolder(e *core.RequestEvent) error {
	var in struct {
		Path string `json:"path"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	return e.JSON(200, media.ValidateFolder(strings.TrimSpace(in.Path)))
}

// handleAccount changes the admin email and/or password.
// Body: {"current_password":"...","email":"...","new_password":"..."}
func (s *Service) handleAccount(e *core.RequestEvent) error {
	var in struct {
		CurrentPassword string `json:"current_password"`
		Email           string `json:"email"`
		NewPassword     string `json:"new_password"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	admin := e.Auth
	if !admin.ValidatePassword(in.CurrentPassword) {
		return httpx.Err(e, 403, "wrong_password", "The current password is not correct.")
	}
	relogin := false
	em, err := Email(in.Email)
	if err != nil {
		return httpx.Err(e, 400, "invalid", err.Error())
	}
	if em != "" && em != admin.Email() {
		if _, err := s.st.App.FindAuthRecordByEmail(core.CollectionNameSuperusers, em); err == nil {
			return httpx.Err(e, 400, "invalid", "Another admin already uses this email.")
		}
		admin.SetEmail(em)
	}
	if in.NewPassword != "" {
		if err := Password(in.NewPassword); err != nil {
			return httpx.Err(e, 400, "invalid", err.Error())
		}
		admin.SetPassword(in.NewPassword)
		relogin = true // PocketBase invalidates old tokens on a password change
	}
	if err := s.st.App.Save(admin); err != nil {
		return httpx.Err(e, 400, "invalid", "Could not save: "+err.Error())
	}
	return e.JSON(200, map[string]any{"ok": true, "relogin": relogin, "admin_email": admin.Email()})
}
