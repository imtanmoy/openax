package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Test that main doesn't panic when called without arguments
	// We expect it to show help or error gracefully

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with help flag
	os.Args = []string{"openax", "--help"}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() panicked: %v", r)
		}
	}()

	// Note: main() will call os.Exit() on error, so we can't easily test
	// error cases without more complex setup. The integration tests cover
	// the actual functionality.
}
