// Package openax provides OpenAPI specification filtering and validation capabilities.
// It can be used both as a library and as a CLI tool to filter OpenAPI specs
// by paths, operations, and tags while resolving component dependencies.
package openax

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FilterOptions defines the filtering criteria for OpenAPI specifications.
type FilterOptions struct {
	// Paths specifies which paths to include (e.g., "/users", "/orders")
	// If empty, all paths are considered
	Paths []string

	// Operations specifies which HTTP operations to include (e.g., "get", "post")
	// Can also include operation IDs
	// If empty, all operations are considered
	Operations []string

	// Tags specifies which tags to include
	// If empty, all tags are considered
	Tags []string
}

// LoadOptions defines options for loading OpenAPI specifications.
type LoadOptions struct {
	// AllowExternalRefs enables loading of external references
	AllowExternalRefs bool

	// Context for the loading operation
	Context context.Context
}

// Client provides the main interface for OpenAPI operations.
type Client struct {
	loader *openapi3.Loader
}

// New creates a new OpenAx client with default options.
func New() *Client {
	return NewWithOptions(LoadOptions{
		AllowExternalRefs: true,
		Context:           context.Background(),
	})
}

// NewWithOptions creates a new OpenAx client with custom options.
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
func (c *Client) LoadFromFile(filePath string) (*openapi3.T, error) {
	return c.loader.LoadFromFile(filePath)
}

// LoadFromURL loads an OpenAPI specification from a URL.
func (c *Client) LoadFromURL(urlStr string) (*openapi3.T, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return c.loader.LoadFromURI(u)
}

// LoadFromData loads an OpenAPI specification from raw data.
func (c *Client) LoadFromData(data []byte) (*openapi3.T, error) {
	return c.loader.LoadFromData(data)
}

// Validate validates an OpenAPI specification.
func (c *Client) Validate(doc *openapi3.T) error {
	return doc.Validate(c.loader.Context)
}

// Filter applies filtering to an OpenAPI specification based on the provided options.
// It returns a new specification containing only the requested paths, operations, and tags,
// along with all referenced components (schemas, parameters, request bodies, responses).
func (c *Client) Filter(doc *openapi3.T, opts FilterOptions) (*openapi3.T, error) {
	return applyFilter(doc, opts)
}

// LoadAndFilter is a convenience method that loads and filters a specification in one call.
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
