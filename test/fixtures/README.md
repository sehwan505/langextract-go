# Test Fixtures

This directory contains test data, example documents, and golden files used throughout the test suite.

## Structure

- `documents/` - Sample documents for testing extraction
- `examples/` - Example extraction tasks and expected results
- `golden/` - Golden files with expected outputs
- `schemas/` - JSON schemas for testing schema validation
- `prompts/` - Standard prompts for testing

## Usage

Test fixtures are organized by functionality and should be referenced by tests using relative paths from the test directory.

```go
// Example usage in test files
const fixturesPath = "../fixtures/documents/sample.txt"
```

## Adding New Fixtures

1. Place files in appropriate subdirectories
2. Use descriptive names that indicate the test scenario
3. Include a comment or README explaining complex fixtures
4. Keep fixtures minimal but representative