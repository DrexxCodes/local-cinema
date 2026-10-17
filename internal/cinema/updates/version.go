package updates

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseVersion reads "1.2.3" or "v1.2.3-beta.1" into numbers plus a pre-release flag.
func ParseVersion(v string) (nums [3]int, pre bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		pre = v[i] == '-'
		v = v[:i]
	}
	for i, part := range strings.SplitN(v, ".", 3) {
		nums[i], _ = strconv.Atoi(part)
	}
	return
}

// IsNewer reports whether version a is newer than b.
func IsNewer(a, b string) bool {
	an, apre := ParseVersion(a)
	bn, bpre := ParseVersion(b)
	for i := range an {
		if an[i] != bn[i] {
			return an[i] > bn[i]
		}
	}
	return bpre && !apre // 1.2.0 is newer than 1.2.0-beta.1
}

// Asset is one downloadable file of a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// PickAsset chooses the download for a computer: the installer on Windows, the archive for
// the operating system and CPU elsewhere.
func PickAsset(assets []Asset, goos, goarch string) (name, url string) {
	for _, a := range assets {
		switch {
		case goos == "windows" && strings.HasPrefix(a.Name, "LocalCinema-Setup-") && strings.HasSuffix(a.Name, ".exe"):
			return a.Name, a.URL
		case goos != "windows" && strings.Contains(a.Name, fmt.Sprintf("_%s_%s.", goos, goarch)):
			return a.Name, a.URL
		}
	}
	return "", ""
}
