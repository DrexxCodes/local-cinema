package netinfo

import (
	"net"
	"testing"
)

func TestExternalBase(t *testing.T) {
	lan := []Addr{{Name: "eth0", IP: "192.168.1.20"}, {Name: "wlan0", IP: "10.0.0.5"}}
	cases := []struct {
		name string
		host string
		tls  bool
		lan  []Addr
		want string
	}{
		{"local name becomes the LAN IP", "cinema.local:8765", false, lan, "http://192.168.1.20:8765"},
		{"an IP host is kept", "192.168.1.77:8765", false, lan, "http://192.168.1.77:8765"},
		{"no port", "cinema.local", false, lan, "http://192.168.1.20"},
		{"https", "cinema.local:8765", true, lan, "https://192.168.1.20:8765"},
		{"no LAN address keeps the name", "cinema.local:8765", false, nil, "http://cinema.local:8765"},
		{"IPv6 with port", "[fe80::1]:8765", false, lan, "http://[fe80::1]:8765"},
		{"localhost is a name, so it is swapped", "localhost:8765", false, lan, "http://192.168.1.20:8765"},
	}
	for _, c := range cases {
		if got := ExternalBase(c.host, c.tls, c.lan); got != c.want {
			t.Errorf("%s: ExternalBase(%q) = %q, want %q", c.name, c.host, got, c.want)
		}
	}
}

func TestLANOnlyReturnsUsableIPv4(t *testing.T) {
	for _, a := range LAN() {
		ip := net.ParseIP(a.IP)
		if ip == nil || ip.To4() == nil {
			t.Errorf("%q is not an IPv4 address", a.IP)
		}
		if ip != nil && ip.IsLoopback() {
			t.Errorf("loopback address %q should be skipped", a.IP)
		}
		if a.Name == "" {
			t.Errorf("address %q has no interface name", a.IP)
		}
	}
}

func TestListenPort(t *testing.T) {
	t.Setenv("CINEMA_PORT", "")
	if got := ListenPort(); got != "8765" {
		t.Errorf("default port = %q, want 8765", got)
	}
	t.Setenv("CINEMA_PORT", "9000")
	if got := ListenPort(); got != "9000" {
		t.Errorf("port from the environment = %q, want 9000", got)
	}
}
