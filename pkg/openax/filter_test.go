package openax

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestExtractRefName(t *testing.T) {
	testCases := []struct {
		ref      string
		expected string
	}{
		{"#/components/schemas/User", "User"},
		{"#/components/parameters/UserId", "UserId"},
		{"#/components/responses/ErrorResponse", "ErrorResponse"},
		{"#/components/requestBodies/UserRequest", "UserRequest"},
	}

	for _, tc := range testCases {
		t.Run(tc.ref, func(t *testing.T) {
			result := extractRefName(tc.ref)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestValidateRef(t *testing.T) {
	testCases := []struct {
		name        string
		ref         string
		expected    string
		expectError bool
	}{
		{
			name:        "valid schema ref",
			ref:         "#/components/schemas/User",
			expected:    "User",
			expectError: false,
		},
		{
			name:        "valid parameter ref",
			ref:         "#/components/parameters/UserId",
			expected:    "UserId",
			expectError: false,
		},
		{
			name:        "empty ref",
			ref:         "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid ref format",
			ref:         "User",
			expected:    "",
			expectError: true,
		},
		{
			name:        "wrong prefix",
			ref:         "#/definitions/User",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validateRef(tc.ref)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestPathMatchesFilter(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		filters     []string
		expected    bool
	}{
		{
			name:     "exact match",
			path:     "/users",
			filters:  []string{"/users"},
			expected: true,
		},
		{
			name:     "prefix match",
			path:     "/users/123",
			filters:  []string{"/users"},
			expected: true,
		},
		{
			name:     "multiple filters - match first",
			path:     "/users",
			filters:  []string{"/users", "/posts"},
			expected: true,
		},
		{
			name:     "multiple filters - match second",
			path:     "/posts",
			filters:  []string{"/users", "/posts"},
			expected: true,
		},
		{
			name:     "no match",
			path:     "/admin",
			filters:  []string{"/users", "/posts"},
			expected: false,
		},
		{
			name:     "empty filters",
			path:     "/users",
			filters:  []string{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pathMatchesFilter(tc.path, tc.filters)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestFindAllMimeTypes(t *testing.T) {
	// Create a minimal OpenAPI doc for testing
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	// Add a path with operations that have different content types
	pathItem := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
		Post: &openapi3.Operation{
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{},
						"application/xml":  &openapi3.MediaType{},
					},
				},
			},
			Responses: &openapi3.Responses{},
		},
	}

	// Add response with custom content type
	description := "OK"
	response := &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &description,
			Content: openapi3.Content{
				"application/custom": &openapi3.MediaType{},
			},
		},
	}
	pathItem.Get.Responses.Set("200", response)

	doc.Paths.Set("/test", pathItem)

	mimeTypes := findAllMimeTypes(doc)

	// Should include defaults plus custom types
	expectedTypes := map[string]bool{
		"application/json":                   true,
		"application/x-www-form-urlencoded": true,
		"multipart/form-data":               true,
		"application/xml":                   true,
		"text/plain":                        true,
		"application/custom":                true, // Custom type from response
	}

	found := make(map[string]bool)
	for _, mt := range mimeTypes {
		found[mt] = true
	}

	for expectedType := range expectedTypes {
		if !found[expectedType] {
			t.Errorf("Expected MIME type %s not found", expectedType)
		}
	}
}

func TestExtractSchemaReferences(t *testing.T) {
	refs := make(map[string]bool)

	// Test direct reference
	schemaRef := &openapi3.SchemaRef{
		Ref: "#/components/schemas/User",
	}

	err := extractSchemaReferences(schemaRef, refs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !refs["User"] {
		t.Error("Expected User reference to be extracted")
	}

	// Test nested reference in array items
	refs = make(map[string]bool)
	arrayType := &openapi3.Types{"array"}
	arraySchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: arrayType,
			Items: &openapi3.SchemaRef{
				Ref: "#/components/schemas/Post",
			},
		},
	}

	err = extractSchemaReferences(arraySchema, refs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !refs["Post"] {
		t.Error("Expected Post reference to be extracted from array items")
	}

	// Test nil schema
	refs = make(map[string]bool)
	err = extractSchemaReferences(nil, refs)
	if err != nil {
		t.Errorf("Unexpected error for nil schema: %v", err)
	}

	if len(refs) != 0 {
		t.Error("Expected no references for nil schema")
	}
}