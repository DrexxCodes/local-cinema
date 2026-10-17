package cinema

import (
	"testing"
)

func TestListenPort(t *testing.T) {
	t.Setenv("CINEMA_PORT", "")
	if ListenPort() != "8765" {
		t.Errorf("default port = %s", ListenPort())
	}
	t.Setenv("CINEMA_PORT", "9000")
	if ListenPort() != "9000" {
		t.Errorf("CINEMA_PORT is not used, got %s", ListenPort())
	}
}
