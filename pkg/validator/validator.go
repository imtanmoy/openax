// Package validator provides OpenAPI specification validation utilities.
package validator

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
)

// Validator provides validation functionality for OpenAPI specifications.
type Validator struct {
	ctx context.Context
}

// New creates a new validator with default context.
func New() *Validator {
	return NewWithContext(context.Background())
}

// NewWithContext creates a new validator with the given context.
func NewWithContext(ctx context.Context) *Validator {
	return &Validator{ctx: ctx}
}

// Validate validates an OpenAPI specification.
func (v *Validator) Validate(doc *openapi3.T) error {
	return doc.Validate(v.ctx)
}

// ValidateWithOptions validates with custom validation options.
func (v *Validator) ValidateWithOptions(doc *openapi3.T, opts ...openapi3.ValidationOption) error {
	return doc.Validate(v.ctx, opts...)
}