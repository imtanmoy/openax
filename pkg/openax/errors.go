package openax

import "fmt"

// ComponentNotFoundError indicates that a referenced component was not found.
type ComponentNotFoundError struct {
	Name    string
	Type    string
	Context string
}

func (e ComponentNotFoundError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s not found: %s (referenced from %s)", e.Type, e.Name, e.Context)
	}
	return fmt.Sprintf("%s not found: %s", e.Type, e.Name)
}

// InvalidReferenceError indicates that a reference string is invalid.
type InvalidReferenceError struct {
	Ref    string
	Reason string
}

func (e InvalidReferenceError) Error() string {
	return fmt.Sprintf("invalid reference '%s': %s", e.Ref, e.Reason)
}
