package validator_test

import (
	"context"
	"testing"

	"github.com/imtanmoy/openax/pkg/loader"
	"github.com/imtanmoy/openax/pkg/validator"
)

func TestNew(t *testing.T) {
	v := validator.New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewWithContext(t *testing.T) {
	ctx := context.Background()
	v := validator.NewWithContext(ctx)
	if v == nil {
		t.Fatal("NewWithContext() returned nil")
	}
}

func TestValidate(t *testing.T) {
	l := loader.New()
	v := validator.New()
	
	testCases := []struct {
		name        string
		specFile    string
		expectError bool
	}{
		{
			name:        "valid simple spec",
			specFile:    "../../testdata/specs/simple.yaml",
			expectError: false,
		},
		{
			name:        "valid petstore spec",
			specFile:    "../../testdata/specs/petstore.yaml",
			expectError: false,
		},
		{
			name:        "invalid spec",
			specFile:    "../../testdata/specs/invalid.yaml",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := l.LoadFromFile(tc.specFile)
			if err != nil {
				t.Fatalf("Failed to load spec: %v", err)
			}
			
			err = v.Validate(doc)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestValidateWithOptions(t *testing.T) {
	l := loader.New()
	v := validator.New()
	
	doc, err := l.LoadFromFile("../../testdata/specs/simple.yaml")
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}
	
	// Test with no options (should pass)
	err = v.ValidateWithOptions(doc)
	if err != nil {
		t.Errorf("Validation with no options failed: %v", err)
	}
	
	// The actual validation options would depend on what's available
	// in the kin-openapi library. This is a basic test structure.
}