# AGENTS

## Build
- `go build ./...`

## Test
- All tests: `go test ./...`
- Single test: `go test -run ^TestName$ ./...`

## Lint / Vet
- `go vet ./...`
- `golangci-lint run`

## Format
- `go fmt ./...` - `goimports -l -w .` - `go mod tidy`

## Code style guidelines
- Imports std first, external, internal, alphabetically; Types exported UpperCamelCase, unexported camelCase; Constants UpperCamelCase, typed; Errors wrapped via `fmt.Errorf("%w", err)`, never nil; JSON tags `json:"fieldName,omitempty"`.
- Tests *_test.go, use t.Helper(); Logging single logger interface; defer context.

## Git commit guidelines
- Commit message format (semantic) - `type(scope): subject`
- The `(scope)` is optional