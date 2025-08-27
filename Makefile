# LangExtract-Go Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build information
BINARY_NAME = langextract
MAIN_PACKAGE = ./cmd/langextract
BUILD_DIR = build
DIST_DIR = dist

# Go build flags
LDFLAGS = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)
BUILD_FLAGS = -ldflags="$(LDFLAGS)"

# Cross-compilation targets
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: help build test lint fmt vet clean install-tools tidy deps-upgrade build-cross build-all-platforms release dist security-scan gosec vuln-check license-check install-security-tools docs docs-html docs-markdown docs-serve docs-clean

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  build             - Build the project"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  lint              - Run linters"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"
	@echo "  clean             - Clean build artifacts"
	@echo "  install-tools     - Install development tools"
	@echo "  tidy              - Clean up go.mod"
	@echo "  deps-upgrade      - Upgrade dependencies"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  build-cross       - Build for current OS, multiple architectures"
	@echo "  build-all-platforms - Build for all supported platforms"
	@echo "  build-linux       - Build for Linux (amd64, arm64)"
	@echo "  build-darwin      - Build for macOS (amd64, arm64)"
	@echo "  build-windows     - Build for Windows (amd64, arm64)"
	@echo ""
	@echo "Release:"
	@echo "  release           - Create release builds with version info"
	@echo "  dist              - Create distribution archives"
	@echo ""
	@echo "Version Management:"
	@echo "  version-show      - Show current version information"
	@echo "  version-bump-patch - Bump patch version (1.0.0 -> 1.0.1)"
	@echo "  version-bump-minor - Bump minor version (1.0.0 -> 1.1.0)"
	@echo "  version-bump-major - Bump major version (1.0.0 -> 2.0.0)"
	@echo ""
	@echo "Release Automation:"
	@echo "  release-patch     - Create patch release"
	@echo "  release-minor     - Create minor release"
	@echo "  release-major     - Create major release"
	@echo "  release-custom VERSION=v1.2.3 - Create custom version release"
	@echo ""
	@echo "Security:"
	@echo "  security-scan     - Run all security scans"
	@echo "  gosec            - Run gosec security scanner"
	@echo "  vuln-check       - Run vulnerability check"
	@echo "  license-check    - Check license compliance"
	@echo "  install-security-tools - Install security scanning tools"
	@echo ""
	@echo "Documentation:"
	@echo "  docs             - Generate HTML documentation"
	@echo "  docs-html        - Generate HTML documentation"
	@echo "  docs-markdown    - Generate Markdown documentation"
	@echo "  docs-serve       - Generate and serve HTML documentation"
	@echo "  docs-clean       - Clean generated documentation"

# Build the project
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

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
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f build/gosec-report.json build/vuln-report.txt build/licenses.csv
	rm -rf docs/api/

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

# Cross-compilation targets
build-cross:
	@echo "Building for current OS with multiple architectures..."
	@mkdir -p $(BUILD_DIR)
	@OS=$$(go env GOOS); \
	echo "Building for $$OS/amd64..."; \
	GOOS=$$OS GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-amd64 $(MAIN_PACKAGE); \
	echo "Building for $$OS/arm64..."; \
	GOOS=$$OS GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-arm64 $(MAIN_PACKAGE); \
	echo "Cross-compilation completed for $$OS"

build-all-platforms:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_EXT=""; \
		if [ "$$OS" = "windows" ]; then \
			BINARY_EXT=".exe"; \
		fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$BINARY_EXT $(MAIN_PACKAGE) || exit 1; \
	done
	@echo "All platform builds completed"

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@echo "Linux builds completed"

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@echo "macOS builds completed"

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Building for windows/arm64..."
	@GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PACKAGE)
	@echo "Windows builds completed"

# Release builds with optimizations
release: clean
	@echo "Creating optimized release builds..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_EXT=""; \
		if [ "$$OS" = "windows" ]; then \
			BINARY_EXT=".exe"; \
		fi; \
		echo "Creating release build for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags="$(LDFLAGS) -s -w" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$BINARY_EXT $(MAIN_PACKAGE) || exit 1; \
	done
	@echo "Release builds completed"

