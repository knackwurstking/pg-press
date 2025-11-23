# AGENTS.md

## Build/Lint/Test Commands

```bash
make build          # Build production binary
make run            # Generate templates and run server
make dev            # Run with hot reload using gow
make generate       # Generate templ templates
make init           # Initialize project (tidy modules, update submodules)
go test ./...       # Run all tests
go test -v ./...    # Run all tests with verbose output
go test -run TestName ./...  # Run a specific test
```

## Code Style Guidelines

### Imports
- Group imports in order: standard library, external libraries, internal packages
- Use descriptive aliases for long package names (e.g., `echo "github.com/labstack/echo/v4"`)

### Formatting
- Use Go's standard formatting (gofmt) for all Go files
- Maintain consistent indentation with tabs (not spaces)

### Types
- Use specific types instead of `interface{}` where possible
- Prefer explicit type declarations over type inference for readability
- Use pointer types for large structs to avoid copying

### Naming Conventions
- Use camelCase for Go identifiers (e.g., `toolName`, `serverAddr`)
- Use PascalCase for exported names (e.g., `Tool`, `Server`)
- Use lowercase for package names (e.g., `services`, `utils`)

### Error Handling
- Handle errors explicitly and provide meaningful error messages
- Use custom error types for better error categorization (e.g., `errors.AlreadyExists`)
- Log errors with context using structured logging
