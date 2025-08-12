package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Options struct {
	Paths      []string
	Operations []string
	Tags       []string
}

func Apply(doc *openapi3.T, opts Options) (*openapi3.T, error) {
	filtered := &openapi3.T{
		OpenAPI:      doc.OpenAPI,
		Info:         doc.Info,
		Servers:      doc.Servers,
		ExternalDocs: doc.ExternalDocs,
		Security:     make(openapi3.SecurityRequirements, 0),
		Paths:        &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas:       make(openapi3.Schemas),
			Parameters:    make(openapi3.ParametersMap),
			RequestBodies: make(openapi3.RequestBodies),
			Responses:     make(openapi3.ResponseBodies),
		},
	}

	if doc.Components != nil {
		filtered.Components.Headers = doc.Components.Headers
		filtered.Components.SecuritySchemes = doc.Components.SecuritySchemes
		filtered.Components.Examples = doc.Components.Examples
		filtered.Components.Links = doc.Components.Links
		filtered.Components.Callbacks = doc.Components.Callbacks
	}

	mimeTypes := findAllMimeTypes(doc)
	usedTagNames := make(map[string]bool)

	processedSchemaRefs := make(map[string]bool)
	processedRequestBodyRefs := make(map[string]bool)
	processedParameterRefs := make(map[string]bool)
	processedResponseRefs := make(map[string]bool)

	// Process paths and operations
	for path, pathItem := range doc.Paths.Map() {
		// Include entire path if it's in the paths list
		if len(opts.Paths) > 0 && pathMatchesFilter(path, opts.Paths) {
			filtered.Paths.Set(path, pathItem)

			// Process all operations in this path to collect references and tags
			for _, operation := range pathItem.Operations() {
				if operation != nil {
					err := collectReferencesFromOperation(doc, operation, mimeTypes,
						processedSchemaRefs, processedRequestBodyRefs,
						processedParameterRefs, processedResponseRefs)
					if err != nil {
						return nil, err
					}

					// Collect tags used by this operation
					for _, tag := range operation.Tags {
						usedTagNames[tag] = true
					}
				}
			}
			continue
		}

		// Check for operations that match either operation IDs or tags
		matchedOps := make(map[string]*openapi3.Operation)
		for method, operation := range pathItem.Operations() {
			operationMatches := true

			// Check operation filter (if specified)
			if len(opts.Operations) > 0 {
				operationMatches = slices.Contains(opts.Operations, operation.OperationID) || 
					slices.ContainsFunc(opts.Operations, func(op string) bool {
						return strings.EqualFold(op, method)
					})
			}

			// Check tag filter (if specified) - must match at least one tag
			if len(opts.Tags) > 0 && operationMatches {
				tagMatches := false
				for _, operationTag := range operation.Tags {
					if slices.Contains(opts.Tags, operationTag) {
						tagMatches = true
						break
					}
				}
				operationMatches = operationMatches && tagMatches
			}

			// Include if all specified filters match
			if operationMatches && (len(opts.Operations) > 0 || len(opts.Tags) > 0 || (len(opts.Operations) == 0 && len(opts.Tags) == 0 && len(opts.Paths) == 0)) {
				matchedOps[method] = operation
			}

			// If we matched this operation, process its references and tags
			if op, included := matchedOps[method]; included {
				err := collectReferencesFromOperation(doc, op, mimeTypes,
					processedSchemaRefs, processedRequestBodyRefs,
					processedParameterRefs, processedResponseRefs)
				if err != nil {
					return nil, err
				}

				// Collect tags used by this operation
				for _, tag := range op.Tags {
					usedTagNames[tag] = true
				}
			}
		}

		if len(matchedOps) > 0 {
			pItem := &openapi3.PathItem{}
			for method, operation := range matchedOps {
				pItem.SetOperation(method, operation)
			}
			filtered.Paths.Set(path, pItem)
		}
	}

	// Only include tags that are used by filtered operations
	if len(usedTagNames) > 0 {
		filtered.Tags = make(openapi3.Tags, 0)

		// Find matching tags from the original document
		for _, tag := range doc.Tags {
			if usedTagNames[tag.Name] {
				filtered.Tags = append(filtered.Tags, tag)
			}
		}
	}

	// Process all collected schema references recursively
	for schemaName := range processedSchemaRefs {
		if err := resolveSchemaRefsRecursively(doc, filtered, schemaName, make(map[string]bool), "root"); err != nil {
			return nil, err
		}
	}

	// Process all request body references
	if doc.Components != nil {
		for requestBodyName := range processedRequestBodyRefs {
			requestBody, ok := doc.Components.RequestBodies[requestBodyName]
			if !ok {
				return nil, fmt.Errorf("request body not found: %s", requestBodyName)
			}
			filtered.Components.RequestBodies[requestBodyName] = requestBody
		}

		// Process all parameter references
		for paramName := range processedParameterRefs {
			param, ok := doc.Components.Parameters[paramName]
			if !ok {
				return nil, fmt.Errorf("parameter not found: %s", paramName)
			}
			filtered.Components.Parameters[paramName] = param
		}

		// Process all response references
		for responseName := range processedResponseRefs {
			response, ok := doc.Components.Responses[responseName]
			if !ok {
				return nil, fmt.Errorf("response not found: %s", responseName)
			}
			filtered.Components.Responses[responseName] = response
		}
	}

	return filtered, nil
}

