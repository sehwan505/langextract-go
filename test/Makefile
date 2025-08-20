# Test Makefile for LangExtract-Go

.PHONY: test test-unit test-integration test-e2e test-benchmarks test-all clean coverage

# Go test parameters
GO_TEST_FLAGS ?= -v
GO_TEST_TIMEOUT ?= 10m
COVERAGE_OUT ?= coverage.out

# Test directories
UNIT_DIR := ./unit/...
INTEGRATION_DIR := ./integration/...
E2E_DIR := ./e2e/...
BENCHMARKS_DIR := ./benchmarks/...

# Default target
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test $(GO_TEST_FLAGS) -timeout $(GO_TEST_TIMEOUT) $(UNIT_DIR)

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test $(GO_TEST_FLAGS) -timeout $(GO_TEST_TIMEOUT) $(INTEGRATION_DIR)

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	go test $(GO_TEST_FLAGS) -timeout $(GO_TEST_TIMEOUT) $(E2E_DIR)

# Run benchmarks
test-benchmarks:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem $(BENCHMARKS_DIR)

# Run all tests
test-all: test-unit test-integration test-e2e

# Generate test coverage
coverage:
	@echo "Generating test coverage..."
	go test -coverprofile=$(COVERAGE_OUT) $(UNIT_DIR)
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	rm -f $(COVERAGE_OUT) coverage.html
	go clean -testcache

# Lint tests
lint:
	@echo "Linting test files..."
	golangci-lint run $(UNIT_DIR) $(INTEGRATION_DIR) $(E2E_DIR)

# Test with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race $(GO_TEST_FLAGS) $(UNIT_DIR)

# Run tests with verbose output and keep test cache
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -count=1 $(UNIT_DIR)

help:
	@echo "Available targets:"
	@echo "  test          - Run unit tests (default)"
	@echo "  test-unit     - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e      - Run end-to-end tests"
	@echo "  test-benchmarks - Run benchmarks"
	@echo "  test-all      - Run all test suites"
	@echo "  coverage      - Generate test coverage report"
	@echo "  clean         - Clean test artifacts"
	@echo "  lint          - Lint test files"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  help          - Show this help message"