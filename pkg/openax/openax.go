// Package openax provides OpenAPI 3.x specification filtering and validation capabilities.
//
// OpenAx allows you to filter large OpenAPI specifications by paths, operations, and tags
// while automatically resolving and including only the referenced components (schemas,
// parameters, request bodies, responses).
//
// # Basic Usage
//
//	client := openax.New()
//
//	// Load and filter by tags
//	filtered, err := client.LoadAndFilter("api.yaml", openax.FilterOptions{
//		Tags: []string{"users", "orders"},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Filtered spec has %d paths\n", len(filtered.Paths))
//
// # Advanced Usage
//
//	// Create client with custom options
//	client := openax.NewWithOptions(openax.LoadOptions{
//		AllowExternalRefs: true,
//		Context:           ctx,
//	})
//
//	// Load from different sources
//	doc, err := client.LoadFromFile("api.yaml")
//	doc, err = client.LoadFromURL("https://api.example.com/openapi.yaml")
//	doc, err = client.LoadFromData(yamlBytes)
//
//	// Apply filtering with multiple criteria
//	filtered, err := client.Filter(doc, openax.FilterOptions{
//		Paths:      []string{"/users", "/orders"},
//		Operations: []string{"get", "post"},
//		Tags:       []string{"public"},
//		PruneComponents: true,
//	})
//
// # Validation
//
//	// Validate only (no filtering)
//	err := client.ValidateOnly("api.yaml")
//
//	// Validate loaded document
//	err = client.Validate(doc)
//
// The package handles component dependency resolution automatically, ensuring that
// filtered specifications remain valid and complete.
package openax

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FilterOptions defines the filtering criteria for OpenAPI specifications.
//
// All fields are optional. If a field is empty, no filtering is applied for that criteria.
// Multiple criteria are combined with AND logic (all must match).
//
// Example:
//
//	opts := FilterOptions{
//		Paths:      []string{"/api/v1/users", "/api/v1/orders"},
//		Operations: []string{"get", "post", "updateUser"},
//		Tags:       []string{"public", "v1"},
//		PruneComponents: true,
//	}
type FilterOptions struct {
	// Paths specifies which path prefixes to include (e.g., "/users", "/api/v1").
	// Paths are matched using prefix matching, so "/users" matches "/users/{id}".
	// If empty, all paths are included.
	Paths []string

	// Operations specifies which HTTP operations to include (e.g., "get", "post").
	// Can also include specific operation IDs for more precise filtering.
	// Case-insensitive matching is used for HTTP methods.
	// If empty, all operations are included.
	Operations []string

	// Tags specifies which OpenAPI tags to include.
	// Only operations with at least one of these tags will be included.
	// If empty, all tags are included.
	Tags []string

	// PruneComponents removes unused components (schemas, parameters, etc.)
	// from the filtered specification to reduce size.
	// This is useful when creating minimal API specifications.
	// This helps reduce specification size and improves readability
	PruneComponents bool
}

// LoadOptions defines configuration options for creating OpenAx clients.
//
// These options control how OpenAPI specifications are loaded and processed.
type LoadOptions struct {
	// AllowExternalRefs enables loading of external references in OpenAPI specs.
	// When true, allows $ref to external files or URLs.
	// Default: false for security reasons.
	AllowExternalRefs bool

	// Context provides cancellation and deadline control for loading operations.
	// If nil, context.Background() is used.
	Context context.Context
}

// Client provides the main OpenAx functionality for loading, filtering, and validating
// OpenAPI specifications.
//
// Clients are safe for concurrent use and should be reused across operations.
// Create clients using New() or NewWithOptions().
//
// Example:
//
//	client := openax.New()
//	doc, err := client.LoadFromFile("api.yaml")
//	filtered, err := client.Filter(doc, options)
type Client struct {
	loader *openapi3.Loader
}

// New creates a new OpenAx client with default options.
//
// The default configuration enables external references and uses a background context.
// This is suitable for most use cases.
//
// For custom configuration, use NewWithOptions instead.
//
// Example:
//
//	client := openax.New()
//	doc, err := client.LoadFromFile("api.yaml")
func New() *Client {
	return NewWithOptions(LoadOptions{
		AllowExternalRefs: true,
		Context:           context.Background(),
	})
}

// NewWithOptions creates a new OpenAx client with custom options.
//
// This allows fine-grained control over loading behavior, such as disabling
// external references for security or providing a custom context for cancellation.
//
// Example:
//
//	client := openax.NewWithOptions(openax.LoadOptions{
//		AllowExternalRefs: false,
//		Context:           ctx,
//	})
func NewWithOptions(opts LoadOptions) *Client {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return &Client{
		loader: &openapi3.Loader{
			Context:               ctx,
			IsExternalRefsAllowed: opts.AllowExternalRefs,
		},
	}
}

