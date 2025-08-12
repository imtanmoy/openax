# CLI Examples

This directory contains examples of using OpenAx as a CLI tool.

## Basic Usage

```bash
# Validate an OpenAPI spec
openax --validate-only -i spec.yaml

# Filter by tags
openax -i spec.yaml --tags pet,user

# Filter by operations
openax -i spec.yaml --operations get,post

# Filter by paths
openax -i spec.yaml --paths /pet,/user

# Combine filters
openax -i spec.yaml --tags pet --operations get

# Output to JSON
openax -i spec.yaml --tags pet --format json

# Save to file
openax -i spec.yaml --tags pet -o filtered.yaml
```

## Advanced Examples

```bash
# Filter complex spec with multiple criteria
openax -i petstore.yaml \
  --tags "pet,store" \
  --operations "get,post,put" \
  --paths "/pet,/store" \
  --format json \
  -o filtered-petstore.json

# Validate from URL
openax --validate-only -i https://api.example.com/openapi.yaml

# Filter from URL and save locally
openax -i https://api.example.com/openapi.yaml \
  --tags "public" \
  -o public-api.yaml
```

## Use Cases

1. **API Documentation**: Filter large specs to create focused documentation
2. **Client Generation**: Generate clients for specific service areas
3. **Testing**: Create minimal specs for testing specific functionality
4. **API Contracts**: Extract public-facing APIs from internal specs