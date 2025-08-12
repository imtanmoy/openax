// Package main demonstrates how to create custom filtering logic using openax.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/imtanmoy/openax/pkg/loader"
	"github.com/imtanmoy/openax/pkg/openax"
	"github.com/imtanmoy/openax/pkg/validator"
)

func main() {
	// Example: Custom filtering logic
	customFilterExample()
}

func customFilterExample() {
	fmt.Println("=== Custom Filter Example ===")

	// Use individual packages for more control
	l := loader.New()
	v := validator.New()

	// Load the spec
	doc, err := l.LoadFromFile("../../testdata/specs/petstore.yaml")
	if err != nil {
		log.Fatalf("Failed to load spec: %v", err)
	}

	// Validate it
	if err := v.Validate(doc); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Apply custom filtering logic
	filtered := customFilter(doc)

	fmt.Printf("Original spec had %d paths\n", doc.Paths.Len())
	fmt.Printf("Custom filtered spec has %d paths\n", filtered.Paths.Len())

	// Now apply standard openax filtering on top
	client := openax.New()
	finalFiltered, err := client.Filter(filtered, openax.FilterOptions{
		Operations: []string{"get"},
	})
	if err != nil {
		log.Fatalf("Failed to apply final filter: %v", err)
	}

	fmt.Printf("Final filtered spec has %d paths\n", finalFiltered.Paths.Len())
}

// customFilter demonstrates custom filtering logic
// This example filters out any paths that contain "upload" or "delete"
func customFilter(doc *openapi3.T) *openapi3.T {
	filtered := &openapi3.T{
		OpenAPI:      doc.OpenAPI,
		Info:         doc.Info,
		Servers:      doc.Servers,
		ExternalDocs: doc.ExternalDocs,
		Tags:         doc.Tags,
		Security:     doc.Security,
		Components:   doc.Components,
		Paths:        &openapi3.Paths{},
	}

	// Custom logic: exclude paths containing "upload" or operations with "delete"
	for path, pathItem := range doc.Paths.Map() {
		// Skip paths containing "upload"
		if strings.Contains(strings.ToLower(path), "upload") {
			continue
		}

		// Create a new path item without delete operations
		newPathItem := &openapi3.PathItem{
			Ref:         pathItem.Ref,
			Summary:     pathItem.Summary,
			Description: pathItem.Description,
			Servers:     pathItem.Servers,
			Parameters:  pathItem.Parameters,
		}

		// Copy all operations except DELETE
		if pathItem.Get != nil {
			newPathItem.Get = pathItem.Get
		}
		if pathItem.Post != nil {
			newPathItem.Post = pathItem.Post
		}
		if pathItem.Put != nil {
			newPathItem.Put = pathItem.Put
		}
		if pathItem.Patch != nil {
			newPathItem.Patch = pathItem.Patch
		}
		if pathItem.Head != nil {
			newPathItem.Head = pathItem.Head
		}
		if pathItem.Options != nil {
			newPathItem.Options = pathItem.Options
		}
		if pathItem.Trace != nil {
			newPathItem.Trace = pathItem.Trace
		}
		// Intentionally skip pathItem.Delete

		// Only include if it has at least one operation
		if hasAnyOperation(newPathItem) {
			filtered.Paths.Set(path, newPathItem)
		}
	}

	return filtered
}

func hasAnyOperation(pathItem *openapi3.PathItem) bool {
	return pathItem.Get != nil ||
		pathItem.Post != nil ||
		pathItem.Put != nil ||
		pathItem.Patch != nil ||
		pathItem.Head != nil ||
		pathItem.Options != nil ||
		pathItem.Trace != nil
}
