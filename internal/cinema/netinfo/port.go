package netinfo

import "os"

// DefaultPort is where the server listens unless CINEMA_PORT says otherwise.
const DefaultPort = "8765"

// ListenPort returns the HTTP port: CINEMA_PORT when set, else DefaultPort.
func ListenPort() string {
	if p := os.Getenv("CINEMA_PORT"); p != "" {
		return p
	}
	return DefaultPort
}