# Create distribution archives
dist: release
	@echo "Creating distribution archives..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && for binary in $(BINARY_NAME)-*; do \
		if [[ "$$binary" == *windows* ]]; then \
			echo "Creating $$binary.zip..."; \
			zip -q ../$(DIST_DIR)/$$binary.zip $$binary; \
		else \
			echo "Creating $$binary.tar.gz..."; \
			tar -czf ../$(DIST_DIR)/$$binary.tar.gz $$binary; \
		fi; \
	done
	@echo "Distribution archives created in $(DIST_DIR)/"

# Version management targets
version-show:
	@scripts/version.sh show

version-bump-patch:
	@scripts/version.sh bump patch -f

version-bump-minor:
	@scripts/version.sh bump minor -f

version-bump-major:
	@scripts/version.sh bump major -f

# Release automation targets
release-patch: version-bump-patch
	@scripts/release.sh $(shell cat VERSION)

release-minor: version-bump-minor
	@scripts/release.sh $(shell cat VERSION)

release-major: version-bump-major
	@scripts/release.sh $(shell cat VERSION)

release-custom:
	@echo "Usage: make release-custom VERSION=v1.2.3"
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required"; \
		exit 1; \
	fi
	@scripts/release.sh $(VERSION)

# CI target - runs all checks
ci: fmt vet lint test
	@echo "All CI checks passed!"

# Development setup
dev-setup: install-tools tidy
	@echo "Development environment setup complete!"

# Security scanning tools installation
install-security-tools:
	@echo "Installing security scanning tools..."
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/google/go-licenses@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Security tools installed successfully!"

# Run gosec security scanner
gosec:
	@echo "Running gosec security scanner..."
	@mkdir -p build
	gosec -fmt json -out build/gosec-report.json -stdout -verbose=text ./...
	@echo "Gosec scan completed. Report saved to build/gosec-report.json"

# Run vulnerability check
vuln-check:
	@echo "Running vulnerability check..."
	@mkdir -p build
	govulncheck ./... | tee build/vuln-report.txt
	@echo "Vulnerability check completed. Report saved to build/vuln-report.txt"

# Check license compliance
license-check:
	@echo "Checking license compliance..."
	@mkdir -p build
	go-licenses csv ./... | tee build/licenses.csv
	@echo "License check completed. Report saved to build/licenses.csv"
	@echo ""
	@echo "Checking for forbidden licenses..."
	@FORBIDDEN="GPL-3.0,AGPL-3.0,LGPL-3.0"; \
	for license in $$(echo $$FORBIDDEN | tr ',' ' '); do \
		if grep -q "$$license" build/licenses.csv; then \
			echo "❌ Found forbidden license: $$license"; \
			grep "$$license" build/licenses.csv; \
			exit 1; \
		fi; \
	done; \
	echo "✅ All licenses are compliant"

# Run all security scans
security-scan: gosec vuln-check license-check
	@echo ""
	@echo "🔒 Security scan summary:"
	@echo "  - Gosec security scanner: ✓"
	@echo "  - Vulnerability check: ✓"
	@echo "  - License compliance: ✓"
	@echo ""
	@echo "Reports available in build/ directory:"
	@ls -la build/*report* build/licenses.csv 2>/dev/null || true
	@echo ""
	@echo "✅ All security scans completed successfully!"

# Enhanced CI target with security
ci-security: fmt vet lint test security-scan
	@echo "All CI checks including security scans passed!"

# Documentation generation
docs: docs-html
	@echo "Default documentation (HTML) generated successfully!"

docs-html:
	@echo "Generating HTML API documentation..."
	@scripts/generate-docs.sh -f html
	@echo "HTML documentation available at docs/api/html/index.html"

docs-markdown:
	@echo "Generating Markdown API documentation..."
	@scripts/generate-docs.sh -f markdown
	@echo "Markdown documentation available at docs/api/markdown/README.md"

docs-serve: docs-html
	@echo "Serving documentation at http://localhost:6060..."
	@scripts/generate-docs.sh -c -s -p 6060

docs-clean:
	@echo "Cleaning generated documentation..."
	@rm -rf docs/api/
	@echo "Documentation cleaned successfully!"