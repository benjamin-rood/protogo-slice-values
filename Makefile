.PHONY: build install clean test test-unit test-integration example

# Build the plugin
build:
	go build -o protoc-gen-go-value-slices ./cmd/protoc-gen-go-value-slices

# Install to GOBIN (or GOPATH/bin)
install:
	go install ./cmd/protoc-gen-go-value-slices

# Clean build artifacts
clean:
	rm -f protoc-gen-go-value-slices
	rm -rf testdata/gen

# Run all tests
test: test-unit test-integration

# Run unit tests only
test-unit:
	go test ./internal/...

# Run integration tests (requires protoc and protoc-gen-go)
test-integration:
	go test -tags=integration .

# Build and run example (if example files exist)
example: build
	@echo "Building example requires .proto files in examples/ directory"
	@if [ -d "examples" ]; then \
		cd examples && buf generate; \
	fi

# Check if required dependencies are available
check-deps:
	@echo "Checking dependencies..."
	@which protoc > /dev/null || (echo "protoc not found. Please install Protocol Buffers compiler." && exit 1)
	@which protoc-gen-go > /dev/null || (echo "protoc-gen-go not found. Please install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@echo "All dependencies found!"

# Development setup
dev-setup: check-deps install

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the plugin binary"
	@echo "  install         - Install the plugin to GOBIN"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run all tests (unit + integration)"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests (requires protoc)"
	@echo "  example         - Build and run example"
	@echo "  check-deps      - Check if required dependencies are installed"
	@echo "  dev-setup       - Set up development environment"
	@echo "  help            - Show this help message"