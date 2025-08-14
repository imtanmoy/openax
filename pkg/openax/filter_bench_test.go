package openax

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// BenchmarkApplyFilter_Small benchmarks filtering a small OpenAPI spec
func BenchmarkApplyFilter_Small(b *testing.B) {
	doc := createTestAPISpec(10, 2) // 10 paths, 2 operations each
	opts := FilterOptions{
		Tags: []string{"users"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := applyFilter(doc, opts)
		if err != nil {
			b.Fatalf("Filter failed: %v", err)
		}
	}
}

// BenchmarkApplyFilter_Medium benchmarks filtering a medium OpenAPI spec
func BenchmarkApplyFilter_Medium(b *testing.B) {
	doc := createTestAPISpec(100, 4) // 100 paths, 4 operations each
	opts := FilterOptions{
		Tags: []string{"users", "posts"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := applyFilter(doc, opts)
		if err != nil {
			b.Fatalf("Filter failed: %v", err)
		}
	}
}

// BenchmarkApplyFilter_Large benchmarks filtering a large OpenAPI spec
func BenchmarkApplyFilter_Large(b *testing.B) {
	doc := createTestAPISpec(500, 6) // 500 paths, 6 operations each
	opts := FilterOptions{
		Tags: []string{"users", "posts", "comments"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := applyFilter(doc, opts)
		if err != nil {
			b.Fatalf("Filter failed: %v", err)
		}
	}
}

// BenchmarkSchemaReferenceExtraction benchmarks schema reference extraction from complex schemas
func BenchmarkSchemaReferenceExtraction(b *testing.B) {
	schema := createComplexSchema(50) // Schema with 50 properties

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		refs := make(map[string]bool) // Reset for each iteration
		err := extractSchemaReferences(schema, refs)
		if err != nil {
			b.Fatalf("Schema extraction failed: %v", err)
		}
	}
}

// BenchmarkPathMatching benchmarks path matching performance
func BenchmarkPathMatching(b *testing.B) {
	paths := generateTestPaths(1000)        // 1000 test paths
	filters := []string{"/api/v1", "/auth"} // Common filters

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			pathMatchesFilter(path, filters)
		}
	}
}

// BenchmarkValidateRef benchmarks reference validation
func BenchmarkValidateRef(b *testing.B) {
	refs := []string{
		"#/components/schemas/User",
		"#/components/parameters/UserId",
		"#/components/responses/UserResponse",
		"#/components/requestBodies/UserRequest",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ref := range refs {
			_, err := validateRef(ref, nil)
			if err != nil {
				b.Fatalf("Reference validation failed: %v", err)
			}
		}
	}
}

// BenchmarkDeepSchemaTraversal benchmarks traversing deeply nested schemas
func BenchmarkDeepSchemaTraversal(b *testing.B) {
	schema := createDeeplyNestedSchema(20) // 20 levels deep

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		refs := make(map[string]bool)
		err := extractSchemaReferences(schema, refs)
		if err != nil {
			b.Fatalf("Deep schema traversal failed: %v", err)
		}
	}
}

// Helper functions for creating test data

func createTestAPISpec(numPaths, numOpsPerPath int) *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Benchmark API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
		Tags: []*openapi3.Tag{
			{Name: "users"},
			{Name: "posts"},
			{Name: "comments"},
		},
	}

	// Create schemas
	for i := 0; i < numPaths; i++ {
		schemaName := fmt.Sprintf("Schema%d", i)
		doc.Components.Schemas[schemaName] = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"id": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"integer"},
						},
					},
					"name": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
			},
		}
	}

	// Create paths with operations
	for i := 0; i < numPaths; i++ {
		pathStr := fmt.Sprintf("/resource%d", i)
		pathItem := &openapi3.PathItem{}

		// Add different operations based on numOpsPerPath
		if numOpsPerPath >= 1 {
			pathItem.Get = createTestOperation("users", fmt.Sprintf("Schema%d", i))
		}
		if numOpsPerPath >= 2 {
			pathItem.Post = createTestOperation("posts", fmt.Sprintf("Schema%d", i))
		}
		if numOpsPerPath >= 3 {
			pathItem.Put = createTestOperation("comments", fmt.Sprintf("Schema%d", i))
		}
		if numOpsPerPath >= 4 {
			pathItem.Delete = createTestOperation("users", fmt.Sprintf("Schema%d", i))
		}
		if numOpsPerPath >= 5 {
			pathItem.Patch = createTestOperation("posts", fmt.Sprintf("Schema%d", i))
		}
		if numOpsPerPath >= 6 {
			pathItem.Head = createTestOperation("comments", fmt.Sprintf("Schema%d", i))
		}

		doc.Paths.Set(pathStr, pathItem)
	}

	return doc
}

func createTestOperation(tag, schemaName string) *openapi3.Operation {
	description := "OK"
	operation := &openapi3.Operation{
		Tags:      []string{tag},
		Responses: &openapi3.Responses{},
	}

	operation.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &description,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s", schemaName),
					},
				},
			},
		},
	})

	return operation
}

func createComplexSchema(numProperties int) *openapi3.SchemaRef {
	properties := make(openapi3.Schemas)

	for i := 0; i < numProperties; i++ {
		propName := fmt.Sprintf("prop%d", i)
		switch i % 3 {
		case 0:
			// Reference to another schema
			properties[propName] = &openapi3.SchemaRef{
				Ref: fmt.Sprintf("#/components/schemas/RefSchema%d", i),
			}
		case 1:
			// Array of references
			properties[propName] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"array"},
					Items: &openapi3.SchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/ArrayItem%d", i),
					},
				},
			}
		default:
			// Simple property
			properties[propName] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			}
		}
	}

	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: properties,
		},
	}
}

func createDeeplyNestedSchema(depth int) *openapi3.SchemaRef {
	if depth == 0 {
		return &openapi3.SchemaRef{
			Ref: "#/components/schemas/DeepRef",
		}
	}

	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"nested": createDeeplyNestedSchema(depth - 1),
			},
		},
	}
}

func generateTestPaths(count int) []string {
	paths := make([]string, count)
	for i := 0; i < count; i++ {
		switch i % 4 {
		case 0:
			paths[i] = fmt.Sprintf("/api/v1/users/%d", i)
		case 1:
			paths[i] = fmt.Sprintf("/api/v1/posts/%d", i)
		case 2:
			paths[i] = fmt.Sprintf("/auth/login/%d", i)
		case 3:
			paths[i] = fmt.Sprintf("/admin/config/%d", i)
		}
	}
	return paths
}
