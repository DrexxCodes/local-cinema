// Package netinfo answers "what address can other devices use to reach this computer".
package netinfo

import (
	"net"
	"strings"
)

// Addr is one IPv4 address of a network interface.
type Addr struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// LAN lists the IPv4 addresses of every interface that is up, except loopback.
// It returns nil when there are none.
func LAN() []Addr {
	var out []Addr
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if v4 := ipn.IP.To4(); v4 != nil {
					out = append(out, Addr{Name: i.Name, IP: v4.String()})
				}
			}
		}
	}
	return out
}

// ExternalBase is the address another app on the network should use to reach this
// server, for example "http://192.168.1.20:8765". A host name such as drexx-movies.local
// does not resolve in many players, so it is swapped for the first LAN address.
// requestHost is the Host header of the request, tls says whether it came over HTTPS.
func ExternalBase(requestHost string, tls bool, lan []Addr) string {
	host, port, err := net.SplitHostPort(requestHost)
	if err != nil {
		host, port = requestHost, ""
	}
	if net.ParseIP(host) == nil && len(lan) > 0 {
		host = lan[0].IP
	}
	if port != "" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	scheme := "http"
	if tls {
		scheme = "https"
	}
	return scheme + "://" + host
}
