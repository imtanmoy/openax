package cmd_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/imtanmoy/openax/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	app := cmd.NewApp()
	require.NotNil(t, app, "NewApp() should not return nil")

	assert.Equal(t, "openax", app.Name, "App name should be 'openax'")
	assert.NotEmpty(t, app.Usage, "App usage should not be empty")
	assert.NotEmpty(t, app.Flags, "App should have flags defined")

	// Version is set dynamically in main.go, so we don't test it here
}

func TestAppFlags(t *testing.T) {
	app := cmd.NewApp()

	expectedFlags := []string{
		"input", "output", "format", "paths", "operations", "tags", "validate-only",
	}

	flagNames := make(map[string]bool)
	for _, flag := range app.Flags {
		if f, ok := flag.(interface{ Names() []string }); ok {
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
			name:        "dry run with filters",
			args:        []string{"openax", "-i", specPath, "--tags", "users", "--dry-run"},
			expectError: false,
		},
		{
			name:        "dry run without filters",
			args:        []string{"openax", "-i", specPath, "--dry-run"},
			expectError: false,
		},
		{
			name:        "unsupported output format",
			args:        []string{"openax", "-i", specPath, "--format", "toml"},
			expectError: true,
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

func TestCLIOutputFileWrite(t *testing.T) {
	app := cmd.NewApp()
	specPath := filepath.Join("..", "testdata", "specs", "simple.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Test spec file not found, skipping output file test")
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "filtered.yaml")

	err := app.Run(context.Background(), []string{"openax", "-i", specPath, "-o", outputPath, "--tags", "users"})
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "openapi:")
}

func TestCLIDryRunSummaryOutput(t *testing.T) {
	app := cmd.NewApp()
	specPath := filepath.Join("..", "testdata", "specs", "simple.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("Test spec file not found, skipping dry run summary test")
	}

	output := captureStdout(t, func() {
		err := app.Run(context.Background(), []string{"openax", "-i", specPath, "--dry-run", "--tags", "users"})
		require.NoError(t, err)
	})

	assert.Contains(t, output, "Dry Run Mode - Filtering Results Summary")
	assert.Contains(t, output, "Applied Filters")
	assert.Contains(t, output, "Tags: [users]")
	assert.Contains(t, output, "Dry run completed")
}

// captureStdout is not safe to call from parallel tests — it mutates os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, r.Close())
	}()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	func() {
		defer func() {
			assert.NoError(t, w.Close())
		}()
		fn()
	}()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.String()
}
