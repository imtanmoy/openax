# OpenAx 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/imtanmoy/openax.svg)](https://pkg.go.dev/github.com/imtanmoy/openax)
[![Go Report Card](https://goreportcard.com/badge/github.com/imtanmoy/openax)](https://goreportcard.com/report/github.com/imtanmoy/openax)

OpenAx is a powerful OpenAPI 3.x specification filtering tool and Go library. It allows you to filter large OpenAPI specs by paths, operations, and tags while automatically resolving and including only the referenced components (schemas, parameters, request bodies, responses).

## ✨ Features

- 🔍 **Smart Filtering**: Filter by paths, HTTP operations, and tags
- 🧩 **Dependency Resolution**: Automatically includes referenced components
- 📁 **Multiple Sources**: Load from files, URLs, or raw data
- ✅ **Validation**: Built-in OpenAPI 3.x validation
- 🛠️ **CLI & Library**: Use as a command-line tool or Go library
- 📤 **Multiple Formats**: Output to JSON or YAML
- 🚀 **High Performance**: Efficient filtering with proper reference resolution

## 🚀 Quick Start

### CLI Usage

```bash
# Install
go install github.com/imtanmoy/openax@latest

# Validate an OpenAPI spec
openax --validate-only -i api.yaml

# Filter by tags
openax -i api.yaml --tags "users,orders" --format json

# Filter by operations and paths
openax -i api.yaml --operations "get,post" --paths "/api/v1"

# Save filtered result
openax -i api.yaml --tags "public" -o public-api.yaml
```

### Library Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/imtanmoy/openax/pkg/openax"
)

func main() {
    // Create client
    client := openax.New()
    
    // Filter by tags
    filtered, err := client.LoadAndFilter("api.yaml", openax.FilterOptions{
        Tags: []string{"users", "orders"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Filtered spec has %d paths\\n", filtered.Paths.Len())
}
```

## 📖 Documentation

### CLI Reference

```bash
openax [options]

Flags:
  -i, --input string         Input OpenAPI spec file or URL (required)
  -o, --output string        Output file (stdout if not specified)
  -f, --format string        Output format: json or yaml (default: yaml)
  -p, --paths strings        Filter by paths (e.g., /users, /orders)
      --operations strings   Filter by operations (e.g., get, post, put, delete)
  -t, --tags strings         Filter by tags
      --validate-only        Only validate the spec without filtering
  -h, --help                 Show help
  -v, --version             Show version
```

### Library API

#### Basic Client

```go
// Create client with defaults
client := openax.New()

// Create client with custom options
client := openax.NewWithOptions(openax.LoadOptions{
    AllowExternalRefs: true,
    Context:           context.Background(),
})
```

#### Loading Specifications

```go
// From file
doc, err := client.LoadFromFile("api.yaml")

// From URL
doc, err := client.LoadFromURL("https://api.example.com/openapi.yaml")

// From raw data
doc, err := client.LoadFromData(yamlData)

// Auto-detect source
doc, err := client.LoadAndFilter("api.yaml", openax.FilterOptions{...})
```

#### Filtering Options

```go
opts := openax.FilterOptions{
    Paths:      []string{"/users", "/orders"},           // Path prefixes
    Operations: []string{"get", "post", "updateUser"},   // HTTP methods or operation IDs
    Tags:       []string{"public", "v1"},                // Tag names
}

filtered, err := client.Filter(doc, opts)
```

#### Validation

```go
// Validate only
err := client.ValidateOnly("api.yaml")

// Validate loaded document
err := client.Validate(doc)
```

### Advanced Usage

#### Using Individual Packages

```go
import (
    "github.com/imtanmoy/openax/pkg/loader"
    "github.com/imtanmoy/openax/pkg/validator"
    "github.com/imtanmoy/openax/pkg/openax"
)

// Create specialized components
l := loader.New()
v := validator.New()

// Load and validate
doc, err := l.LoadFromFile("api.yaml")
if err := v.Validate(doc); err != nil {
    // handle validation error
}

// Apply filtering
filtered, err := openax.New().Filter(doc, options)
```

#### Custom Filtering Logic

```go
// Load with openax components
doc, err := client.LoadFromFile("api.yaml")

// Apply your custom filtering logic
customFiltered := myCustomFilter(doc)

// Then apply standard openax filtering
finalResult, err := client.Filter(customFiltered, openax.FilterOptions{
    Operations: []string{"get"},
})
```

## 📁 Project Structure

```
openax/
├── cmd/                    # CLI implementation
├── pkg/                    # Public library packages
│   ├── openax/            # Main library package
│   ├── loader/            # Specification loading utilities
│   └── validator/         # Validation utilities
├── examples/              # Usage examples
│   ├── library/           # Library usage examples
│   ├── cli/               # CLI usage examples
│   └── custom-filter/     # Custom filtering examples
├── testdata/              # Test specifications
│   ├── specs/             # Sample OpenAPI specs
│   └── expected/          # Expected filtering results
└── main.go               # CLI entry point
```

## 🎯 Use Cases

- **API Documentation**: Create focused docs from large specifications
- **Client Generation**: Generate clients for specific service areas
- **Testing**: Create minimal specs for testing specific functionality
- **Microservices**: Extract service-specific APIs from monolithic specs
- **Public APIs**: Filter internal specs to expose only public endpoints
- **Versioning**: Create version-specific API specifications

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [kin-openapi](https://github.com/getkin/kin-openapi) for OpenAPI 3.x support
- CLI powered by [urfave/cli](https://github.com/urfave/cli) v3
- Inspired by the need for better OpenAPI specification management