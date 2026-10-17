package settings

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"local-cinema/internal/cinema/util"
)

var (
	hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	hostNameRe = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)
)

// The checks below return the cleaned value, or an error whose text is shown to the admin.

// CinemaName trims the name and checks its length (1 to 50 characters).
func CinemaName(s string) (string, error) {
	n := strings.TrimSpace(s)
	if n == "" || len(n) > 50 {
		return "", errors.New("The cinema name must be 1 to 50 characters.")
	}
	return n, nil
}

// Tagline trims the tagline and checks it is at most 120 characters. It may be empty.
func Tagline(s string) (string, error) {
	t := strings.TrimSpace(s)
	if len(t) > 120 {
		return "", errors.New("The tagline can be at most 120 characters.")
	}
	return t, nil
}

// AccentColor checks for a color such as #7C3AED.
func AccentColor(s string) (string, error) {
	if !hexColorRe.MatchString(s) {
		return "", errors.New("The accent color must look like #7C3AED.")
	}
	return s, nil
}

// HostName lower cases and checks the network name (letters, numbers and dashes, up to 63).
func HostName(s string) (string, error) {
	h := strings.ToLower(strings.TrimSpace(s))
	if !hostNameRe.MatchString(h) {
		return "", errors.New("The network name can use letters, numbers and dashes only.")
	}
	return h, nil
}

// MovieFolders trims the list, drops empty entries, and checks that each folder exists.
// At least one folder must remain.
func MovieFolders(in []string) ([]string, error) {
	clean := []string{}
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err != nil || !st.IsDir() {
			return nil, errors.New("This folder does not exist or is not readable: " + p)
		}
		clean = append(clean, p)
	}
	if len(clean) == 0 {
		return nil, errors.New("Keep at least one movie folder.")
	}
	return clean, nil
}

// MaxStreams limits the cinema wide stream count to 1 to 1000.
func MaxStreams(n int) int { return util.ClampInt(n, 1, 1000) }

// TMDBKey decides what to store for the TMDB key. clear removes it. A blank key leaves the
// stored one alone (change is false).
func TMDBKey(key string, clear bool) (value string, change bool) {
	if clear {
		return "", true
	}
	if k := strings.TrimSpace(key); k != "" {
		return k, true
	}
	return "", false
}

// Email lower cases and checks an admin email. An empty one means "no change" (ok is true, email is "").
func Email(s string) (email string, err error) {
	em := strings.ToLower(strings.TrimSpace(s))
	if em != "" && !strings.Contains(em, "@") {
		return "", errors.New("Enter a valid email.")
	}
	return em, nil
}

// Password checks a new admin password (at least 8 characters).
func Password(s string) error {
	if len(s) < 8 {
		return errors.New("Use a password with at least 8 characters.")
	}
	return nil
}
