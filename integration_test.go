package main_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/imtanmoy/openax/pkg/openax"
)

// TestIntegrationFilterPetstore tests filtering the full petstore spec
func TestIntegrationFilterPetstore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := openax.New()
	specPath := filepath.Join("testdata", "specs", "petstore.yaml")
	testCases := createPetstoreTestCases()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runPetstoreTestCase(t, client, specPath, tc)
		})
	}
}

// PetstoreTestCase represents a test case for petstore filtering
type PetstoreTestCase struct {
	name           string
	options        openax.FilterOptions
	minPaths       int
	minSchemas     int
	requiredPaths  []string
	requiredSchema []string
}

// createPetstoreTestCases creates all test cases for petstore filtering
func createPetstoreTestCases() []PetstoreTestCase {
	return []PetstoreTestCase{
		{
			name: "filter pet operations",
			options: openax.FilterOptions{
				Tags: []string{"pet"},
			},
			minPaths:       4, // Multiple pet endpoints
			minSchemas:     3, // Pet, Category, Tag, etc.
			requiredPaths:  []string{"/pet"},
			requiredSchema: []string{"Pet"},
		},
		{
			name: "filter user operations",
			options: openax.FilterOptions{
				Tags: []string{"user"},
			},
			minPaths:       3, // User endpoints
			minSchemas:     1, // User schema
			requiredPaths:  []string{"/user"},
			requiredSchema: []string{"User"},
		},
		{
			name: "filter GET operations only",
			options: openax.FilterOptions{
				Operations: []string{"get"},
			},
			minPaths:   8, // All paths with GET operations
			minSchemas: 5, // All referenced schemas
		},
		{
			name: "filter store path prefix",
			options: openax.FilterOptions{
				Paths: []string{"/store"},
			},
			minPaths:       2, // Store endpoints
			minSchemas:     1, // Order schema
			requiredPaths:  []string{"/store/inventory"},
			requiredSchema: []string{"Order"},
		},
		{
			name: "complex filter - pet GET operations",
			options: openax.FilterOptions{
				Tags:       []string{"pet"},
				Operations: []string{"get"},
			},
			minPaths:       2, // GET pet endpoints
			minSchemas:     3, // Pet and related schemas
			requiredSchema: []string{"Pet"},
		},
	}
}

// runPetstoreTestCase executes a single petstore test case
func runPetstoreTestCase(t *testing.T, client *openax.Client, specPath string, tc PetstoreTestCase) {
	filtered, err := client.LoadAndFilter(specPath, tc.options)
	if err != nil {
		t.Fatalf("LoadAndFilter failed: %v", err)
	}

	validateMinimumExpectations(t, filtered, tc)
	validateRequiredPaths(t, filtered, tc.requiredPaths)
	validateRequiredSchemas(t, filtered, tc.requiredSchema)
	validateFilteredSpec(t, client, filtered)
}

// validateMinimumExpectations validates minimum paths and schemas
func validateMinimumExpectations(t *testing.T, filtered *openapi3.T, tc PetstoreTestCase) {
	if filtered.Paths.Len() < tc.minPaths {
		t.Errorf("Expected at least %d paths, got %d", tc.minPaths, filtered.Paths.Len())
	}

	if len(filtered.Components.Schemas) < tc.minSchemas {
		t.Errorf("Expected at least %d schemas, got %d", tc.minSchemas, len(filtered.Components.Schemas))
	}
}

// validateRequiredPaths validates that required paths are present
func validateRequiredPaths(t *testing.T, filtered *openapi3.T, requiredPaths []string) {
	for _, requiredPath := range requiredPaths {
		found := false
		for path := range filtered.Paths.Map() {
			if pathMatches(path, requiredPath) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required path %s not found in filtered results", requiredPath)
		}
	}
}

// validateRequiredSchemas validates that required schemas are present
func validateRequiredSchemas(t *testing.T, filtered *openapi3.T, requiredSchemas []string) {
	for _, requiredSchema := range requiredSchemas {
		if _, exists := filtered.Components.Schemas[requiredSchema]; !exists {
			t.Errorf("Required schema %s not found in filtered results", requiredSchema)
		}
	}
}

