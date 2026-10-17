package store

import "testing"

func TestFoldersKey(t *testing.T) {
	if got := FoldersKey(nil); got != "" {
		t.Errorf("no folders = %q, want empty", got)
	}
	a := FoldersKey([]string{"/movies", "/more"})
	b := FoldersKey([]string{"/more", "/movies"})
	if a == b {
		t.Error("the order of folders is part of the key")
	}
	if a != "/movies|/more" {
		t.Errorf("key = %q", a)
	}
	if FoldersKey([]string{"/a"}) == FoldersKey([]string{"/a", ""}) {
		t.Error("an extra empty folder must change the key")
	}
}
