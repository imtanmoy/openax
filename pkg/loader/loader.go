// Package loader provides utilities for loading OpenAPI specifications from various sources.
package loader

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Loader wraps the OpenAPI loader with additional functionality.
type Loader struct {
	loader *openapi3.Loader
}

// Options defines loading options.
type Options struct {
	AllowExternalRefs bool
	Context           context.Context
}

// New creates a new loader with default options.
func New() *Loader {
	return NewWithOptions(Options{
		AllowExternalRefs: true,
		Context:           context.Background(),
	})
}

// NewWithOptions creates a new loader with custom options.
func NewWithOptions(opts Options) *Loader {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return &Loader{
		loader: &openapi3.Loader{
			Context:               ctx,
			IsExternalRefsAllowed: opts.AllowExternalRefs,
		},
	}
}

// LoadFromFile loads an OpenAPI specification from a local file.
func (l *Loader) LoadFromFile(filePath string) (*openapi3.T, error) {
	return l.loader.LoadFromFile(filePath)
}

// LoadFromURL loads an OpenAPI specification from a URL.
func (l *Loader) LoadFromURL(urlStr string) (*openapi3.T, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return l.loader.LoadFromURI(u)
}

// LoadFromData loads an OpenAPI specification from raw data.
func (l *Loader) LoadFromData(data []byte) (*openapi3.T, error) {
	return l.loader.LoadFromData(data)
}

// LoadFromSource automatically detects and loads from file or URL.
func (l *Loader) LoadFromSource(source string) (*openapi3.T, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return l.LoadFromURL(source)
	}
	return l.LoadFromFile(source)
}