// LoadFromFile loads an OpenAPI specification from a local file.
//
// The file can be in YAML or JSON format. The file path should be absolute
// or relative to the current working directory.
//
// Example:
//
//	doc, err := client.LoadFromFile("api.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
func (c *Client) LoadFromFile(filePath string) (*openapi3.T, error) {
	return c.loader.LoadFromFile(filePath)
}

// LoadFromURL loads an OpenAPI specification from a remote URL.
//
// Supports both HTTP and HTTPS URLs. The response content-type should be
// application/json, application/yaml, or text/yaml.
//
// Example:
//
//	doc, err := client.LoadFromURL("https://api.example.com/openapi.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
func (c *Client) LoadFromURL(urlStr string) (*openapi3.T, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return c.loader.LoadFromURI(u)
}

// LoadFromData loads an OpenAPI specification from raw byte data.
//
// The data should contain a valid OpenAPI specification in YAML or JSON format.
// This is useful when you have the specification content in memory.
//
// Example:
//
//	yamlData := []byte(`openapi: 3.0.0...`)
//	doc, err := client.LoadFromData(yamlData)
//	if err != nil {
//		log.Fatal(err)
//	}
func (c *Client) LoadFromData(data []byte) (*openapi3.T, error) {
	return c.loader.LoadFromData(data)
}

// Validate validates an OpenAPI specification against the OpenAPI 3.x standard.
//
// This checks for structural correctness, required fields, and schema compliance.
// It does not perform filtering - use this to validate specifications before
// or after filtering operations.
//
// Example:
//
//	doc, _ := client.LoadFromFile("api.yaml")
//	if err := client.Validate(doc); err != nil {
//		log.Printf("Validation failed: %v", err)
//	}
func (c *Client) Validate(doc *openapi3.T) error {
	return doc.Validate(c.loader.Context)
}

// Filter applies filtering to an OpenAPI specification based on the provided options.
//
// It returns a new specification containing only the requested paths, operations, and tags,
// along with all referenced components (schemas, parameters, request bodies, responses).
// The original specification is not modified.
//
// Component dependency resolution is handled automatically - all components referenced
// by filtered operations are included in the result, ensuring the filtered specification
// remains valid and complete.
//
// Example:
//
//	filtered, err := client.Filter(doc, openax.FilterOptions{
//		Paths: []string{"/users"},
//		Operations: []string{"get", "post"},
//		Tags: []string{"public"},
//		PruneComponents: true,
//	})
func (c *Client) Filter(doc *openapi3.T, opts FilterOptions) (*openapi3.T, error) {
	return applyFilter(doc, opts)
}

// LoadAndFilter is a convenience method that loads and filters a specification in one call.
//
// This combines loading (from file or URL) and filtering into a single operation.
// The source is automatically detected - URLs starting with http:// or https:// are
// loaded from the network, otherwise treated as file paths.
//
// The specification is validated after loading and before filtering to ensure correctness.
//
// Example:
//
//	// Load and filter from file
//	filtered, err := client.LoadAndFilter("api.yaml", openax.FilterOptions{
//		Tags: []string{"users"},
//	})
//
//	// Load and filter from URL
//	filtered, err := client.LoadAndFilter("https://api.example.com/spec.yaml", opts)
func (c *Client) LoadAndFilter(source string, opts FilterOptions) (*openapi3.T, error) {
	var doc *openapi3.T
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		doc, err = c.LoadFromURL(source)
	} else {
		doc, err = c.LoadFromFile(source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	if err := c.Validate(doc); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}

	return c.Filter(doc, opts)
}

// ValidateOnly loads and validates a specification without filtering.
//
// This is useful for checking if an OpenAPI specification is valid before
// performing other operations. The source is automatically detected as file or URL.
//
// Example:
//
//	// Validate a local file
//	if err := client.ValidateOnly("api.yaml"); err != nil {
//		log.Printf("Invalid spec: %v", err)
//	}
//
//	// Validate a remote spec
//	err := client.ValidateOnly("https://api.example.com/openapi.yaml")
func (c *Client) ValidateOnly(source string) error {
	var doc *openapi3.T
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		doc, err = c.LoadFromURL(source)
	} else {
		doc, err = c.LoadFromFile(source)
	}

	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	return c.Validate(doc)
}
