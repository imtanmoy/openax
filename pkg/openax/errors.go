package openax

import (
	"fmt"
	"strings"
)

// SourceLocation represents a location in a source file or OpenAPI specification.
type SourceLocation struct {
	FilePath string // Path to the source file
	Line     int    // Line number (0-based)
	Column   int    // Column number (0-based)
	Path     string // JSONPath or YAML path within the document (e.g., "paths./pet.get")
}

// String returns a human-readable representation of the source location.
func (sl SourceLocation) String() string {
	var parts []string
	if sl.FilePath != "" {
		parts = append(parts, sl.FilePath)
	}
	if sl.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d", sl.Line))
	}
	if sl.Column > 0 {
		parts = append(parts, fmt.Sprintf("col %d", sl.Column))
	}
	if sl.Path != "" {
		parts = append(parts, fmt.Sprintf("path %s", sl.Path))
	}
	if len(parts) == 0 {
		return "unknown location"
	}
	return strings.Join(parts, ", ")
}

// ComponentNotFoundError indicates that a referenced component was not found.
type ComponentNotFoundError struct {
	Name     string
	Type     string
	Context  string
	Location *SourceLocation // Location where the reference was found
	Cause    error           // Underlying cause of the error
}

func (e ComponentNotFoundError) Error() string {
	var msg string
	if e.Context != "" {
		msg = fmt.Sprintf("%s not found: %s (referenced from %s)", e.Type, e.Name, e.Context)
	} else {
		msg = fmt.Sprintf("%s not found: %s", e.Type, e.Name)
	}

	if e.Location != nil {
		msg = fmt.Sprintf("%s at %s", msg, e.Location.String())
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the error.
func (e ComponentNotFoundError) Unwrap() error {
	return e.Cause
}

// InvalidReferenceError indicates that a reference string is invalid.
type InvalidReferenceError struct {
	Ref      string
	Reason   string
	Location *SourceLocation // Location where the invalid reference was found
	Cause    error           // Underlying cause of the error
}

func (e InvalidReferenceError) Error() string {
	msg := fmt.Sprintf("invalid reference '%s': %s", e.Ref, e.Reason)

	if e.Location != nil {
		msg = fmt.Sprintf("%s at %s", msg, e.Location.String())
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the error.
func (e InvalidReferenceError) Unwrap() error {
	return e.Cause
}

// FilterError represents an error that occurred during the filtering process.
type FilterError struct {
	Operation string          // The operation being performed (e.g., "filtering paths", "resolving schema")
	Location  *SourceLocation // Location in the OpenAPI specification
	Cause     error           // Underlying cause of the error
}

func (e FilterError) Error() string {
	msg := fmt.Sprintf("error %s", e.Operation)

	if e.Location != nil {
		msg = fmt.Sprintf("%s at %s", msg, e.Location.String())
	}

	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause of the error.
func (e FilterError) Unwrap() error {
	return e.Cause
}

// WrapError wraps an error with additional context and location information.
func WrapError(err error, operation string, location *SourceLocation) error {
	if err == nil {
		return nil
	}
	return FilterError{
		Operation: operation,
		Location:  location,
		Cause:     err,
	}
}
