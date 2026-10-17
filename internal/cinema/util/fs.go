package util

import "os"

// FileExists reports whether p is a regular, non-empty file. Empty files count as
// missing because every file this program writes is checked right after a rename, and
// a zero-byte result means the writer failed.
func FileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir() && st.Size() > 0
}