// validateFilteredSpec validates that the filtered spec is still valid OpenAPI
func validateFilteredSpec(t *testing.T, client *openax.Client, filtered *openapi3.T) {
	err := client.Validate(filtered)
	if err != nil {
		t.Errorf("Filtered spec is not valid: %v", err)
	}
}

// pathMatches checks if a path matches a required path (exact or prefix match)
func pathMatches(path, requiredPath string) bool {
	return path == requiredPath || (len(path) > len(requiredPath) && path[:len(requiredPath)] == requiredPath)
}

// TestIntegrationEndToEnd tests the complete workflow
func TestIntegrationEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := openax.NewWithOptions(openax.LoadOptions{
		AllowExternalRefs: true,
		Context:           context.Background(),
	})

	// Test loading
	specPath := filepath.Join("testdata", "specs", "simple.yaml")
	doc, err := client.LoadFromFile(specPath)
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}

	// Test validation
	err = client.Validate(doc)
	if err != nil {
		t.Fatalf("Spec validation failed: %v", err)
	}

	// Test filtering
	filtered, err := client.Filter(doc, openax.FilterOptions{
		Tags: []string{"users"},
	})
	if err != nil {
		t.Fatalf("Filtering failed: %v", err)
	}

	// Verify filtering results
	if filtered.Paths.Len() != 1 {
		t.Errorf("Expected 1 path after filtering, got %d", filtered.Paths.Len())
	}

	// Check that only user-related schemas are included
	expectedSchemas := map[string]bool{
		"User":       true,
		"CreateUser": true,
	}

	for schemaName := range filtered.Components.Schemas {
		if !expectedSchemas[schemaName] {
			t.Errorf("Unexpected schema %s in filtered result", schemaName)
		}
	}

	for expectedSchema := range expectedSchemas {
		if _, exists := filtered.Components.Schemas[expectedSchema]; !exists {
			t.Errorf("Expected schema %s not found in filtered result", expectedSchema)
		}
	}

	// Test that the filtered spec is still valid
	err = client.Validate(filtered)
	if err != nil {
		t.Errorf("Filtered spec is not valid: %v", err)
	}
}

// TestIntegrationErrorHandling tests error scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	client := openax.New()

	testCases := []struct {
		name   string
		action func() error
	}{
		{
			name: "load non-existent file",
			action: func() error {
				_, err := client.LoadFromFile("nonexistent.yaml")
				return err
			},
		},
		{
			name: "validate only non-existent file",
			action: func() error {
				return client.ValidateOnly("nonexistent.yaml")
			},
		},
		{
			name: "load and filter non-existent file",
			action: func() error {
				_, err := client.LoadAndFilter("nonexistent.yaml", openax.FilterOptions{})
				return err
			},
		},
		{
			name: "validate invalid spec",
			action: func() error {
				return client.ValidateOnly(filepath.Join("testdata", "specs", "invalid.yaml"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action()
			if err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

// Benchmark filtering performance
func BenchmarkFilterPetstore(b *testing.B) {
	client := openax.New()
	specPath := filepath.Join("testdata", "specs", "petstore.yaml")

	// Load once outside the benchmark
	doc, err := client.LoadFromFile(specPath)
	if err != nil {
		b.Fatalf("Failed to load spec: %v", err)
	}

	filterOptions := openax.FilterOptions{
		Tags: []string{"pet"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Filter(doc, filterOptions)
		if err != nil {
			b.Fatalf("Filter failed: %v", err)
		}
	}
}

func BenchmarkLoadAndFilter(b *testing.B) {
	client := openax.New()
	specPath := filepath.Join("testdata", "specs", "simple.yaml")

	filterOptions := openax.FilterOptions{
		Tags: []string{"users"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.LoadAndFilter(specPath, filterOptions)
		if err != nil {
			b.Fatalf("LoadAndFilter failed: %v", err)
		}
	}
}

// Helper to check if file exists
