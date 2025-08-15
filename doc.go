// Package openax provides a powerful OpenAPI 3.x specification filtering tool and Go library.
//
// OpenAx allows you to filter large OpenAPI specifications by paths, operations, and tags
// while automatically resolving and including only the referenced components. This is
// particularly useful for creating focused API documentation, generating clients for
// specific service areas, or extracting public APIs from comprehensive specifications.
//
// # Key Features
//
//   - Smart filtering by paths, HTTP operations, and tags
//   - Automatic dependency resolution for components
//   - Support for multiple input sources (files, URLs, raw data)
//   - Built-in OpenAPI 3.x validation
//   - Component pruning to reduce specification size
//   - High performance with proper reference resolution
//
// # Installation
//
// Install the CLI tool via npm:
//
//	npm install -g openax-cli
//
// Or use as a Go library:
//
//	go get github.com/imtanmoy/openax
//
// # Command Line Usage
//
// Filter an OpenAPI specification by tags:
//
//	openax -i api.yaml --tags "users,orders" --format json
//
// Filter by operations and paths:
//
//	openax -i api.yaml --operations "get,post" --paths "/api/v1"
//
// Validate a specification:
//
//	openax --validate-only -i api.yaml
//
// # Library Usage
//
// Basic filtering example:
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"github.com/imtanmoy/openax/pkg/openax"
//	)
//
//	func main() {
//		client := openax.New()
//
//		// Load and filter by tags
//		filtered, err := client.LoadAndFilter("api.yaml", openax.FilterOptions{
//			Tags: []string{"users", "orders"},
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Printf("Filtered spec has %d paths\n", len(filtered.Paths))
//	}
//
// Advanced usage with custom options:
//
//	client := openax.NewWithOptions(openax.LoadOptions{
//		AllowExternalRefs: false,
//		Context:           ctx,
//	})
//
//	// Load from different sources
//	doc, err := client.LoadFromFile("api.yaml")
//	doc, err = client.LoadFromURL("https://api.example.com/openapi.yaml")
//	doc, err = client.LoadFromData(yamlBytes)
//
//	// Apply sophisticated filtering
//	filtered, err := client.Filter(doc, openax.FilterOptions{
//		Paths:           []string{"/users", "/orders"},
//		Operations:      []string{"get", "post", "updateUser"},
//		Tags:            []string{"public", "v1"},
//		PruneComponents: true,
//	})
//
// # Use Cases
//
//   - API Documentation: Create focused docs from large specifications
//   - Client Generation: Generate clients for specific service areas
//   - Testing: Create minimal specs for testing specific functionality
//   - Micro-services: Extract service-specific APIs from monolithic specs
//   - Public APIs: Filter internal specs to expose only public endpoints
//   - Versioning: Create version-specific API specifications
//
// # Error Handling
//
// The library provides detailed error information for common scenarios:
//
//	doc, err := client.LoadAndFilter("api.yaml", options)
//	if err != nil {
//		var validationErr *openapi3.ValidationError
//		if errors.As(err, &validationErr) {
//			fmt.Printf("Validation failed: %v\n", validationErr)
//			return
//		}
//		fmt.Printf("Failed to load and filter: %v\n", err)
//		return
//	}
//
// # Performance Considerations
//
// For optimal performance:
//
//   - Use specific path filters rather than broad patterns
//   - Enable component pruning for large specifications
//   - Validate specifications before expensive filtering operations
//   - Reuse client instances for multiple operations
//
// # Package Structure
//
//   - pkg/openax: Main library package with client and filtering logic
//   - pkg/loader: Utilities for loading specifications from various sources
//   - pkg/validator: OpenAPI specification validation utilities
//   - cmd/: CLI implementation
//   - examples/: Usage examples and demonstrations
//
// For more information, visit: https://github.com/imtanmoy/openax
package main