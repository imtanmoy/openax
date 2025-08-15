# OpenAx 🚀

[![npm version](https://img.shields.io/npm/v/openax-cli.svg)](https://www.npmjs.com/package/openax-cli)
[![CI](https://github.com/imtanmoy/openax/workflows/CI/badge.svg)](https://github.com/imtanmoy/openax/actions)
[![MIT License](https://img.shields.io/github/license/imtanmoy/openax)](LICENSE)
[![Release](https://img.shields.io/github/release/imtanmoy/openax.svg)](https://github.com/imtanmoy/openax/releases/latest)

A powerful command-line tool for filtering OpenAPI 3.x specifications. Extract specific paths, operations, and tags from large API specs while automatically resolving and including only the referenced components.

## ✨ Features

- 🔍 **Smart Filtering**: Filter by paths, HTTP operations, and tags
- 🧩 **Dependency Resolution**: Automatically includes referenced components
- 📁 **Multiple Sources**: Load from files, URLs, or raw data
- ✅ **Validation**: Built-in OpenAPI 3.x validation
- 🛠️ **CLI & Library**: Use as a command-line tool or Go library
- 📤 **Multiple Formats**: Output to JSON or YAML
- 🚀 **High Performance**: Efficient filtering with proper reference resolution

## 🚀 Quick Start

### Installation

```bash
# Install globally via npm
npm install -g openax-cli

# Or use directly with npx (no installation required)
npx openax-cli --help
```

#### Alternative Installation
- **Binary downloads**: [GitHub Releases](https://github.com/imtanmoy/openax/releases)
- **For Go developers**: `go install github.com/imtanmoy/openax@latest`

### CLI Usage

```bash
# Validate an OpenAPI spec
openax --validate-only -i api.yaml

# Filter by tags
openax -i api.yaml --tags "users,orders" --format json

# Filter by operations and paths
openax -i api.yaml --operations "get,post" --paths "/api/v1"

# Save filtered result
openax -i api.yaml --tags "public" -o public-api.yaml
```

### Examples

```bash
# Validate an OpenAPI spec
openax --validate-only -i api.yaml

# Filter by tags and save as JSON
openax -i api.yaml --tags "users,orders" --format json -o filtered.json

# Filter by operations and paths
openax -i api.yaml --operations "get,post" --paths "/api/v1" 

# Filter public APIs only
openax -i api.yaml --tags "public" -o public-api.yaml

# Use with remote URLs
openax -i https://api.example.com/openapi.yaml --tags "v1"
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

## 🎯 Common Use Cases

- **API Documentation**: Create focused docs from large specifications
- **Client Generation**: Generate clients for specific service areas  
- **Testing**: Create minimal specs for testing specific functionality
- **Micro-services**: Extract service-specific APIs from monolithic specs
- **Public APIs**: Filter internal specs to expose only public endpoints
- **Versioning**: Create version-specific API specifications

## 🔧 Advanced Usage

### Working with Large APIs
```bash
# Pre-validate large specs before filtering
openax --validate-only -i large-api.yaml

# Use specific filters for better performance
openax -i large-api.yaml --paths "/api/v1/users" --format json

# Combine multiple filters
openax -i api.yaml --tags "public" --operations "get,post" --paths "/api"
```

### Integration with CI/CD
```bash
# Validate API specs in CI
openax --validate-only -i api.yaml || exit 1

# Generate client-specific specs
openax -i api.yaml --tags "mobile" -o mobile-api.yaml
openax -i api.yaml --tags "web" -o web-api.yaml
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔗 Links

- **npm package**: [openax-cli](https://www.npmjs.com/package/openax-cli)
- **GitHub repository**: [imtanmoy/openax](https://github.com/imtanmoy/openax)
- **Issues & Support**: [GitHub Issues](https://github.com/imtanmoy/openax/issues)
- **Go library docs**: [pkg.go.dev](https://pkg.go.dev/github.com/imtanmoy/openax)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**For Go developers**: This tool is also available as a Go library. See the [Go documentation](https://pkg.go.dev/github.com/imtanmoy/openax) for library usage examples.
