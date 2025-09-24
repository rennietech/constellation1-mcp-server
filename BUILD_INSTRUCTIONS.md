# RESO MCP Server - Build Instructions for macOS

## Prerequisites

1. **Go Programming Language** (version 1.21 or later)
   ```bash
   # Check if Go is installed
   go version

   # If not installed, download from: https://golang.org/dl/
   # Or install via Homebrew:
   brew install go
   ```

2. **Git** (usually pre-installed on macOS)
   ```bash
   git --version
   ```

## Build Steps

### 1. Extract and Navigate
```bash
# Extract the zip file and navigate to the directory
cd constellation1-mcp-server
```

### 2. Initialize Go Module
```bash
# Download dependencies
go mod tidy
```

### 3. Build for macOS
```bash
# Build for your current architecture (Intel or Apple Silicon)
CGO_ENABLED=0 go build -a -installsuffix cgo -o constellation1-mcp-server .

# Or build specifically for Apple Silicon (M1/M2/M3)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -installsuffix cgo -o constellation1-mcp-server-arm64 .

# Or build specifically for Intel Macs
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o constellation1-mcp-server-amd64 .

# Alternatively, use the Makefile (recommended):
make build          # For current platform
make build-arm64    # For Apple Silicon
make build-amd64    # For Intel Macs
make build-all      # For all macOS architectures
```

### 4. Make Executable
```bash
chmod +x constellation1-mcp-server*
```

### 5. Test the Build
```bash
# Test that the binary works
./constellation1-mcp-server -h
```

## Build Flags Explained

- `CGO_ENABLED=0`: Disables CGO for better compatibility and static linking
- `-a`: Force rebuilding of all packages
- `-installsuffix cgo`: Ensures proper static linking

**Note for Beta macOS Users**: If you're running a beta version of macOS (like macOS 15.0 or later betas), the `CGO_ENABLED=0` approach is required to avoid `LC_UUID load command` errors. The updated build process creates statically linked binaries that work reliably on beta macOS versions.

## Cross-Platform Building

If you want to build for other platforms:

```bash
# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o constellation1-mcp-server-linux .

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o constellation1-mcp-server.exe .
```

## Configuration

After building, configure your MCP client (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "reso": {
      "command": "/full/path/to/constellation1-mcp-server",
      "args": [
        "-client-id", "your_reso_client_id",
        "-client-secret", "your_reso_client_secret"
      ]
    }
  }
}
```

## Troubleshooting

### Go Not Found
```bash
# Add Go to your PATH (add to ~/.zshrc or ~/.bash_profile)
export PATH=$PATH:/usr/local/go/bin
```

### Permission Denied
```bash
# Make sure the binary is executable
chmod +x constellation1-mcp-server
```

### Module Download Issues
```bash
# Clear module cache and retry
go clean -modcache
go mod tidy
```

### LC_UUID Load Command Error (Beta macOS)
If you encounter the error:
```
dyld[xxxx]: missing LC_UUID load command in /path/to/constellation1-mcp-server
```

This typically happens on beta versions of macOS. Solution:
```bash
# Clean and rebuild with the updated Makefile
make clean
make build-arm64  # or make build for current platform

# Or build manually:
CGO_ENABLED=0 go build -a -installsuffix cgo -o constellation1-mcp-server .
```

## Development

For development and testing:

```bash
# Run without building
go run . -client-id="test" -client-secret="test"

# Run tests (if any)
go test ./...

# Format code
go fmt ./...
```

## File Structure

```
constellation1-mcp-source-final/
├── main.go              # Main server implementation
├── go.mod               # Go module definition
├── auth/                # OAuth2 authentication
│   └── oauth.go
├── api/                 # RESO API client
│   ├── client.go
│   └── types.go
├── config/              # Configuration management
│   └── config.go
├── tools/               # MCP tool implementation
│   └── reso_query.go
├── README.md            # Usage documentation
├── INSTALLATION.md      # Installation guide
└── BUILD_INSTRUCTIONS.md # This file
```

## Support

- For Go installation issues: https://golang.org/doc/install
- For RESO API credentials: Contact your RESO API provider
- For MCP client setup: Check your MCP client documentation

