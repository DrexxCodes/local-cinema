package main

import (
	"os"
	"reflect"
	"testing"
)

func TestTakeFlag(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()

	os.Args = []string{"prog", "--background"}
	if !takeFlag("--background") || !reflect.DeepEqual(os.Args, []string{"prog"}) {
		t.Fatalf("flag not removed: %v", os.Args)
	}
	os.Args = []string{"prog", "serve", "--http=0.0.0.0:1"}
	if takeFlag("--background") || len(os.Args) != 3 {
		t.Fatalf("args changed without the flag: %v", os.Args)
	}
}
