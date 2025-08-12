package openax_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/imtanmoy/openax/pkg/openax"
)

func TestNew(t *testing.T) {
	client := openax.New()
	require.NotNil(t, client, "New() should not return nil")
}

func TestNewWithOptions(t *testing.T) {
	opts := openax.LoadOptions{
		AllowExternalRefs: false,
		Context:           context.Background(),
	}
	client := openax.NewWithOptions(opts)
	require.NotNil(t, client, "NewWithOptions() should not return nil")
}

func TestLoadFromFile(t *testing.T) {
	client := openax.New()
	
	testCases := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid spec",
			filePath:    "../../testdata/specs/simple.yaml",
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
			doc, err := client.LoadFromFile(tc.filePath)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for %s", tc.name)
				assert.Nil(t, doc, "Document should be nil on error")
				return
			}
			
			require.NoError(t, err, "Unexpected error for %s", tc.name)
			require.NotNil(t, doc, "Document should not be nil")
		})
	}
}

func TestValidate(t *testing.T) {
	client := openax.New()
	
	doc, err := client.LoadFromFile("../../testdata/specs/simple.yaml")
	require.NoError(t, err, "Failed to load spec")
	
	err = client.Validate(doc)
	assert.NoError(t, err, "Validation should succeed for valid spec")
}

func TestValidateOnly(t *testing.T) {
	client := openax.New()
	
	testCases := []struct {
		name        string
		source      string
		expectError bool
	}{
		{
			name:        "valid spec",
			source:      "../../testdata/specs/simple.yaml",
			expectError: false,
		},
		{
			name:        "invalid spec",
			source:      "../../testdata/specs/invalid.yaml",
			expectError: true,
		},
		{
			name:        "non-existent file",
			source:      "../../testdata/specs/nonexistent.yaml",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.ValidateOnly(tc.source)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Unexpected error for %s", tc.name)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	client := openax.New()
	
	doc, err := client.LoadFromFile("../../testdata/specs/simple.yaml")
	require.NoError(t, err, "Failed to load spec")
	
	testCases := []struct {
		name           string
		options        openax.FilterOptions
		expectedPaths  int
		expectedSchema int
	}{
		{
			name:           "no filters",
			options:        openax.FilterOptions{},
			expectedPaths:  2, // /users and /posts
			expectedSchema: 3, // User, CreateUser, Post
		},
		{
			name: "filter by tags - users",
			options: openax.FilterOptions{
				Tags: []string{"users"},
			},
			expectedPaths:  1, // only /users
			expectedSchema: 2, // User, CreateUser
		},
		{
			name: "filter by tags - posts",
			options: openax.FilterOptions{
				Tags: []string{"posts"},
			},
			expectedPaths:  1, // only /posts
			expectedSchema: 2, // Post, User (referenced by Post.author)
		},
		{
			name: "filter by operations - GET",
			options: openax.FilterOptions{
				Operations: []string{"get"},
			},
			expectedPaths:  2, // both paths have GET
			expectedSchema: 2, // User (from /users GET) and Post+User (from /posts GET, Post.author refs User)
		},
		{
			name: "filter by operations - POST",
			options: openax.FilterOptions{
				Operations: []string{"post"},
			},
			expectedPaths:  1, // only /users has POST
			expectedSchema: 2, // User, CreateUser
		},
		{
			name: "filter by paths",
			options: openax.FilterOptions{
				Paths: []string{"/users"},
			},
			expectedPaths:  1, // only /users
			expectedSchema: 2, // User, CreateUser
		},
		{
			name: "combine filters - users tag AND get operation",
			options: openax.FilterOptions{
				Tags:       []string{"users"},
				Operations: []string{"get"},
			},
			expectedPaths:  1, // only /users GET
			expectedSchema: 1, // only User schema
		},
		{
			name: "no matching filters",
			options: openax.FilterOptions{
				Tags: []string{"nonexistent"},
			},
			expectedPaths:  0, // no paths match
			expectedSchema: 0, // no schemas needed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered, err := client.Filter(doc, tc.options)
			require.NoError(t, err, "Filter should not fail")
			require.NotNil(t, filtered, "Filtered document should not be nil")
			
			actualPaths := filtered.Paths.Len()
			assert.Equal(t, tc.expectedPaths, actualPaths, "Path count mismatch")
			
			actualSchemas := len(filtered.Components.Schemas)
			assert.Equal(t, tc.expectedSchema, actualSchemas, "Schema count mismatch")
			
			// Verify the filtered document still has required fields
			assert.NotEmpty(t, filtered.OpenAPI, "Filtered document should have OpenAPI version")
			assert.NotNil(t, filtered.Info, "Filtered document should have Info")
		})
	}
}

func TestLoadAndFilter(t *testing.T) {
	client := openax.New()
	
	testCases := []struct {
		name        string
		source      string
		options     openax.FilterOptions
		expectError bool
	}{
		{
			name:   "valid spec with filter",
			source: "../../testdata/specs/simple.yaml",
			options: openax.FilterOptions{
				Tags: []string{"users"},
			},
			expectError: false,
		},
		{
			name:   "non-existent file",
			source: "../../testdata/specs/nonexistent.yaml",
			options: openax.FilterOptions{
				Tags: []string{"users"},
			},
			expectError: true,
		},
		{
			name:   "invalid spec",
			source: "../../testdata/specs/invalid.yaml",
			options: openax.FilterOptions{
				Tags: []string{"users"},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered, err := client.LoadAndFilter(tc.source, tc.options)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for %s", tc.name)
				assert.Nil(t, filtered, "Document should be nil on error")
				return
			}
			
			require.NoError(t, err, "Unexpected error for %s", tc.name)
			require.NotNil(t, filtered, "Filtered document should not be nil")
		})
	}
}

func TestFilterOptions(t *testing.T) {
	// Test that FilterOptions struct can be created and used
	opts := openax.FilterOptions{
		Paths:      []string{"/users", "/posts"},
		Operations: []string{"get", "post"},
		Tags:       []string{"public", "v1"},
	}
	
	assert.Len(t, opts.Paths, 2, "Expected 2 paths")
	assert.Len(t, opts.Operations, 2, "Expected 2 operations")
	assert.Len(t, opts.Tags, 2, "Expected 2 tags")
}