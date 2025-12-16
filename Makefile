.PHONY: help build test test-coverage-html test-integration test-all clean release lint fmt run mod-tidy all

help:
	@echo "MCPGate - MCP Server Gateway"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build snapshot using goreleaser (RECOMMENDED)"
	@echo "  release            Build and release using goreleaser (requires git tag)"
	@echo "  test               Run unit tests with coverage"
	@echo "  test-coverage-html Generate HTML coverage report"
	@echo "  test-integration   Run integration tests"
	@echo "  test-all           Run all tests (unit + integration)"
	@echo "  lint               Run golangci-lint"
	@echo "  fmt                Format code"
	@echo "  clean              Remove build artifacts"
	@echo "  run                Run mcpgate with example config"
	@echo ""
	@echo "Note: All builds use GoReleaser with UPX compression for optimal binaries"

build:
	@echo "Building snapshot binaries with GoReleaser..."
	@goreleaser build --snapshot --clean
	@echo "✓ Snapshot binaries created in dist/"
	@echo "Binary locations:"
	@find dist -name "mcpgate" -type f 2>/dev/null | sed 's/^/  - /'

release:
	@echo "Building and releasing with GoReleaser..."
	@echo "Note: This requires a git tag"
	@goreleaser release --clean
	@echo "✓ Release binaries created in dist/"

clean:
	@echo "Cleaning..."
	@rm -rf bin/ dist/ *.o
	@go clean
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage report:"
	@go tool cover -func=coverage.out | grep total | awk '{print "  Total coverage: " $$3}'
	@echo "✓ Tests complete"

test-coverage-html: test
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

test-integration:
	@echo "Running integration tests..."
	@go test -v -race -tags=integration ./...
	@echo "✓ Integration tests complete"

test-all: test test-integration
	@echo "✓ All tests complete"

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w .
	@echo "✓ Format complete"

run: build
	@echo "Running mcpgate..."
	@# Find the binary for current OS/ARCH
	@BINARY=$$(find dist -name "mcpgate" -type f | grep -E "(linux|darwin|windows)-(amd64|arm64|386)" | head -1); \
	if [ -z "$$BINARY" ]; then \
		echo "Error: No binary found in dist/"; \
		exit 1; \
	fi; \
	$$BINARY -config example/config.toml

mod-tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Dependencies updated"

all: clean build test
	@echo "✓ All done!"
