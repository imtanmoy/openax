package openax

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const okDescription = "OK"

func TestComponentPruningBasic(t *testing.T) {
	t.Run("prune unused schemas", func(t *testing.T) {
		// Create a spec with used and unused schemas
		doc := createTestSpecWithUnusedComponents()

		// Filter to only include paths that reference UsedSchema
		filteredDoc, err := applyFilter(doc, FilterOptions{
			Paths:           []string{"/users"},
			PruneComponents: true,
		})

		require.NoError(t, err)

		// Should contain UsedSchema but not UnusedSchema
		assert.Contains(t, filteredDoc.Components.Schemas, "UsedSchema")
		assert.NotContains(t, filteredDoc.Components.Schemas, "UnusedSchema")
	})

	t.Run("preserve transitively used schemas", func(t *testing.T) {
		doc := createTestSpecWithTransitiveReferences()

		// Filter to only include the main path
		filteredDoc, err := applyFilter(doc, FilterOptions{
			Paths:           []string{"/main"},
			PruneComponents: true,
		})

		require.NoError(t, err)

		// Should contain all transitively referenced schemas
		assert.Contains(t, filteredDoc.Components.Schemas, "MainSchema")
		assert.Contains(t, filteredDoc.Components.Schemas, "NestedSchema")
		assert.Contains(t, filteredDoc.Components.Schemas, "DeepSchema")
		// Should not contain unrelated schema
		assert.NotContains(t, filteredDoc.Components.Schemas, "UnrelatedSchema")
	})
}

// Helper functions to create test data

func createTestSpecWithUnusedComponents() *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	// Create used schema
	doc.Components.Schemas["UsedSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"id": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
				},
			},
		},
	}

	// Create unused schema
	doc.Components.Schemas["UnusedSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"name": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
			},
		},
	}

	// Create path that uses only UsedSchema
	description := okDescription
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
	}

	pathItem.Get.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &description,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: "#/components/schemas/UsedSchema",
					},
				},
			},
		},
	})

	doc.Paths.Set("/users", pathItem)

	return doc
}

func createTestSpecWithTransitiveReferences() *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	// Create DeepSchema (leaf)
	doc.Components.Schemas["DeepSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"string"},
		},
	}

	// Create NestedSchema that references DeepSchema
	doc.Components.Schemas["NestedSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"deep": &openapi3.SchemaRef{
					Ref: "#/components/schemas/DeepSchema",
				},
			},
		},
	}

	// Create MainSchema that references NestedSchema
	doc.Components.Schemas["MainSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"nested": &openapi3.SchemaRef{
					Ref: "#/components/schemas/NestedSchema",
				},
			},
		},
	}

	// Create an unrelated schema
	doc.Components.Schemas["UnrelatedSchema"] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"boolean"},
		},
	}

	// Create path that uses MainSchema
	description := okDescription
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
	}

	pathItem.Get.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &description,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: "#/components/schemas/MainSchema",
					},
				},
			},
		},
	})

	doc.Paths.Set("/main", pathItem)

	return doc
}
