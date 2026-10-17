package desktop

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"local-cinema/internal/cinema/netinfo"
)

var httpc = &http.Client{Timeout: 3 * time.Second}

// BaseURL is the address of the cinema as seen from this computer.
func BaseURL(port string) string { return "http://localhost:" + port }

// guestLink is the address guests type on their phones and TVs.
func guestLink(port string) string {
	if lan := netinfo.LAN(); len(lan) > 0 {
		return fmt.Sprintf("http://%s:%s", lan[0].IP, port)
	}
	return BaseURL(port)
}

// pathFor picks the first page to open from the body of /api/setup/status:
// the setup wizard until setup is complete, the admin page after that.
func pathFor(status []byte) string {
	var s struct {
		SetupComplete *bool `json:"setup_complete"`
	}
	if json.Unmarshal(status, &s) == nil && s.SetupComplete != nil && !*s.SetupComplete {
		return "/setup/"
	}
	return "/admin/"
}

// StartURL is the page "Open admin" opens.
func StartURL(port string) string {
	base := BaseURL(port)
	resp, err := httpc.Get(base + "/api/setup/status")
	if err != nil {
		return base + "/admin/"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	return base + pathFor(body)
}

// Ping says whether the server answers.
func Ping(port string) bool {
	resp, err := httpc.Get(BaseURL(port) + "/api/ping")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusNoContent
}
