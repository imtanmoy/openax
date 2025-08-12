package cmd_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/imtanmoy/openax/cmd"
)

func TestNewApp(t *testing.T) {
	app := cmd.NewApp()
	require.NotNil(t, app, "NewApp() should not return nil")

	assert.Equal(t, "openax", app.Name, "App name should be 'openax'")
	assert.NotEmpty(t, app.Usage, "App usage should not be empty")
	assert.NotEmpty(t, app.Version, "App version should not be empty")
	assert.NotEmpty(t, app.Flags, "App should have flags defined")
}

func TestAppFlags(t *testing.T) {
	app := cmd.NewApp()

	expectedFlags := []string{
		"input", "output", "format", "paths", "operations", "tags", "validate-only",
	}

	flagNames := make(map[string]bool)
	for _, flag := range app.Flags {
		switch f := flag.(type) {
		case interface{ Names() []string }:
			for _, name := range f.Names() {
				flagNames[name] = true
			}
		}
	}

	for _, expectedFlag := range expectedFlags {
		assert.True(t, flagNames[expectedFlag], "Expected flag '%s' not found", expectedFlag)
	}
}

// Test that app can be created and run with help flag
func TestAppHelp(t *testing.T) {
	app := cmd.NewApp()

	// Create test args for help
	args := []string{"openax", "--help"}

	// This would normally exit, but we can test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "App panicked with help flag", "panic: %v", r)
		}
	}()

	// We expect this to "exit" with help, so we catch that
	err := app.Run(context.Background(), args)
	// The help flag should trigger an exit, so we might get an error
	// This is expected behavior for CLI apps
	_ = err // It's ok if this errors due to help display
}

// Test CLI integration with actual files
func TestCLIIntegration(t *testing.T) {
	app := cmd.NewApp()

	// Get the test spec path
	specPath := filepath.Join("..", "testdata", "specs", "simple.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Test spec file not found, skipping CLI integration test")
	}

	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "validate only",
			args:        []string{"openax", "--validate-only", "-i", specPath},
			expectError: false,
		},
		{
			name:        "filter by tags",
			args:        []string{"openax", "-i", specPath, "--tags", "users", "--format", "json"},
			expectError: false,
		},
		{
			name:        "missing input file",
			args:        []string{"openax", "--tags", "users"},
			expectError: true,
		},
		{
			name:        "non-existent file",
			args:        []string{"openax", "-i", "nonexistent.yaml"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}