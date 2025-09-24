# RESO MCP Server Makefile

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	go mod tidy
	CGO_ENABLED=0 go build -a -installsuffix cgo -o reso-mcp-server .
	chmod +x reso-mcp-server
	@echo "✅ Built reso-mcp-server for current platform"

# Build for Apple Silicon (M1/M2/M3)
.PHONY: build-arm64
build-arm64:
	go mod tidy
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -installsuffix cgo -o reso-mcp-server-arm64 .
	chmod +x reso-mcp-server-arm64
	@echo "✅ Built reso-mcp-server-arm64 for Apple Silicon"

# Build for Intel Macs
.PHONY: build-amd64
build-amd64:
	go mod tidy
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o reso-mcp-server-amd64 .
	chmod +x reso-mcp-server-amd64
	@echo "✅ Built reso-mcp-server-amd64 for Intel Macs"

# Build for all macOS architectures
.PHONY: build-all
build-all: build-arm64 build-amd64
	@echo "✅ Built for all macOS architectures"

# Clean build artifacts
.PHONY: clean
clean:
	rm -f reso-mcp-server*
	@echo "🧹 Cleaned build artifacts"

# Test the build
.PHONY: test
test: build
	./reso-mcp-server -h
	@echo "✅ Build test completed"

# Development run
.PHONY: run
run:
	go run . -client-id="test" -client-secret="test"

# Format code
.PHONY: fmt
fmt:
	go fmt ./...
	@echo "✅ Code formatted"

# Show help
.PHONY: help
help:
	@echo "RESO MCP Server Build Commands:"
	@echo ""
	@echo "  make build      - Build for current platform"
	@echo "  make build-arm64 - Build for Apple Silicon (M1/M2/M3)"
	@echo "  make build-amd64 - Build for Intel Macs"
	@echo "  make build-all  - Build for all macOS architectures"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Test the build"
	@echo "  make run        - Run in development mode"
	@echo "  make fmt        - Format source code"
	@echo "  make help       - Show this help"

