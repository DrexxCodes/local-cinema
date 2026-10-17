package main

import "os"

// takeFlag removes name from os.Args and says whether it was there. It is used for the
// program's own flags, so PocketBase's command line parser never sees them.
func takeFlag(name string) bool {
	args := []string{os.Args[0]}
	found := false
	for _, a := range os.Args[1:] {
		if a == name {
			found = true
			continue
		}
		args = append(args, a)
	}
	os.Args = args
	return found
}
