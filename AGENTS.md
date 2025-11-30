# PG Press Agent Guidelines

## Build/Lint/Test Commands
- `make init` - Initialize project dependencies
- `make build` - Build the application
- `make dev` - Run development server with hot reloading
- `make test` - Run all tests
- `make test -run TestFunctionName` - Run a specific test
- `make lint` - Run linters

## Code Style Guidelines

### Imports
- Group imports by standard library, external dependencies, and internal packages
- Use `gofumpt` for import formatting
- Sort imports alphabetically within each group

### Formatting
- Use `gofumpt` for code formatting
- Follow Go's official style guide
- Maintain consistent indentation (tabs/spaces)

### Types
- Use clear, descriptive type names
- Prefer named types over anonymous types when possible
- Use interfaces for dependencies and abstractions

### Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for exported types and methods
- Use descriptive names that clearly indicate purpose

### Error Handling
- Always handle errors explicitly
- Use named return values for error handling when appropriate
- Log errors with context using structured logging

### Testing
- Write tests in the same directory as the code they test
- Use table-driven tests for multiple test cases
- Test both happy path and error conditions

### Security
- Sanitize all user inputs
- Use secure cookie configuration
- Validate and authenticate all requests

### Performance
- Optimize database queries with proper indexing
- Use connection pooling for database access
- Minimize memory allocations in hot paths