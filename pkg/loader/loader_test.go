package loader_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/imtanmoy/openax/pkg/loader"
)

func TestNew(t *testing.T) {
	l := loader.New()
	require.NotNil(t, l, "New() should not return nil")
}

func TestNewWithOptions(t *testing.T) {
	opts := loader.Options{
		AllowExternalRefs: false,
		Context:           context.Background(),
	}
	l := loader.NewWithOptions(opts)
	require.NotNil(t, l, "NewWithOptions() should not return nil")
}

func TestLoadFromFile(t *testing.T) {
	l := loader.New()
	
	testCases := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid simple spec",
			filePath:    "../../testdata/specs/simple.yaml",
			expectError: false,
		},
		{
			name:        "valid petstore spec",
			filePath:    "../../testdata/specs/petstore.yaml",
			expectError: false,
		},
		{
			name:        "non-existent file",
			filePath:    "../../testdata/specs/nonexistent.yaml",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := l.LoadFromFile(tc.filePath)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for %s", tc.name)
				assert.Nil(t, doc, "Document should be nil on error")
				return
			}
			
			require.NoError(t, err, "Unexpected error for %s", tc.name)
			require.NotNil(t, doc, "Document should not be nil")
			
			assert.NotNil(t, doc.Info, "Document info should not be nil")
			assert.NotEmpty(t, doc.Info.Title, "Document title should not be empty")
		})
	}
}

func TestLoadFromData(t *testing.T) {
	l := loader.New()
	
	validYAML := `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`
	

	testCases := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "valid YAML",
			data:        []byte(validYAML),
			expectError: false,
		},
		{
			name:        "invalid YAML syntax",
			data:        []byte("invalid: yaml: content: ["),
			expectError: true,
		},
		{
			name:        "empty data",
			data:        []byte{},
			expectError: false, // Empty data might be handled gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := l.LoadFromData(tc.data)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if doc == nil {
				t.Fatal("Document is nil")
			}
		})
	}
}

func TestLoadFromSource(t *testing.T) {
	l := loader.New()
	
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_spec_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name()) // Best effort cleanup
	}()
	
	validSpec := `
openapi: 3.0.3
info:
  title: Temp Test API
  version: 1.0.0
paths:
  /temp:
    get:
      responses:
        '200':
          description: OK
`
	
	if _, err := tempFile.WriteString(validSpec); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	testCases := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name:        "file path",
			source:      tempFile.Name(),
			expectError: false,
		},
		{
			name:        "http URL (will fail without network)",
			source:      "https://example.com/api.yaml",
			expectError: true, // Expected to fail in test environment
		},
		{
			name:        "non-existent file",
			source:      "/non/existent/file.yaml",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := l.LoadFromSource(tc.source)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if doc == nil {
				t.Fatal("Document is nil")
			}
		})
	}
}

func TestLoadFromURL(t *testing.T) {
	l := loader.New()
	
	testCases := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "malformed URL",
			url:         "http://[invalid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := l.LoadFromURL(tc.url)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// Helper function to get absolute path to test data
