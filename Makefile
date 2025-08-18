# LangExtract-Go Makefile

.PHONY: help build test lint fmt vet clean install-tools tidy deps-upgrade

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the project"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  tidy          - Clean up go.mod"
	@echo "  deps-upgrade  - Upgrade dependencies"

# Build the project
build:
	@echo "Building..."
	go build -v ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -w .
	goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean ./...
	rm -f coverage.out coverage.html

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Clean up go.mod
tidy:
	@echo "Tidying go.mod..."
	go mod tidy

# Upgrade dependencies
deps-upgrade:
	@echo "Upgrading dependencies..."
	go get -u ./...
	go mod tidy

# CI target - runs all checks
ci: fmt vet lint test
	@echo "All CI checks passed!"

# Development setup
dev-setup: install-tools tidy
	@echo "Development environment setup complete!"