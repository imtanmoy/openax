package openax

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceLocation_String(t *testing.T) {
	testCases := []struct {
		name     string
		location SourceLocation
		expected string
	}{
		{
			name:     "empty location",
			location: SourceLocation{},
			expected: "unknown location",
		},
		{
			name: "file path only",
			location: SourceLocation{
				FilePath: "/path/to/spec.yaml",
			},
			expected: "/path/to/spec.yaml",
		},
		{
			name: "line and column",
			location: SourceLocation{
				Line:   42,
				Column: 10,
			},
			expected: "line 42, col 10",
		},
		{
			name: "spec path only",
			location: SourceLocation{
				Path: "paths./pet.get",
			},
			expected: "path paths./pet.get",
		},
		{
			name: "complete location",
			location: SourceLocation{
				FilePath: "/path/to/spec.yaml",
				Line:     42,
				Column:   10,
				Path:     "paths./pet.get",
			},
			expected: "/path/to/spec.yaml, line 42, col 10, path paths./pet.get",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.location.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestComponentNotFoundError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := ComponentNotFoundError{
			Name: "Pet",
			Type: "schema",
		}

		assert.Equal(t, "schema not found: Pet", err.Error())
	})

	t.Run("error with context", func(t *testing.T) {
		err := ComponentNotFoundError{
			Name:    "Pet",
			Type:    "schema",
			Context: "operation.requestBody",
		}

		assert.Equal(t, "schema not found: Pet (referenced from operation.requestBody)", err.Error())
	})

	t.Run("error with location", func(t *testing.T) {
		err := ComponentNotFoundError{
			Name: "Pet",
			Type: "schema",
			Location: &SourceLocation{
				FilePath: "spec.yaml",
				Line:     42,
			},
		}

		expected := "schema not found: Pet at spec.yaml, line 42"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := ComponentNotFoundError{
			Name:  "Pet",
			Type:  "schema",
			Cause: cause,
		}

		assert.Equal(t, "schema not found: Pet: underlying error", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("complete error", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := ComponentNotFoundError{
			Name:    "Pet",
			Type:    "schema",
			Context: "operation.requestBody",
			Location: &SourceLocation{
				FilePath: "spec.yaml",
				Line:     42,
				Path:     "components.schemas.Pet",
			},
			Cause: cause,
		}

		expected := "schema not found: Pet (referenced from operation.requestBody) at spec.yaml, line 42, path components.schemas.Pet: underlying error"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestInvalidReferenceError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := InvalidReferenceError{
			Ref:    "#/invalid/reference",
			Reason: "invalid format",
		}

		assert.Equal(t, "invalid reference '#/invalid/reference': invalid format", err.Error())
	})

	t.Run("error with location", func(t *testing.T) {
		err := InvalidReferenceError{
			Ref:    "#/invalid/reference",
			Reason: "invalid format",
			Location: &SourceLocation{
				FilePath: "spec.yaml",
				Line:     10,
			},
		}

		expected := "invalid reference '#/invalid/reference': invalid format at spec.yaml, line 10"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("parsing failed")
		err := InvalidReferenceError{
			Ref:    "#/invalid/reference",
			Reason: "invalid format",
			Cause:  cause,
		}

		assert.Equal(t, "invalid reference '#/invalid/reference': invalid format: parsing failed", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestFilterError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := FilterError{
			Operation: "resolving schema references",
		}

		assert.Equal(t, "error resolving schema references", err.Error())
	})

	t.Run("error with location", func(t *testing.T) {
		err := FilterError{
			Operation: "resolving schema references",
			Location: &SourceLocation{
				Path: "components.schemas.Pet",
			},
		}

		expected := "error resolving schema references at path components.schemas.Pet"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := ComponentNotFoundError{
			Name: "Pet",
			Type: "schema",
		}
		err := FilterError{
			Operation: "resolving schema references",
			Cause:     cause,
		}

		assert.Equal(t, "error resolving schema references: schema not found: Pet", err.Error())
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wraps error with context", func(t *testing.T) {
		originalErr := errors.New("original error")
		location := &SourceLocation{Path: "test.path"}

		wrapped := WrapError(originalErr, "test operation", location)

		var filterErr FilterError
		require.True(t, errors.As(wrapped, &filterErr))
		assert.Equal(t, "test operation", filterErr.Operation)
		assert.Equal(t, location, filterErr.Location)
		assert.Equal(t, originalErr, filterErr.Cause)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := WrapError(nil, "test operation", nil)
		assert.Nil(t, wrapped)
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("error unwrapping works", func(t *testing.T) {
		originalErr := errors.New("root cause")

		componentErr := ComponentNotFoundError{
			Name:  "Pet",
			Type:  "schema",
			Cause: originalErr,
		}

		wrappedErr := WrapError(componentErr, "filtering operation", &SourceLocation{Path: "test"})

		// Test that we can unwrap to the original error
		assert.True(t, errors.Is(wrappedErr, componentErr))
		assert.True(t, errors.Is(wrappedErr, originalErr))
	})

	t.Run("error types can be checked", func(t *testing.T) {
		componentErr := ComponentNotFoundError{
			Name: "Pet",
			Type: "schema",
		}

		wrappedErr := WrapError(componentErr, "filtering operation", nil)

		var filterErr FilterError
		assert.True(t, errors.As(wrappedErr, &filterErr))

		var compErr ComponentNotFoundError
		assert.True(t, errors.As(wrappedErr, &compErr))
		assert.Equal(t, "Pet", compErr.Name)
	})
}
