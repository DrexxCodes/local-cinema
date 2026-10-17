package setup

import "testing"

func TestStepDataKeepsOnlyWizardSettings(t *testing.T) {
	in := map[string]any{
		"cinema_name": "Drexx Cinema", "tagline": "t", "accent_color": "#000000",
		"movie_folders": []string{"/m"}, "watch_folders": true, "global_max_streams": 5,
		"mdns_name": "cinema", "tmdb_api_key": "k",
		"setup_complete": true, // the wizard must not be able to finish itself this way
		"setup_step":     9,
		"id":             "x",
		"admin_email":    "a@b.c",
	}
	got := StepData(in)
	if len(got) != 8 {
		t.Errorf("expected the 8 wizard keys, got %d: %v", len(got), got)
	}
	for _, bad := range []string{"setup_complete", "setup_step", "id", "admin_email"} {
		if _, ok := got[bad]; ok {
			t.Errorf("%q must not pass through", bad)
		}
	}
}

func TestStepDataEmpty(t *testing.T) {
	if got := StepData(nil); got == nil || len(got) != 0 {
		t.Errorf("got %#v", got)
	}
}

func TestValidAdmin(t *testing.T) {
	cases := []struct {
		email, pw string
		want      bool
	}{
		{"a@b.c", "12345678", true},
		{"a@b.c", "1234567", false},
		{"ab.c", "12345678", false},
		{"", "", false},
	}
	for _, c := range cases {
		if got := ValidAdmin(c.email, c.pw); got != c.want {
			t.Errorf("ValidAdmin(%q, %q) = %v, want %v", c.email, c.pw, got, c.want)
		}
	}
}
