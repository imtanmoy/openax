package openax_test

import (
	"context"
	"testing"

	"github.com/imtanmoy/openax/pkg/openax"
)

func TestNew(t *testing.T) {
	client := openax.New()
	if client == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewWithOptions(t *testing.T) {
	opts := openax.LoadOptions{
		AllowExternalRefs: false,
		Context:           context.Background(),
	}
	client := openax.NewWithOptions(opts)
	if client == nil {
		t.Fatal("NewWithOptions() returned nil")
	}
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

func TestValidate(t *testing.T) {
	client := openax.New()
	
	doc, err := client.LoadFromFile("../../testdata/specs/simple.yaml")
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}
	
	err = client.Validate(doc)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
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

func TestFilter(t *testing.T) {
	client := openax.New()
	
	doc, err := client.LoadFromFile("../../testdata/specs/simple.yaml")
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}
	
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
			if err != nil {
				t.Fatalf("Filter failed: %v", err)
			}
			
			if filtered == nil {
				t.Fatal("Filtered document is nil")
			}
			
			actualPaths := filtered.Paths.Len()
			if actualPaths != tc.expectedPaths {
				t.Errorf("Expected %d paths, got %d", tc.expectedPaths, actualPaths)
			}
			
			actualSchemas := len(filtered.Components.Schemas)
			if actualSchemas != tc.expectedSchema {
				t.Errorf("Expected %d schemas, got %d", tc.expectedSchema, actualSchemas)
			}
			
			// Verify the filtered document still has required fields
			if filtered.OpenAPI == "" {
				t.Error("Filtered document missing OpenAPI version")
			}
			
			if filtered.Info == nil {
				t.Error("Filtered document missing Info")
			}
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
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if filtered == nil {
				t.Fatal("Filtered document is nil")
			}
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
	
	if len(opts.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(opts.Paths))
	}
	
	if len(opts.Operations) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(opts.Operations))
	}
	
	if len(opts.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(opts.Tags))
	}
}