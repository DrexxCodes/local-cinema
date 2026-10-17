// Package setup is the first run wizard and the public cinema info the home page needs.
package setup

import (
	"github.com/pocketbase/pocketbase/core"

	"local-cinema/internal/cinema/httpx"
	"local-cinema/internal/cinema/media"
	"local-cinema/internal/cinema/netinfo"
	"local-cinema/internal/cinema/scanner"
	"local-cinema/internal/cinema/store"
	"strings"
)

// Service serves the wizard endpoints.
type Service struct {
	st      *store.Store
	scanner *scanner.Service
}

// New returns the setup service.
func New(st *store.Store, sc *scanner.Service) *Service { return &Service{st: st, scanner: sc} }

// Mount registers the endpoints. They are public, and the ones that change things refuse to
// work once setup is complete.
func (s *Service) Mount(rt httpx.Routes) {
	rt.Public.GET("/api/cinema", s.handleCinema)
	rt.Public.GET("/api/setup/status", s.handleStatus)
	rt.Public.GET("/api/setup/interfaces", s.handleInterfaces)
	rt.Public.GET("/api/setup/scan-status", s.handleScanStatus)
	rt.Public.POST("/api/setup/step", s.handleStep)
	rt.Public.POST("/api/setup/validate-folder", s.handleValidateFolder)
	rt.Public.POST("/api/setup/complete", s.handleComplete)
}

// stepKeys are the settings the wizard may write.
var stepKeys = map[string]bool{
	"cinema_name": true, "tagline": true, "accent_color": true, "movie_folders": true,
	"watch_folders": true, "global_max_streams": true, "mdns_name": true, "tmdb_api_key": true,
}

// StepData keeps only the settings the wizard is allowed to write.
func StepData(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if stepKeys[k] {
			out[k] = v
		}
	}
	return out
}

// ValidAdmin checks the admin account asked for at the end of the wizard.
func ValidAdmin(email, password string) bool {
	return strings.Contains(email, "@") && len(password) >= 8
}

// open sends 403 and returns false when setup is already complete.
func (s *Service) open(e *core.RequestEvent) bool {
	if s.st.Settings().GetBool("setup_complete") {
		_ = httpx.Err(e, 403, "setup_complete", "Setup is already complete.")
		return false
	}
	return true
}

// handleCinema is the public info the home page needs.
func (s *Service) handleCinema(e *core.RequestEvent) error {
	c := s.st.Settings()
	return e.JSON(200, map[string]any{
		"name":           c.GetString("cinema_name"),
		"tagline":        c.GetString("tagline"),
		"accent_color":   c.GetString("accent_color"),
		"setup_complete": c.GetBool("setup_complete"),
	})
}

func (s *Service) handleStatus(e *core.RequestEvent) error {
	c := s.st.Settings()
	out := map[string]any{"setup_complete": c.GetBool("setup_complete")}
	if c.GetBool("setup_complete") {
		return e.JSON(200, out)
	}
	out["setup_step"] = int(c.GetFloat("setup_step"))
	out["cinema_name"] = c.GetString("cinema_name")
	out["tagline"] = c.GetString("tagline")
	out["accent_color"] = c.GetString("accent_color")
	out["movie_folders"] = s.st.Folders()
	out["watch_folders"] = c.GetBool("watch_folders")
	out["global_max_streams"] = int(c.GetFloat("global_max_streams"))
	out["mdns_name"] = c.GetString("mdns_name")
	out["has_tmdb_key"] = c.GetString("tmdb_api_key") != ""
	return e.JSON(200, out)
}

func (s *Service) handleInterfaces(e *core.RequestEvent) error {
	return e.JSON(200, netinfo.LAN())
}

func (s *Service) handleScanStatus(e *core.RequestEvent) error {
	return e.JSON(200, s.scanner.State.Snapshot())
}

func (s *Service) handleStep(e *core.RequestEvent) error {
	if !s.open(e) {
		return nil
	}
	var in struct {
		Step int            `json:"step"`
		Data map[string]any `json:"data"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	c := s.st.Settings()
	for k, v := range StepData(in.Data) {
		c.Set(k, v)
	}
	if in.Step > 0 {
		c.Set("setup_step", in.Step)
	}
	if err := s.st.App.Save(c); err != nil {
		return httpx.Err(e, 400, "invalid", "Could not save: "+err.Error())
	}
	return httpx.OK(e)
}

func (s *Service) handleValidateFolder(e *core.RequestEvent) error {
	if !s.open(e) {
		return nil
	}
	var in struct {
		Path string `json:"path"`
	}
	if err := e.BindBody(&in); err != nil {
		return httpx.Err(e, 400, "bad_request", "Invalid request.")
	}
	return e.JSON(200, media.ValidateFolder(in.Path))
}

func (s *Service) handleComplete(e *core.RequestEvent) error {
	if !s.open(e) {
		return nil
	}
	var in struct {
		AdminEmail    string `json:"admin_email"`
		AdminPassword string `json:"admin_password"`
	}
	if err := e.BindBody(&in); err != nil || !ValidAdmin(in.AdminEmail, in.AdminPassword) {
		return httpx.Err(e, 400, "bad_request", "Enter a valid email and a password of at least 8 characters.")
	}
	if _, err := s.st.App.FindAuthRecordByEmail(core.CollectionNameSuperusers, in.AdminEmail); err != nil {
		coll, err := s.st.App.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return httpx.Err(e, 500, "internal", "Server error.")
		}
		su := core.NewRecord(coll)
		su.SetEmail(in.AdminEmail)
		su.SetPassword(in.AdminPassword)
		if err := s.st.App.Save(su); err != nil {
			return httpx.Err(e, 400, "invalid", "Could not create the admin: "+err.Error())
		}
	}
	c := s.st.Settings()
	c.Set("setup_complete", true)
	c.Set("setup_step", 5)
	if err := s.st.App.Save(c); err != nil {
		return httpx.Err(e, 500, "internal", "Could not finish setup.")
	}
	go s.scanner.Run()
	return httpx.OK(e)
}
