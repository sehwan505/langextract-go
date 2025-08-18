# LangExtract-Go Code Conventions

## Overview
This document outlines the coding standards, conventions, and best practices for the LangExtract-Go project.

## Go Standards
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` and `goimports` for consistent formatting
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `golangci-lint` for static analysis

## Project Structure

```
langextract-go/
├── cmd/                    # Main applications (CLI tools)
├── pkg/                    # Public API packages
│   ├── types/             # Core data types and interfaces
│   ├── document/          # Document-related structures
│   ├── extraction/        # Extraction-related structures  
│   ├── providers/         # LLM provider interfaces and implementations
│   ├── pipeline/          # Extraction pipeline
│   ├── alignment/         # Text alignment and grounding
│   └── visualization/     # HTML visualization
├── internal/              # Private packages
│   ├── common/           # Shared utilities
│   ├── config/           # Configuration management
│   └── testutil/         # Test utilities
├── api/                   # API definitions (OpenAPI, protobuf)
├── configs/               # Configuration files
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
├── examples/              # Example code and use cases
└── test/                  # Integration and end-to-end tests
```

## Naming Conventions

### Packages
- Use lowercase, single-word package names
- Avoid underscores, hyphens, or mixed case
- Package names should be concise and descriptive

### Types
- Use PascalCase for exported types: `CharInterval`, `Document`
- Use camelCase for unexported types: `internalConfig`
- Interface names should describe behavior: `Provider`, `Extractor`
- Avoid stuttering: `types.Type` not `types.TypeType`

### Functions and Methods
- Use PascalCase for exported functions: `NewDocument()`
- Use camelCase for unexported functions: `parseDocument()`
- Use descriptive names: `ExtractEntities()` not `Extract()`

### Variables and Constants
- Use camelCase for variables: `maxRetries`
- Use PascalCase for exported constants: `DefaultTimeout`
- Use ALL_CAPS for unexported constants: `MAX_CHUNK_SIZE`

### Files
- Use lowercase with underscores for separation: `char_interval.go`
- Test files should end with `_test.go`
- Benchmark files should end with `_bench_test.go`

## Error Handling
- Always handle errors explicitly
- Use descriptive error messages with context
- Wrap errors with `fmt.Errorf` for additional context
- Use custom error types for specific error conditions
- Return errors as the last return value

```go
func ProcessDocument(doc Document) (*AnnotatedDocument, error) {
    if doc.Text == "" {
        return nil, fmt.Errorf("document text cannot be empty")
    }
    
    result, err := extractEntities(doc)
    if err != nil {
        return nil, fmt.Errorf("failed to extract entities: %w", err)
    }
    
    return result, nil
}
```

## Documentation
- All exported types, functions, and methods must have doc comments
- Doc comments should start with the name of the item being documented
- Use complete sentences in documentation
- Include examples in doc comments when helpful

```go
// CharInterval represents a character position range in text.
// StartPos is inclusive, EndPos is exclusive.
type CharInterval struct {
    StartPos int // Inclusive start position
    EndPos   int // Exclusive end position
}

// Length returns the number of characters in the interval.
func (ci CharInterval) Length() int {
    return ci.EndPos - ci.StartPos
}
```

## Testing
- Every public function should have tests
- Use table-driven tests for multiple test cases
- Test file names should end with `_test.go`
- Use descriptive test function names: `TestCharInterval_Length_ReturnsCorrectValue`
- Use testify for assertions where appropriate

```go
func TestCharInterval_Length(t *testing.T) {
    tests := []struct {
        name     string
        interval CharInterval
        expected int
    }{
        {"empty interval", CharInterval{0, 0}, 0},
        {"single character", CharInterval{0, 1}, 1},
        {"multiple characters", CharInterval{5, 10}, 5},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.interval.Length())
        })
    }
}
```

## Code Organization
- Keep functions small and focused (prefer <50 lines)
- Group related functionality in the same package
- Use interfaces to define contracts between packages
- Minimize dependencies between packages
- Use dependency injection for better testability

## Performance Considerations
- Use value receivers for small types, pointer receivers for large types
- Avoid allocations in hot paths
- Use sync.Pool for frequently allocated objects
- Profile code to identify bottlenecks
- Use context.Context for cancellation and timeouts

## Dependencies
- Minimize external dependencies
- Prefer standard library when possible
- Pin dependency versions in go.mod
- Regularly update dependencies for security patches
- Use go mod tidy to keep dependencies clean

## Git Conventions
- Use conventional commit messages
- Keep commits atomic and focused
- Use descriptive branch names: `feature/char-interval`, `fix/alignment-bug`
- Squash commits before merging to main
- Use pull requests for all changes