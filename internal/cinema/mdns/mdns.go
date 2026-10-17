// Package mdns announces the cinema on the local network, so phones can open
// http://cinema.local:8765 without knowing the IP address.
package mdns

import (
	"log/slog"
	"strconv"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/pocketbase/dbx"

	"local-cinema/internal/cinema/netinfo"
	"local-cinema/internal/cinema/store"
)

const serviceType = "_localcinema._tcp"

// Service announces the cinema.
type Service struct {
	st  *store.Store
	mu  sync.Mutex
	srv *zeroconf.Server
}

// New returns the announcer.
func New(st *store.Store) *Service { return &Service{st: st} }

// HostName is the name announced as <name>.local, "cinema" when the setting is empty.
func HostName(setting string) string {
	if setting == "" {
		return "cinema"
	}
	return setting
}

// TXT is the information sent along with the announcement.
func TXT(name string, setupComplete bool, movies int64) []string {
	return []string{
		"name=" + name,
		"version=1.0",
		"setup=" + strconv.FormatBool(setupComplete),
		"movies=" + strconv.FormatInt(movies, 10),
	}
}

// Start announces the service. It only logs when the network does not allow it.
func (s *Service) Start() {
	cfg := s.st.Settings()
	name := cfg.GetString("cinema_name")
	host := HostName(cfg.GetString("mdns_name"))
	port, _ := strconv.Atoi(netinfo.ListenPort())
	var ips []string
	for _, a := range netinfo.LAN() {
		ips = append(ips, a.IP)
	}
	count, _ := s.st.App.CountRecords("movies", dbx.HashExp{"available": true})
	txt := TXT(name, cfg.GetBool("setup_complete"), count)

	var srv *zeroconf.Server
	var err error
	if len(ips) > 0 {
		// RegisterProxy also publishes <host>.local so http://cinema.local works.
		srv, err = zeroconf.RegisterProxy(name, serviceType, "local.", port, host, ips, txt, nil)
	} else {
		srv, err = zeroconf.Register(name, serviceType, "local.", port, txt, nil)
	}
	if err != nil {
		slog.Warn("mDNS unavailable, use the IP address instead", "err", err)
		srv = nil
	}
	s.mu.Lock()
	s.srv = srv
	s.mu.Unlock()
}

// Stop ends the announcement.
func (s *Service) Stop() {
	s.mu.Lock()
	srv := s.srv
	s.srv = nil
	s.mu.Unlock()
	if srv != nil {
		srv.Shutdown()
	}
}

// Restart announces again after the cinema name or the network name changed.
func (s *Service) Restart() {
	s.Stop()
	s.Start()
}
