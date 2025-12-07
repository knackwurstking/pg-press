Build/Run:
- Build: make build
- Lint: make lint (golangci-lint)
- Test: go test ./...
- Single test: go test -run ^TestX$

Code Style:
- Imports: stdlib, external, local groups
- Naming: PascalCase (types), camelCase (variables)
- Errors: Explicit checks with if err != nil
- Formatting: gofmt -w
- Linting: golangci-lint run

Guidelines:
- No unused imports or unhandled errors
- Comment public APIs
- Use validation package for inputs
- Keep functions under 50 lines
- Always log errors with stack trace
- Commit messages: Always use semantic git commit message style

Cursor/Copilot:
- No custom rules detected