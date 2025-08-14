package openax

import (
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// createLocation creates a SourceLocation for the given spec path
func createLocation(specPath string) *SourceLocation {
	return &SourceLocation{
		Path: specPath,
	}
}

// applyFilter applies filtering to an OpenAPI specification based on the provided options.
func applyFilter(doc *openapi3.T, opts FilterOptions) (*openapi3.T, error) {
	filtered := createFilteredSpec(doc)
	mimeTypes := findAllMimeTypes(doc)
	usedTagNames := make(map[string]bool)

	processedRefs := &ProcessedRefs{
		Schemas:       make(map[string]bool),
		RequestBodies: make(map[string]bool),
		Parameters:    make(map[string]bool),
		Responses:     make(map[string]bool),
	}

	// Process paths and operations
	if err := processPathsAndOperations(doc, filtered, opts, mimeTypes, usedTagNames, processedRefs); err != nil {
		return nil, err
	}

	// Process tags
	processUsedTags(doc, filtered, usedTagNames)

	// Resolve all collected references
	if err := resolveAllReferences(doc, filtered, processedRefs); err != nil {
		return nil, err
	}

	// Prune unused components if enabled
	if opts.PruneComponents {
		pruneUnusedComponents(filtered, processedRefs)
	}

	return filtered, nil
}

// pruneUnusedComponents removes components that are not referenced by the filtered spec
func pruneUnusedComponents(filtered *openapi3.T, processedRefs *ProcessedRefs) {
	if filtered.Components == nil {
		return
	}

	// Create sets of all components and used components
	usedComponents := &ComponentUsage{
		Schemas:       processedRefs.Schemas,
		Parameters:    processedRefs.Parameters,
		RequestBodies: processedRefs.RequestBodies,
		Responses:     processedRefs.Responses,
	}

	// Recursively find all transitively used components
	findTransitivelyUsedComponents(filtered, usedComponents)

	// Remove unused schemas
	for schemaName := range filtered.Components.Schemas {
		if !usedComponents.Schemas[schemaName] {
			delete(filtered.Components.Schemas, schemaName)
		}
	}

	// Remove unused parameters
	for paramName := range filtered.Components.Parameters {
		if !usedComponents.Parameters[paramName] {
			delete(filtered.Components.Parameters, paramName)
		}
	}

	// Remove unused request bodies
	for rbName := range filtered.Components.RequestBodies {
		if !usedComponents.RequestBodies[rbName] {
			delete(filtered.Components.RequestBodies, rbName)
		}
	}

	// Remove unused responses
	for respName := range filtered.Components.Responses {
		if !usedComponents.Responses[respName] {
			delete(filtered.Components.Responses, respName)
		}
	}
}

// ComponentUsage tracks which components are used
type ComponentUsage struct {
	Schemas       map[string]bool
	Parameters    map[string]bool
	RequestBodies map[string]bool
	Responses     map[string]bool
}

// findTransitivelyUsedComponents finds all components that are transitively referenced
func findTransitivelyUsedComponents(filtered *openapi3.T, usage *ComponentUsage) {
	// Keep iterating until no new components are found
	changed := true
	for changed {
		changed = false

		// Check schemas for transitive references
		for schemaName := range usage.Schemas {
			if schema, exists := filtered.Components.Schemas[schemaName]; exists && schema != nil {
				refs := make(map[string]bool)
				if err := extractSchemaReferences(schema, refs); err == nil {
					for refName := range refs {
						if !usage.Schemas[refName] {
							usage.Schemas[refName] = true
							changed = true
						}
					}
				}
			}
		}

		// Check parameters for schema references
		for paramName := range usage.Parameters {
			if param, exists := filtered.Components.Parameters[paramName]; exists && param.Value != nil && param.Value.Schema != nil {
				refs := make(map[string]bool)
				if err := extractSchemaReferences(param.Value.Schema, refs); err == nil {
					for refName := range refs {
						if !usage.Schemas[refName] {
							usage.Schemas[refName] = true
							changed = true
						}
					}
				}
			}
		}

		// Check request bodies for schema references
		for rbName := range usage.RequestBodies {
			if rb, exists := filtered.Components.RequestBodies[rbName]; exists && rb.Value != nil {
				for _, mediaType := range rb.Value.Content {
					if mediaType.Schema != nil {
						refs := make(map[string]bool)
						if err := extractSchemaReferences(mediaType.Schema, refs); err == nil {
							for refName := range refs {
								if !usage.Schemas[refName] {
									usage.Schemas[refName] = true
									changed = true
								}
							}
						}
					}
				}
			}
		}

		// Check responses for schema references
		for respName := range usage.Responses {
			if resp, exists := filtered.Components.Responses[respName]; exists && resp.Value != nil {
				for _, mediaType := range resp.Value.Content {
					if mediaType.Schema != nil {
						refs := make(map[string]bool)
						if err := extractSchemaReferences(mediaType.Schema, refs); err == nil {
							for refName := range refs {
								if !usage.Schemas[refName] {
									usage.Schemas[refName] = true
									changed = true
								}
							}
						}
					}
				}
			}
		}
	}
}

// ProcessedRefs holds all processed reference maps
type ProcessedRefs struct {
	Schemas       map[string]bool
	RequestBodies map[string]bool
	Parameters    map[string]bool
	Responses     map[string]bool
}

// createFilteredSpec creates the initial filtered OpenAPI spec structure
func createFilteredSpec(doc *openapi3.T) *openapi3.T {
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

	return filtered
}

// processPathsAndOperations processes all paths and operations based on filter options
func processPathsAndOperations(doc *openapi3.T, filtered *openapi3.T, opts FilterOptions, mimeTypes []string, usedTagNames map[string]bool, processedRefs *ProcessedRefs) error {
	for path, pathItem := range doc.Paths.Map() {
		// Include entire path if it's in the paths list
		if len(opts.Paths) > 0 && pathMatchesFilter(path, opts.Paths) {
			filtered.Paths.Set(path, pathItem)
			if err := processAllOperationsInPath(doc, pathItem, mimeTypes, usedTagNames, processedRefs); err != nil {
				return err
			}
			continue
		}

		// Check for operations that match filters
		matchedOps, err := findMatchingOperations(doc, pathItem, opts, mimeTypes, usedTagNames, processedRefs)
		if err != nil {
			return err
		}

		if len(matchedOps) > 0 {
			pItem := &openapi3.PathItem{}
			for method, operation := range matchedOps {
				pItem.SetOperation(method, operation)
			}
			filtered.Paths.Set(path, pItem)
		}
	}
	return nil
}

// processAllOperationsInPath processes all operations in a path item
func processAllOperationsInPath(doc *openapi3.T, pathItem *openapi3.PathItem, mimeTypes []string, usedTagNames map[string]bool, processedRefs *ProcessedRefs) error {
	for _, operation := range pathItem.Operations() {
		if operation != nil {
			err := collectReferencesFromOperation(doc, operation, mimeTypes,
				processedRefs.Schemas, processedRefs.RequestBodies,
				processedRefs.Parameters, processedRefs.Responses)
			if err != nil {
				return err
			}

			// Collect tags used by this operation
			for _, tag := range operation.Tags {
				usedTagNames[tag] = true
			}
		}
	}
	return nil
}

// findMatchingOperations finds operations that match the filter criteria
func findMatchingOperations(doc *openapi3.T, pathItem *openapi3.PathItem, opts FilterOptions, mimeTypes []string, usedTagNames map[string]bool, processedRefs *ProcessedRefs) (map[string]*openapi3.Operation, error) {
	matchedOps := make(map[string]*openapi3.Operation)

	for method, operation := range pathItem.Operations() {
		if operationMatches := checkOperationMatches(operation, method, opts); operationMatches {
			matchedOps[method] = operation

			// Process references and tags for matched operation
			err := collectReferencesFromOperation(doc, operation, mimeTypes,
				processedRefs.Schemas, processedRefs.RequestBodies,
				processedRefs.Parameters, processedRefs.Responses)
			if err != nil {
				return nil, err
			}

			// Collect tags used by this operation
			for _, tag := range operation.Tags {
				usedTagNames[tag] = true
			}
		}
	}

	return matchedOps, nil
}

// checkOperationMatches checks if an operation matches the filter criteria
func checkOperationMatches(operation *openapi3.Operation, method string, opts FilterOptions) bool {
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
	return operationMatches && (len(opts.Operations) > 0 || len(opts.Tags) > 0 || (len(opts.Operations) == 0 && len(opts.Tags) == 0 && len(opts.Paths) == 0))
}

// processUsedTags processes tags that are used by filtered operations
func processUsedTags(doc *openapi3.T, filtered *openapi3.T, usedTagNames map[string]bool) {
	if len(usedTagNames) > 0 {
		filtered.Tags = make(openapi3.Tags, 0)

		// Find matching tags from the original document
		for _, tag := range doc.Tags {
			if usedTagNames[tag.Name] {
				filtered.Tags = append(filtered.Tags, tag)
			}
		}
	}
}

// resolveAllReferences resolves all collected references
func resolveAllReferences(doc *openapi3.T, filtered *openapi3.T, processedRefs *ProcessedRefs) error {
	// Process all collected schema references recursively
	for schemaName := range processedRefs.Schemas {
		if err := resolveSchemaRefsRecursively(doc, filtered, schemaName, make(map[string]bool), "root"); err != nil {
			return err
		}
	}

	// Process all other references
	if doc.Components != nil {
		if err := resolveRequestBodyRefs(doc, filtered, processedRefs.RequestBodies); err != nil {
			return err
		}
		if err := resolveParameterRefs(doc, filtered, processedRefs.Parameters); err != nil {
			return err
		}
		if err := resolveResponseRefs(doc, filtered, processedRefs.Responses); err != nil {
			return err
		}
	}

	return nil
}

// resolveRequestBodyRefs resolves request body references
func resolveRequestBodyRefs(doc *openapi3.T, filtered *openapi3.T, requestBodyRefs map[string]bool) error {
	for requestBodyName := range requestBodyRefs {
		requestBody, ok := doc.Components.RequestBodies[requestBodyName]
		if !ok {
			return &ComponentNotFoundError{Name: requestBodyName, Type: "request body"}
		}
		filtered.Components.RequestBodies[requestBodyName] = requestBody
	}
	return nil
}

// resolveParameterRefs resolves parameter references
func resolveParameterRefs(doc *openapi3.T, filtered *openapi3.T, parameterRefs map[string]bool) error {
	for paramName := range parameterRefs {
		param, ok := doc.Components.Parameters[paramName]
		if !ok {
			return &ComponentNotFoundError{Name: paramName, Type: "parameter"}
		}
		filtered.Components.Parameters[paramName] = param
	}
	return nil
}

// resolveResponseRefs resolves response references
func resolveResponseRefs(doc *openapi3.T, filtered *openapi3.T, responseRefs map[string]bool) error {
	for responseName := range responseRefs {
		response, ok := doc.Components.Responses[responseName]
		if !ok {
			return &ComponentNotFoundError{Name: responseName, Type: "response"}
		}
		filtered.Components.Responses[responseName] = response
	}
	return nil
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
func validateRef(ref string, location *SourceLocation) (string, error) {
	if ref == "" {
		return "", InvalidReferenceError{
			Ref:      ref,
			Reason:   "empty reference",
			Location: location,
		}
	}
	if !strings.HasPrefix(ref, "#/components/") {
		return "", InvalidReferenceError{
			Ref:      ref,
			Reason:   "invalid format",
			Location: location,
		}
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
	if err := processOperationRequestBody(doc, operation, mimeTypes, processedSchemaRefs, processedRequestBodyRefs); err != nil {
		return err
	}

	// Process parameter references
	if err := processOperationParameters(doc, operation, processedSchemaRefs, processedParameterRefs); err != nil {
		return err
	}

	// Process response references
	if err := processOperationResponses(doc, operation, mimeTypes, processedSchemaRefs, processedResponseRefs); err != nil {
		return err
	}

	return nil
}

// processOperationRequestBody processes request body references in an operation
func processOperationRequestBody(doc *openapi3.T, operation *openapi3.Operation, mimeTypes []string, processedSchemaRefs map[string]bool, processedRequestBodyRefs map[string]bool) error {
	if operation.RequestBody == nil {
		return nil
	}

	if operation.RequestBody.Ref != "" {
		requestBodyName, err := validateRef(operation.RequestBody.Ref, createLocation("requestBody"))
		if err != nil {
			return err
		}
		processedRequestBodyRefs[requestBodyName] = true

		// Get the actual request body
		if requestBody, ok := doc.Components.RequestBodies[requestBodyName]; ok {
			return processContentSchemas(requestBody.Value.Content, mimeTypes, processedSchemaRefs)
		}
	} else if operation.RequestBody.Value != nil {
		// Process inline request body
		return processContentSchemas(operation.RequestBody.Value.Content, mimeTypes, processedSchemaRefs)
	}

	return nil
}

// processOperationParameters processes parameter references in an operation
func processOperationParameters(doc *openapi3.T, operation *openapi3.Operation, processedSchemaRefs map[string]bool, processedParameterRefs map[string]bool) error {
	for _, param := range operation.Parameters {
		if param.Ref != "" {
			paramName, err := validateRef(param.Ref, createLocation("parameter"))
			if err != nil {
				return err
			}
			processedParameterRefs[paramName] = true

			// Get the actual parameter to check its schema
			if parameter, ok := doc.Components.Parameters[paramName]; ok {
				if parameter.Value != nil && parameter.Value.Schema != nil && parameter.Value.Schema.Ref != "" {
					schemaName, err := validateRef(parameter.Value.Schema.Ref, createLocation("parameter.schema"))
					if err != nil {
						return err
					}
					processedSchemaRefs[schemaName] = true
				}
			}
		} else if param.Value != nil && param.Value.Schema != nil && param.Value.Schema.Ref != "" {
			schemaName, err := validateRef(param.Value.Schema.Ref, createLocation("parameter.schema"))
			if err != nil {
				return err
			}
			processedSchemaRefs[schemaName] = true
		}
	}
	return nil
}

// processOperationResponses processes response references in an operation
func processOperationResponses(doc *openapi3.T, operation *openapi3.Operation, mimeTypes []string, processedSchemaRefs map[string]bool, processedResponseRefs map[string]bool) error {
	for _, response := range operation.Responses.Map() {
		if response.Ref != "" {
			responseName, err := validateRef(response.Ref, createLocation("response"))
			if err != nil {
				return err
			}
			processedResponseRefs[responseName] = true

			// Get the actual response to check its schema
			if responseBody, ok := doc.Components.Responses[responseName]; ok {
				if err := processContentSchemas(responseBody.Value.Content, mimeTypes, processedSchemaRefs); err != nil {
					return err
				}
			}
		} else if response.Value != nil {
			if err := processContentSchemas(response.Value.Content, mimeTypes, processedSchemaRefs); err != nil {
				return err
			}
		}
	}
	return nil
}

// processContentSchemas processes schemas in content for different MIME types
func processContentSchemas(content openapi3.Content, mimeTypes []string, processedSchemaRefs map[string]bool) error {
	for _, mimeType := range mimeTypes {
		if mediaType := content.Get(mimeType); mediaType != nil {
			if mediaType.Schema != nil {
				if err := extractSchemaReferences(mediaType.Schema, processedSchemaRefs); err != nil {
					return err
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
		return &ComponentNotFoundError{Name: "components", Type: "section"}
	}

	schema, ok := doc.Components.Schemas[schemaName]
	if !ok {
		return &ComponentNotFoundError{Name: schemaName, Type: "schema", Context: parentContext}
	}

	// Add to filtered spec
	filtered.Components.Schemas[schemaName] = schema

	// If this schema itself references another schema
	if schema.Ref != "" {
		refName, err := validateRef(schema.Ref, createLocation(fmt.Sprintf("schema.%s", schemaName)))
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

	// Process schema components
	if err := processSchemaItems(doc, filtered, schema, schemaName, processedRefs); err != nil {
		return err
	}

	if err := processSchemaProperties(doc, filtered, schema, schemaName, processedRefs); err != nil {
		return err
	}

	if err := processCompositionSchemas(doc, filtered, schema, schemaName, processedRefs); err != nil {
		return err
	}

	return nil
}

// processSchemaItems processes array items in a schema
func processSchemaItems(doc *openapi3.T, filtered *openapi3.T, schema *openapi3.SchemaRef, schemaName string, processedRefs map[string]bool) error {
	if schema.Value.Items == nil {
		return nil
	}

	if schema.Value.Items.Ref != "" {
		refName, err := validateRef(schema.Value.Items.Ref, createLocation(fmt.Sprintf("schema.%s.items", schemaName)))
		if err != nil {
			return fmt.Errorf("%w (in schema %s.items)", err, schemaName)
		}

		if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs, schemaName+".items"); err != nil {
			return err
		}
	}

	// Also process the items if it has a Value
	if schema.Value.Items.Value != nil && schema.Value.Items.Value.Properties != nil {
		return processItemProperties(doc, filtered, schema, schemaName, processedRefs)
	}

	return nil
}

// processItemProperties processes properties within array items
func processItemProperties(doc *openapi3.T, filtered *openapi3.T, schema *openapi3.SchemaRef, schemaName string, processedRefs map[string]bool) error {
	for propName, propSchema := range schema.Value.Items.Value.Properties {
		if propSchema.Ref != "" {
			refName, err := validateRef(propSchema.Ref, createLocation(fmt.Sprintf("schema.%s.items.properties.%s", schemaName, propName)))
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
			refName, err := validateRef(propSchema.Value.Items.Ref, createLocation(fmt.Sprintf("schema.%s.items.properties.%s.items", schemaName, propName)))
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
	return nil
}

// processSchemaProperties processes object properties in a schema
func processSchemaProperties(doc *openapi3.T, filtered *openapi3.T, schema *openapi3.SchemaRef, schemaName string, processedRefs map[string]bool) error {
	if schema.Value.Properties == nil {
		return nil
	}

	for propName, propSchema := range schema.Value.Properties {
		if err := processPropertyRef(doc, filtered, propSchema, schemaName, propName, processedRefs); err != nil {
			return err
		}

		if err := processNestedPropertyObjects(doc, filtered, propSchema, schemaName, propName, processedRefs); err != nil {
			return err
		}
	}
	return nil
}

// processPropertyRef processes a property reference
func processPropertyRef(doc *openapi3.T, filtered *openapi3.T, propSchema *openapi3.SchemaRef, schemaName, propName string, processedRefs map[string]bool) error {
	if propSchema.Ref != "" {
		refName, err := validateRef(propSchema.Ref, createLocation(fmt.Sprintf("schema.%s.properties.%s", schemaName, propName)))
		if err != nil {
			return fmt.Errorf("%w (in schema %s.properties.%s)", err, schemaName, propName)
		}

		if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs, schemaName+".properties."+propName); err != nil {
			return err
		}
	}
	return nil
}

// processNestedPropertyObjects processes nested objects within properties
func processNestedPropertyObjects(doc *openapi3.T, filtered *openapi3.T, propSchema *openapi3.SchemaRef, schemaName, propName string, processedRefs map[string]bool) error {
	if propSchema.Value == nil {
		return nil
	}

	// Handle arrays of objects in properties
	if propSchema.Value.Items != nil && propSchema.Value.Items.Ref != "" {
		refName, err := validateRef(propSchema.Value.Items.Ref, createLocation(fmt.Sprintf("schema.%s.properties.%s.items", schemaName, propName)))
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
		return processNestedProperties(doc, filtered, propSchema, schemaName, propName, processedRefs)
	}

	return nil
}

// processNestedProperties processes deeply nested properties
func processNestedProperties(doc *openapi3.T, filtered *openapi3.T, propSchema *openapi3.SchemaRef, schemaName, propName string, processedRefs map[string]bool) error {
	for nestedPropName, nestedProp := range propSchema.Value.Properties {
		if nestedProp.Ref != "" {
			refName, err := validateRef(nestedProp.Ref, createLocation(fmt.Sprintf("schema.%s.properties.%s.%s", schemaName, propName, nestedPropName)))
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
			refName, err := validateRef(nestedProp.Value.Items.Ref, createLocation(fmt.Sprintf("schema.%s.properties.%s.%s.items", schemaName, propName, nestedPropName)))
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
	return nil
}

// processCompositionSchemas processes allOf, oneOf, anyOf schemas
func processCompositionSchemas(doc *openapi3.T, filtered *openapi3.T, schema *openapi3.SchemaRef, schemaName string, processedRefs map[string]bool) error {
	compositionTypes := []struct {
		schemas []*openapi3.SchemaRef
		name    string
	}{
		{schema.Value.AllOf, "allOf"},
		{schema.Value.OneOf, "oneOf"},
		{schema.Value.AnyOf, "anyOf"},
	}

	for _, compType := range compositionTypes {
		for i, compositionSchema := range compType.schemas {
			if compositionSchema.Ref != "" {
				refName, err := validateRef(compositionSchema.Ref, createLocation(fmt.Sprintf("schema.%s.%s[%d]", schemaName, compType.name, i)))
				if err != nil {
					return fmt.Errorf("%w (in schema %s.%s[%d])", err, schemaName, compType.name, i)
				}

				if err := resolveSchemaRefsRecursively(doc, filtered, refName, processedRefs,
					fmt.Sprintf("%s.%s[%d]", schemaName, compType.name, i)); err != nil {
					return err
				}
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

	mimeTypeSet := getDefaultMimeTypes()

	// Search for MIME types in operations
	for _, pathItem := range doc.Paths.Map() {
		if pathItem != nil {
			collectMimeTypesFromPathItem(pathItem, mimeTypeSet)
		}
	}

	// Convert set to slice
	return convertMimeTypeSetToSlice(mimeTypeSet)
}

// getDefaultMimeTypes returns the default MIME types to always include
func getDefaultMimeTypes() map[string]struct{} {
	mimeTypeSet := make(map[string]struct{})
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
	return mimeTypeSet
}

// collectMimeTypesFromPathItem collects MIME types from all operations in a path item
func collectMimeTypesFromPathItem(pathItem *openapi3.PathItem, mimeTypeSet map[string]struct{}) {
	for _, operation := range pathItem.Operations() {
		if operation != nil {
			collectMimeTypesFromOperation(operation, mimeTypeSet)
		}
	}
}

// collectMimeTypesFromOperation collects MIME types from an operation
func collectMimeTypesFromOperation(operation *openapi3.Operation, mimeTypeSet map[string]struct{}) {
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

// convertMimeTypeSetToSlice converts a MIME type set to a slice
func convertMimeTypeSetToSlice(mimeTypeSet map[string]struct{}) []string {
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
		schemaName, err := validateRef(schema.Ref, createLocation("schema.ref"))
		if err != nil {
			return err
		}
		processedSchemaRefs[schemaName] = true
	}

	// Process schema value
	if schema.Value != nil {
		if err := extractSchemaValueReferences(schema.Value, processedSchemaRefs); err != nil {
			return err
		}
	}

	return nil
}

// extractSchemaValueReferences extracts references from a schema value
func extractSchemaValueReferences(schemaValue *openapi3.Schema, processedSchemaRefs map[string]bool) error {
	// Array items
	if schemaValue.Items != nil {
		if err := extractSchemaReferences(schemaValue.Items, processedSchemaRefs); err != nil {
			return err
		}
	}

	// Object properties
	for _, propSchema := range schemaValue.Properties {
		if err := extractSchemaReferences(propSchema, processedSchemaRefs); err != nil {
			return err
		}
	}

	// Composition schemas
	if err := extractCompositionSchemaReferences(schemaValue, processedSchemaRefs); err != nil {
		return err
	}

	// Not schema
	if schemaValue.Not != nil {
		if err := extractSchemaReferences(schemaValue.Not, processedSchemaRefs); err != nil {
			return err
		}
	}

	return nil
}

// extractCompositionSchemaReferences extracts references from composition schemas (allOf, oneOf, anyOf)
func extractCompositionSchemaReferences(schemaValue *openapi3.Schema, processedSchemaRefs map[string]bool) error {
	compositionTypes := [][]*openapi3.SchemaRef{
		schemaValue.AllOf,
		schemaValue.OneOf,
		schemaValue.AnyOf,
	}

	for _, compositionSchemas := range compositionTypes {
		for _, compositionSchema := range compositionSchemas {
			if err := extractSchemaReferences(compositionSchema, processedSchemaRefs); err != nil {
				return err
			}
		}
	}

	return nil
}
