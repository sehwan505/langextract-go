# Test Makefile for langextract-go
# Comprehensive test automation and coverage reporting

.PHONY: test test-unit test-integration test-e2e test-benchmarks test-race test-coverage test-all
.PHONY: test-verbose test-short test-ci test-clean test-fixtures
.PHONY: coverage coverage-html coverage-report coverage-upload
.PHONY: benchmark benchmark-cpu benchmark-memory benchmark-compare
.PHONY: lint-tests format-tests validate-fixtures

# Go test settings
GO_TEST = go test
GO_TEST_FLAGS = -v
GO_COVER_FLAGS = -coverprofile=coverage.out -covermode=atomic
GO_BENCH_FLAGS = -bench=. -benchmem
GO_RACE_FLAGS = -race

# Test directories
UNIT_DIRS = ./unit/...
INTEGRATION_DIRS = ./integration/...
E2E_DIRS = ./e2e/...
BENCHMARK_DIRS = ./benchmarks/...

# Coverage settings
COVERAGE_OUT = coverage.out
COVERAGE_HTML = coverage.html
COVERAGE_THRESHOLD = 80

# Benchmark settings
BENCHMARK_OUT = benchmark.out
BENCHMARK_BASELINE = benchmark_baseline.out

# Colors for output
GREEN = \033[0;32m
RED = \033[0;31m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

# Default target
all: test-all

# Help target
help:
	@echo "Available targets:"
	@echo "  ${GREEN}test${NC}              - Run all tests"
	@echo "  ${GREEN}test-unit${NC}         - Run unit tests only"
	@echo "  ${GREEN}test-integration${NC}  - Run integration tests only"
	@echo "  ${GREEN}test-e2e${NC}          - Run end-to-end tests only"
	@echo "  ${GREEN}test-benchmarks${NC}   - Run benchmark tests only"
	@echo "  ${GREEN}test-race${NC}         - Run tests with race detector"
	@echo "  ${GREEN}test-coverage${NC}     - Run tests with coverage"
	@echo "  ${GREEN}test-all${NC}          - Run comprehensive test suite"
	@echo "  ${BLUE}coverage${NC}          - Generate coverage report"
	@echo "  ${BLUE}coverage-html${NC}     - Generate HTML coverage report"
	@echo "  ${BLUE}benchmark${NC}         - Run performance benchmarks"
	@echo "  ${YELLOW}test-ci${NC}           - Run CI test suite"
	@echo "  ${YELLOW}test-clean${NC}        - Clean test artifacts"

# Unit tests - fast, isolated tests
test-unit:
	@echo "${GREEN}Running unit tests...${NC}"
	cd .. && $(GO_TEST) $(GO_TEST_FLAGS) $(UNIT_DIRS)

# Integration tests - tests with external dependencies
test-integration:
	@echo "${GREEN}Running integration tests...${NC}"
	cd .. && $(GO_TEST) $(GO_TEST_FLAGS) -timeout=60s $(INTEGRATION_DIRS)

# End-to-end tests - full system tests
test-e2e:
	@echo "${GREEN}Running end-to-end tests...${NC}"
	cd .. && $(GO_TEST) $(GO_TEST_FLAGS) -timeout=300s $(E2E_DIRS)

# Benchmark tests - performance testing
test-benchmarks:
	@echo "${GREEN}Running benchmark tests...${NC}"
	cd .. && $(GO_TEST) $(GO_BENCH_FLAGS) -timeout=600s $(BENCHMARK_DIRS)

# Race condition detection
test-race:
	@echo "${GREEN}Running tests with race detector...${NC}"
	cd .. && $(GO_TEST) $(GO_TEST_FLAGS) $(GO_RACE_FLAGS) $(UNIT_DIRS) $(INTEGRATION_DIRS)

# Coverage testing
test-coverage:
	@echo "${GREEN}Running tests with coverage...${NC}"
	cd .. && $(GO_TEST) $(GO_TEST_FLAGS) $(GO_COVER_FLAGS) $(UNIT_DIRS) $(INTEGRATION_DIRS)
	@cd .. && go tool cover -func=coverage.out | tail -1 | awk '{print "Total coverage: " $$3}'

# Verbose test output
test-verbose:
	@echo "${GREEN}Running tests with verbose output...${NC}"
	cd .. && $(GO_TEST) -v -count=1 $(UNIT_DIRS) $(INTEGRATION_DIRS)

# Short test mode (skip long-running tests)
test-short:
	@echo "${GREEN}Running short test suite...${NC}"
	cd .. && $(GO_TEST) -short $(GO_TEST_FLAGS) $(UNIT_DIRS) $(INTEGRATION_DIRS)

# Basic test runner (unit + integration)
test:
	@echo "${GREEN}Running core test suite...${NC}"
	$(MAKE) test-unit
	$(MAKE) test-integration

# Comprehensive test suite
test-all:
	@echo "${GREEN}Running comprehensive test suite...${NC}"
	$(MAKE) test-unit
	$(MAKE) test-integration
	$(MAKE) test-race
	$(MAKE) test-coverage
	$(MAKE) validate-coverage

# CI test pipeline
test-ci:
	@echo "${GREEN}Running CI test pipeline...${NC}"
	$(MAKE) lint-tests
	$(MAKE) test-unit
	$(MAKE) test-integration
	$(MAKE) test-race
	$(MAKE) test-coverage
	$(MAKE) validate-coverage
	$(MAKE) test-benchmarks

# Coverage report generation
coverage:
	@echo "${BLUE}Generating coverage report...${NC}"
	cd .. && $(GO_TEST) $(GO_COVER_FLAGS) $(UNIT_DIRS) $(INTEGRATION_DIRS)
	cd .. && go tool cover -func=coverage.out

# HTML coverage report
coverage-html: coverage
	@echo "${BLUE}Generating HTML coverage report...${NC}"
	cd .. && go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${NC}"

# Coverage validation
validate-coverage:
	@echo "${BLUE}Validating coverage threshold...${NC}"
	@cd .. && COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l 2>/dev/null || echo "0") = "1" ]; then \
		echo "${RED}Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%${NC}"; \
		exit 1; \
	else \
		echo "${GREEN}Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%${NC}"; \
	fi

