package validator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/imtanmoy/openax/pkg/loader"
	"github.com/imtanmoy/openax/pkg/validator"
)

func TestNew(t *testing.T) {
	v := validator.New()
	require.NotNil(t, v, "New() should not return nil")
}

func TestNewWithContext(t *testing.T) {
	ctx := context.Background()
	v := validator.NewWithContext(ctx)
	require.NotNil(t, v, "NewWithContext() should not return nil")
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
			require.NoError(t, err, "Failed to load spec for %s", tc.name)
			
			err = v.Validate(doc)
			
			if tc.expectError {
				assert.Error(t, err, "Expected validation error for %s", tc.name)
			} else {
				assert.NoError(t, err, "Unexpected validation error for %s", tc.name)
			}
		})
	}
}

func TestValidateWithOptions(t *testing.T) {
	l := loader.New()
	v := validator.New()
	
	doc, err := l.LoadFromFile("../../testdata/specs/simple.yaml")
	require.NoError(t, err, "Failed to load spec")
	
	// Test with no options (should pass)
	err = v.ValidateWithOptions(doc)
	assert.NoError(t, err, "Validation with no options should not fail")
	
	// The actual validation options would depend on what's available
	// in the kin-openapi library. This is a basic test structure.
}