# Build & Test Commands
go build -o opencode       # Build binary
go test ./...             # Run all tests
go test ./path/to/pkg -v  # Run specific package tests with verbose output
go test -run TestName     # Run specific test
go vet ./...             # Run static analysis
go install               # Install binary

# Code Style
- Go version: 1.24.0
- Imports: Grouped standard lib, 3rd party, local packages; goimports ordering
- Error handling: Return errors, use multierr for combining errors
- Naming: CamelCase for exported, camelCase for private
- Testing: *_test.go files with TestXxx functions, t.Error for failures
- Types: Prefer strong typing, use interfaces for abstraction
- Formatting: gofmt standard, 80-char line limit where practical
- Comments: Full sentences for exported symbols, concise inline comments
- Error wrapping: Use fmt.Errorf("doing x: %w", err)
- Folder structure: cmd/, internal/, pkg/ organization with focused packages