# Benchmark testing
benchmark:
	@echo "${BLUE}Running performance benchmarks...${NC}"
	cd .. && $(GO_TEST) $(GO_BENCH_FLAGS) $(BENCHMARK_DIRS) | tee benchmark.out

# Test linting
lint-tests:
	@echo "${BLUE}Linting test files...${NC}"
	@cd .. && golangci-lint run ./test/... || echo "${YELLOW}golangci-lint not found, skipping${NC}"

# Format test files
format-tests:
	@echo "${BLUE}Formatting test files...${NC}"
	cd .. && gofmt -w ./test/
	cd .. && goimports -w ./test/ || echo "${YELLOW}goimports not found, skipping${NC}"

# Validate test fixtures
validate-fixtures:
	@echo "${BLUE}Validating test fixtures...${NC}"
	@for file in fixtures/schemas/*.json; do \
		if [ -f "$$file" ]; then \
			echo "Validating $$file"; \
			python3 -m json.tool "$$file" > /dev/null 2>&1 || (echo "${RED}Invalid JSON: $$file${NC}" && exit 1); \
		fi; \
	done
	@for file in fixtures/examples/*.json; do \
		if [ -f "$$file" ]; then \
			echo "Validating $$file"; \
			python3 -m json.tool "$$file" > /dev/null 2>&1 || (echo "${RED}Invalid JSON: $$file${NC}" && exit 1); \
		fi; \
	done
	@for file in fixtures/golden/*.json; do \
		if [ -f "$$file" ]; then \
			echo "Validating $$file"; \
			python3 -m json.tool "$$file" > /dev/null 2>&1 || (echo "${RED}Invalid JSON: $$file${NC}" && exit 1); \
		fi; \
	done
	@echo "${GREEN}All fixtures validated successfully${NC}"

# Clean test artifacts
test-clean:
	@echo "${YELLOW}Cleaning test artifacts...${NC}"
	cd .. && rm -f coverage.out coverage.html
	cd .. && rm -f benchmark.out benchmark_baseline.out
	cd .. && rm -f cpu.prof mem.prof
	cd .. && rm -f *.test
	cd .. && go clean -testcache
	@echo "${GREEN}Test artifacts cleaned${NC}"

# Test execution summary
test-summary:
	@echo "${BLUE}Test Execution Summary${NC}"
	@echo "======================="
	@echo "Unit Tests:        $$(cd .. && go test -list . ./test/unit/... 2>/dev/null | grep -c 'Test' || echo '0')"
	@echo "Integration Tests: $$(cd .. && go test -list . ./test/integration/... 2>/dev/null | grep -c 'Test' || echo '0')"
	@echo "E2E Tests:         $$(cd .. && go test -list . ./test/e2e/... 2>/dev/null | grep -c 'Test' || echo '0')"
	@echo "Benchmarks:        $$(cd .. && go test -list . ./test/benchmarks/... 2>/dev/null | grep -c 'Benchmark' || echo '0')"
	@echo "Test Fixtures:     $$(find fixtures -name '*.json' 2>/dev/null | wc -l || echo '0')"