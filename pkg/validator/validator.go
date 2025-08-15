// Package validator provides OpenAPI specification validation utilities.
//
// This package offers comprehensive validation of OpenAPI 3.x specifications,
// checking for structural correctness, required fields, and schema compliance.
// It's designed to be used both standalone and as part of larger processing pipelines.
//
// # Basic Usage
//
//	validator := validator.New()
//	err := validator.ValidateFile("api.yaml")
//	if err != nil {
//		fmt.Printf("Validation failed: %v\n", err)
//	}
//
// # Advanced Usage
//
//	validator := validator.NewWithContext(ctx)
//	doc := loadOpenAPIDoc()
//	if err := validator.Validate(doc); err != nil {
//		// Handle validation errors
//	}
//
// The validator provides detailed error messages to help identify and fix
// specification issues quickly.
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
