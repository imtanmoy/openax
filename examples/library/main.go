// Package main demonstrates how to use openax as a library.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/imtanmoy/openax/pkg/openax"
)

func main() {
	// Example 1: Basic usage with default client
	basicExample()

	// Example 2: Advanced usage with custom options
	advancedExample()

	// Example 3: Validation only
	validationExample()
}

func basicExample() {
	fmt.Println("=== Basic Library Usage ===")

	// Create a client with default options
	client := openax.New()

	// Filter by tags
	result, err := client.LoadAndFilter("testdata/specs/petstore.yaml", openax.FilterOptions{
		Tags: []string{"pet"},
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Print the number of paths in the filtered result
	fmt.Printf("Filtered spec contains %d paths\n", result.Paths.Len())
	fmt.Printf("Filtered spec contains %d schemas\n", len(result.Components.Schemas))
	fmt.Println()
}

func advancedExample() {
	fmt.Println("=== Advanced Library Usage ===")

	// Create a client with custom options
	client := openax.NewWithOptions(openax.LoadOptions{
		AllowExternalRefs: false,
		Context:           context.Background(),
	})

	// Load the document first
	doc, err := client.LoadFromFile("testdata/specs/petstore.yaml")
	if err != nil {
		log.Printf("Error loading: %v", err)
		return
	}

	// Validate it
	if err := client.Validate(doc); err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	// Filter with multiple criteria
	filtered, err := client.Filter(doc, openax.FilterOptions{
		Operations: []string{"get", "post"},
		Tags:       []string{"user"},
	})
	if err != nil {
		log.Printf("Error filtering: %v", err)
		return
	}

	// Convert to JSON for inspection
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		log.Printf("Error marshaling: %v", err)
		return
	}

	fmt.Printf("Filtered spec (first 200 chars): %s...\n", string(data[:min(200, len(data))]))
	fmt.Println()
}

func validationExample() {
	fmt.Println("=== Validation Only ===")

	client := openax.New()

	// Just validate without filtering
	if err := client.ValidateOnly("testdata/specs/petstore.yaml"); err != nil {
		log.Printf("Validation failed: %v", err)
		return
	}

	fmt.Println("✅ OpenAPI spec is valid!")
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}