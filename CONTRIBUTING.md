# Contributing to OpenAx

Thank you for your interest in contributing to OpenAx! 🚀

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Getting Started

1. **Fork the repository**

   ```bash
   # Click "Fork" on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/openax.git
   cd openax
   ```

2. **Set up the development environment**

   ```bash
   # Install dependencies
   go mod download
   
   # Verify everything works
   make test
   ```

3. **Create a feature branch**

   ```bash
   git checkout -b feature/amazing-feature
   ```

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run tests with race detection
make test-race

# Run integration tests
make test-integration
```

### Code Quality

We maintain high code quality standards:

```bash
# Run all quality checks
make quality

# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Pre-commit checks
make pre-commit
```

### Building

```bash
# Build the binary
make build

# Test CLI functionality
make run

# Run examples
make examples
```

## Project Structure

```
openax/
├── cmd/                    # CLI implementation
│   ├── root.go            # CLI command definition
│   └── root_test.go       # CLI tests
├── pkg/                    # Public library packages
│   ├── openax/            # Main library package
│   │   ├── openax.go      # Public API
│   │   ├── filter.go      # Filtering logic
│   │   └── *_test.go      # Tests
│   ├── loader/            # Specification loading
│   └── validator/         # Validation utilities
├── examples/              # Usage examples
├── testdata/              # Test data
├── .github/               # GitHub workflows
├── main.go               # CLI entry point
└── Makefile              # Build automation
```

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Write table-driven tests where appropriate
- Handle errors appropriately
- Use testify for assertions in tests

### Example Function Documentation

```go
// LoadAndFilter loads an OpenAPI specification from the given source and applies
// the specified filters. It returns the filtered specification or an error if
// the operation fails.
//
// The source can be a file path, URL, or any other supported input format.
// Filters are applied in combination (AND logic) - all specified filters
// must match for a path to be included.
func (c *Client) LoadAndFilter(source string, opts FilterOptions) (*openapi3.T, error) {
    // Implementation...
}
```

### Test Guidelines

- Test files should be named `*_test.go`
- Use table-driven tests for multiple test cases
- Include both positive and negative test cases
- Test error conditions
- Use meaningful test names that describe the scenario

```go
func TestFilterByTags(t *testing.T) {
    testCases := []struct {
        name     string
        tags     []string
        expected int
    }{
        {
            name:     "filter by single tag",
            tags:     []string{"users"},
            expected: 2,
        },
        {
            name:     "filter by multiple tags",
            tags:     []string{"users", "public"},
            expected: 1,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Making Changes

### Before You Start

1. Check existing issues to avoid duplication
2. For large changes, create an issue to discuss the approach
3. Make sure tests pass locally

### Commit Guidelines

We follow conventional commit messages:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions or changes
- `refactor:` for code refactoring
- `chore:` for maintenance tasks

Examples:

```
feat: add support for filtering by operation IDs
fix: handle missing schema references correctly
docs: update README with new CLI options
test: add integration tests for complex filtering scenarios
```

### Pull Request Process

1. **Update your branch**

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run quality checks**

   ```bash
   make pre-commit
   ```

3. **Push your changes**

   ```bash
   git push origin feature/amazing-feature
   ```

4. **Create a Pull Request**
   - Use a clear title and description
   - Reference any related issues
   - Include examples if adding new features
   - Update documentation if needed

5. **Address feedback**
   - Respond to review comments
   - Make requested changes
   - Keep the conversation constructive

## Testing

### Test Types

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test complete workflows
- **CLI Tests**: Test command-line interface
- **Example Tests**: Ensure examples work correctly

### Adding Test Data

When adding test data to `testdata/`:

- Use realistic but minimal OpenAPI specifications
- Include both valid and invalid examples
- Document the purpose of each test file
- Keep file sizes small for fast tests

## Documentation

### Code Documentation

- Document all exported functions and types
- Include usage examples in package documentation
- Keep comments concise and helpful

### README Updates

When adding features:

- Update the feature list
- Add usage examples
- Update CLI reference if needed
- Include any new configuration options

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Review**: Ask questions in pull request reviews

## Recognition

Contributors are recognized in several ways:

- Listed in release notes for significant contributions
- Mentioned in commit messages
- Added to contributor lists
- Invited to be maintainers for ongoing contributors

Thank you for contributing to OpenAx! 🙏