func pathMatchesFilter(path string, pathFilters []string) bool {
	for _, filterPath := range pathFilters {
		if strings.HasPrefix(path, filterPath) {
			return true
		}
	}
	return false
}

// extractRefName extracts the component name from a reference string
func extractRefName(ref string) string {
	refParts := strings.Split(ref, "/")
	return refParts[len(refParts)-1]
}

// validateRef checks if a reference string follows the expected pattern
func validateRef(ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("invalid reference: empty reference")
	}
	if !strings.HasPrefix(ref, "#/components/") {
		return "", fmt.Errorf("invalid reference format: %s", ref)
	}
	return extractRefName(ref), nil
}

// collectReferencesFromOperation extracts all references from an operation and tracks them
func collectReferencesFromOperation(
	doc *openapi3.T,
	operation *openapi3.Operation,
	mimeTypes []string,
	processedSchemaRefs map[string]bool,
	processedRequestBodyRefs map[string]bool,
	processedParameterRefs map[string]bool,
	processedResponseRefs map[string]bool,
) error {
	// Process request body references
	if operation.RequestBody != nil {
		if operation.RequestBody.Ref != "" {
			requestBodyName, err := validateRef(operation.RequestBody.Ref)
			if err != nil {
				return err
			}
			processedRequestBodyRefs[requestBodyName] = true

			// Get the actual request body
			if requestBody, ok := doc.Components.RequestBodies[requestBodyName]; ok {
				// Process content schemas in the request body
				for _, mimeType := range mimeTypes {
					if mediaType := requestBody.Value.Content.Get(mimeType); mediaType != nil {
						if mediaType.Schema != nil {
							if err := extractSchemaReferences(mediaType.Schema, processedSchemaRefs); err != nil {
								return err
							}
						}
					}
				}
			}
		} else if operation.RequestBody.Value != nil {
			// Process inline request body
			for _, mimeType := range mimeTypes {
				if mediaType := operation.RequestBody.Value.Content.Get(mimeType); mediaType != nil {
					if mediaType.Schema != nil {
						if err := extractSchemaReferences(mediaType.Schema, processedSchemaRefs); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// Process parameter references
	for _, param := range operation.Parameters {
		if param.Ref != "" {
			paramName, err := validateRef(param.Ref)
			if err != nil {
				return err
			}
			processedParameterRefs[paramName] = true

			// Get the actual parameter to check its schema
			if parameter, ok := doc.Components.Parameters[paramName]; ok {
				if parameter.Value != nil && parameter.Value.Schema != nil && parameter.Value.Schema.Ref != "" {
					schemaName, err := validateRef(parameter.Value.Schema.Ref)
					if err != nil {
						return err
					}
					processedSchemaRefs[schemaName] = true
				}
			}
		} else if param.Value != nil && param.Value.Schema != nil && param.Value.Schema.Ref != "" {
			schemaName, err := validateRef(param.Value.Schema.Ref)
			if err != nil {
				return err
			}
			processedSchemaRefs[schemaName] = true
		}
	}

	// Process response references
	for _, response := range operation.Responses.Map() {
		if response.Ref != "" {
			responseName, err := validateRef(response.Ref)
			if err != nil {
				return err
			}
			processedResponseRefs[responseName] = true

			// Get the actual response to check its schema
			if responseBody, ok := doc.Components.Responses[responseName]; ok {
				for _, mimeType := range mimeTypes {
					if mediaType := responseBody.Value.Content.Get(mimeType); mediaType != nil {
						if mediaType.Schema != nil {
							if err := extractSchemaReferences(mediaType.Schema, processedSchemaRefs); err != nil {
								return err
							}
						}
					}
				}
			}
		} else if response.Value != nil {
			for _, mimeType := range mimeTypes {
				if mediaType := response.Value.Content.Get(mimeType); mediaType != nil {
					if mediaType.Schema != nil {
						if err := extractSchemaReferences(mediaType.Schema, processedSchemaRefs); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// resolveSchemaRefsRecursively resolves all schema references recursively
func resolveSchemaRefsRecursively(
	doc *openapi3.T,
	filtered *openapi3.T,
	schemaName string,
	processedRefs map[string]bool,
	parentContext string,
) error {
	// Check if already processed to prevent infinite recursion
	if processedRefs[schemaName] {
		return nil
	}
	processedRefs[schemaName] = true

	if doc.Components == nil {
		return fmt.Errorf("no components section found in document")
	}

	schema, ok := doc.Components.Schemas[schemaName]
	if !ok {
		return fmt.Errorf("schema not found: %s (referenced from %s)", schemaName, parentContext)
	}

	// Add to filtered spec
	filtered.Components.Schemas[schemaName] = schema

	// If this schema itself references another schema
	if schema.Ref != "" {
		refName, err := validateRef(schema.Ref)
		if err != nil {
			return fmt.Errorf("%w (in schema %s)", err, schemaName)
		}

		if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs, schemaName); err != nil {
			return err
		}
	}

	// No more processing needed if the schema value is nil
	if schema.Value == nil {
		return nil
	}

	// Process array items
	if schema.Value.Items != nil {
		if schema.Value.Items.Ref != "" {
			refName, err := validateRef(schema.Value.Items.Ref)
			if err != nil {
				return fmt.Errorf("%w (in schema %s.items)", err, schemaName)
			}

			if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs, schemaName+".items"); err != nil {
				return err
			}
		}

		// Also process the items if it has a Value
		if schema.Value.Items.Value != nil {
			// Process array item properties
			if schema.Value.Items.Value.Properties != nil {
				for propName, propSchema := range schema.Value.Items.Value.Properties {
					if propSchema.Ref != "" {
						refName, err := validateRef(propSchema.Ref)
						if err != nil {
							return fmt.Errorf("%w (in schema %s.items.properties.%s)", err, schemaName, propName)
						}

						if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
							fmt.Sprintf("%s.items.properties.%s", schemaName, propName)); err != nil {
							return err
						}
					}

					// Process nested items within item properties
					if propSchema.Value != nil && propSchema.Value.Items != nil && propSchema.Value.Items.Ref != "" {
						refName, err := validateRef(propSchema.Value.Items.Ref)
						if err != nil {
							return fmt.Errorf("%w (in schema %s.items.properties.%s.items)",
								err, schemaName, propName)
						}

						if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
							fmt.Sprintf("%s.items.properties.%s.items", schemaName, propName)); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// Process properties for object schemas
	if schema.Value.Properties != nil {
		for propName, propSchema := range schema.Value.Properties {
			if propSchema.Ref != "" {
				refName, err := validateRef(propSchema.Ref)
				if err != nil {
					return fmt.Errorf("%w (in schema %s.properties.%s)", err, schemaName, propName)
				}

				if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs, schemaName+".properties."+propName); err != nil {
					return err
				}
			}

			// Process nested objects within properties
			if propSchema.Value != nil {
				// Handle arrays of objects in properties
				if propSchema.Value.Items != nil && propSchema.Value.Items.Ref != "" {
					refName, err := validateRef(propSchema.Value.Items.Ref)
					if err != nil {
						return fmt.Errorf("%w (in schema %s.properties.%s.items)", err, schemaName, propName)
					}

					if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
						fmt.Sprintf("%s.properties.%s.items", schemaName, propName)); err != nil {
						return err
					}
				}

				// Handle nested object properties
				if propSchema.Value.Properties != nil {
					for nestedPropName, nestedProp := range propSchema.Value.Properties {
						if nestedProp.Ref != "" {
							refName, err := validateRef(nestedProp.Ref)
							if err != nil {
								return fmt.Errorf("%w (in schema %s.properties.%s.%s)",
									err, schemaName, propName, nestedPropName)
							}

							if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
								fmt.Sprintf("%s.properties.%s.%s", schemaName, propName, nestedPropName)); err != nil {
								return err
							}
						}

						// Process even deeper nested items if they exist
						if nestedProp.Value != nil && nestedProp.Value.Items != nil && nestedProp.Value.Items.Ref != "" {
							refName, err := validateRef(nestedProp.Value.Items.Ref)
							if err != nil {
								return fmt.Errorf("%w (in schema %s.properties.%s.%s.items)",
									err, schemaName, propName, nestedPropName)
							}

							if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
								fmt.Sprintf("%s.properties.%s.%s.items", schemaName, propName, nestedPropName)); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	// Process allOf, oneOf, anyOf schemas
	for i, compositionSchema := range schema.Value.AllOf {
		if compositionSchema.Ref != "" {
			refName, err := validateRef(compositionSchema.Ref)
			if err != nil {
				return fmt.Errorf("%w (in schema %s.allOf[%d])", err, schemaName, i)
			}

			if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
				fmt.Sprintf("%s.allOf[%d]", schemaName, i)); err != nil {
				return err
			}
		}
	}

	for i, compositionSchema := range schema.Value.OneOf {
		if compositionSchema.Ref != "" {
			refName, err := validateRef(compositionSchema.Ref)
			if err != nil {
				return fmt.Errorf("%w (in schema %s.oneOf[%d])", err, schemaName, i)
			}

			if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
				fmt.Sprintf("%s.oneOf[%d]", schemaName, i)); err != nil {
				return err
			}
		}
	}

	for i, compositionSchema := range schema.Value.AnyOf {
		if compositionSchema.Ref != "" {
			refName, err := validateRef(compositionSchema.Ref)
			if err != nil {
				return fmt.Errorf("%w (in schema %s.anyOf[%d])", err, schemaName, i)
			}

			if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
				fmt.Sprintf("%s.anyOf[%d]", schemaName, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// findAllMimeTypes extracts all MIME types from an OpenAPI document
func findAllMimeTypes(doc *openapi3.T) []string {
	if doc == nil || doc.Paths == nil {
		return []string{}
	}

	mimeTypeSet := make(map[string]struct{})

	// Default MIME types to always include
	defaults := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"application/xml",
		"text/plain",
	}

	for _, mt := range defaults {
		mimeTypeSet[mt] = struct{}{}
	}

	// Search for MIME types in request bodies
	for _, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		for _, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			// Check request body
			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				for mt := range operation.RequestBody.Value.Content {
					mimeTypeSet[mt] = struct{}{}
				}
			}

			// Check responses
			if operation.Responses != nil {
				for _, response := range operation.Responses.Map() {
					if response != nil && response.Value != nil {
						for mt := range response.Value.Content {
							mimeTypeSet[mt] = struct{}{}
						}
					}
				}
			}
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(mimeTypeSet))
	for mt := range mimeTypeSet {
		result = append(result, mt)
	}

	return result
}

// extractSchemaReferences recursively extracts all schema references from a schema
func extractSchemaReferences(schema *openapi3.SchemaRef, processedSchemaRefs map[string]bool) error {
	if schema == nil {
		return nil
	}

	// Direct reference
	if schema.Ref != "" {
		schemaName, err := validateRef(schema.Ref)
		if err != nil {
			return err
		}
		processedSchemaRefs[schemaName] = true
	}

	// Process schema value
	if schema.Value != nil {
		// Array items
		if schema.Value.Items != nil {
			if err := extractSchemaReferences(schema.Value.Items, processedSchemaRefs); err != nil {
				return err
			}
		}

		// Object properties
		for _, propSchema := range schema.Value.Properties {
			if err := extractSchemaReferences(propSchema, processedSchemaRefs); err != nil {
				return err
			}
		}

		// Composition schemas
		for _, compositionSchema := range schema.Value.AllOf {
			if err := extractSchemaReferences(compositionSchema, processedSchemaRefs); err != nil {
				return err
			}
		}
		for _, compositionSchema := range schema.Value.OneOf {
			if err := extractSchemaReferences(compositionSchema, processedSchemaRefs); err != nil {
				return err
			}
		}
		for _, compositionSchema := range schema.Value.AnyOf {
			if err := extractSchemaReferences(compositionSchema, processedSchemaRefs); err != nil {
				return err
			}
		}
		if schema.Value.Not != nil {
			if err := extractSchemaReferences(schema.Value.Not, processedSchemaRefs); err != nil {
				return err
			}
		}
	}

	return nil
}