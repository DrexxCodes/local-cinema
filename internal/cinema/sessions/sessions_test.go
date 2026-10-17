package sessions

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	cases := map[float64]int{0: 1, -3: 1, 0.9: 1, 1: 1, 2: 2, 2.7: 2, 50: 50}
	for in, want := range cases {
		if got := Limit(in); got != want {
			t.Errorf("Limit(%v) = %d, want %d", in, got, want)
		}
	}
}

func TestValidSlots(t *testing.T) {
	for _, n := range []int{1, 2, 20} {
		if !ValidSlots(n) {
			t.Errorf("%d should be allowed", n)
		}
	}
	for _, n := range []int{-1, 0, 21, 1000} {
		if ValidSlots(n) {
			t.Errorf("%d should be refused", n)
		}
	}
}

func TestRaisedLimit(t *testing.T) {
	cases := []struct{ cur, extra, want int }{
		{2, 3, 5}, {1, 1, 2}, {49, 5, 50}, {50, 1, 50}, {0, 0, 1},
	}
	for _, c := range cases {
		if got := RaisedLimit(c.cur, c.extra); got != c.want {
			t.Errorf("RaisedLimit(%d, %d) = %d, want %d", c.cur, c.extra, got, c.want)
		}
	}
}

func TestCutoffs(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	browser, external := Cutoffs(now)
	if browser != "2026-10-11 11:58:00.000Z" {
		t.Errorf("browser cutoff = %q (2 minutes back)", browser)
	}
	if external != "2026-10-11 11:30:00.000Z" {
		t.Errorf("external cutoff = %q (30 minutes back)", external)
	}
	if browser <= external {
		t.Error("external players get longer, so their cutoff is earlier")
	}
}

func TestCutoffsUseUTC(t *testing.T) {
	lagos := time.FixedZone("WAT", 3600)
	local := time.Date(2026, 10, 11, 13, 0, 0, 0, lagos) // 12:00 UTC
	a, _ := Cutoffs(local)
	b, _ := Cutoffs(local.UTC())
	if a != b || !strings.HasSuffix(a, "Z") {
		t.Errorf("the database stores UTC: %q vs %q", a, b)
	}
}

func TestDeviceKey(t *testing.T) {
	if got := DeviceKey("phone-1", "sess9"); got != "ext:phone-1" {
		t.Errorf("got %q", got)
	}
	if got := DeviceKey("", "sess9"); got != "ext:sess9" {
		t.Errorf("a device without an id uses its session: %q", got)
	}
}

func TestLinks(t *testing.T) {
	q := AuthQuery("a b+c", "s/1")
	parsed, err := url.ParseQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Get("token") != "a b+c" || parsed.Get("sid") != "s/1" {
		t.Errorf("the query must round trip special characters, got %q", q)
	}
	base := "http://192.168.1.20:8765"
	if got := StreamLink(base, "m1", q); got != base+"/stream/m1?"+q {
		t.Errorf("StreamLink = %q", got)
	}
	got := SubtitleLink(base, "m1", "en us/x", q)
	if !strings.HasPrefix(got, base+"/api/movies/m1/subtitles/en%20us%2Fx?") {
		t.Errorf("the track key must be path escaped: %q", got)
	}
	u, err := url.Parse(got)
	if err != nil || u.Query().Get("sid") != "s/1" {
		t.Errorf("the link must carry the session: %q (%v)", got, err)
	}
}
