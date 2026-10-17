package auth

import (
	"net/url"
	"testing"
	"time"
)

func TestCredentials(t *testing.T) {
	q := func(s string) url.Values { v, _ := url.ParseQuery(s); return v }
	cases := []struct {
		name            string
		header, session string
		query           url.Values
		tok, sid        string
	}{
		{"headers", "Bearer abc", "s1", nil, "abc", "s1"},
		{"query string for video and img tags", "", "", q("token=abc&sid=s1"), "abc", "s1"},
		{"headers win over the query", "Bearer h", "hs", q("token=q&sid=qs"), "h", "hs"},
		{"mixed", "Bearer h", "", q("sid=qs"), "h", "qs"},
		{"nothing", "", "", nil, "", ""},
		{"a header without the Bearer prefix is used as is", "raw", "s", nil, "raw", "s"},
	}
	for _, c := range cases {
		tok, sid := Credentials(c.header, c.session, c.query)
		if tok != c.tok || sid != c.sid {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", c.name, tok, sid, c.tok, c.sid)
		}
	}
}

func TestNeedsTouch(t *testing.T) {
	now := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	if NeedsTouch(now.Add(-5*time.Second), now) {
		t.Error("seen 5 seconds ago: no write needed")
	}
	if !NeedsTouch(now.Add(-11*time.Second), now) {
		t.Error("seen 11 seconds ago: write")
	}
	if !NeedsTouch(time.Time{}, now) {
		t.Error("a session that was never seen must be written")
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := map[string]string{
		"room204":              "room204@cinema.local",
		"  Room204  ":          "room204@cinema.local",
		"ROOM204@Cinema.Local": "room204@cinema.local",
		"guest@example.com":    "guest@example.com",
		"":                     "",
		"   ":                  "",
	}
	for in, want := range cases {
		if got := NormalizeEmail(in); got != want {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHeartbeatValues(t *testing.T) {
	cases := []struct {
		rtt   float64
		sig   int
		wantR float64
		wantS int
	}{
		{35, 3, 35, 3},
		{-4, -1, 0, 0},
		{999999, 9, 60000, 4},
	}
	for _, c := range cases {
		r, s := HeartbeatValues(c.rtt, c.sig)
		if r != c.wantR || s != c.wantS {
			t.Errorf("HeartbeatValues(%v, %d) = (%v, %d), want (%v, %d)", c.rtt, c.sig, r, s, c.wantR, c.wantS)
		}
	}
}
