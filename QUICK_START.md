# Quick Start Guide

## TL;DR - Build in 3 Commands

```bash
# 1. Navigate to source directory
cd constellation1-mcp-source-final

# 2. Build (requires Go 1.21+)
make build

# 3. Configure in your MCP client
# Add this to your MCP client config:
```

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

## If You Don't Have Go

1. Install Go: https://golang.org/dl/
2. Or via Homebrew: `brew install go`
3. Then run the build commands above

## Available Make Commands

- `make build` - Build for your Mac
- `make build-arm64` - Build for Apple Silicon specifically
- `make build-amd64` - Build for Intel Macs specifically
- `make clean` - Remove build files
- `make help` - Show all commands

## Manual Build (without Make)

```bash
go mod tidy
go build -ldflags="-s -w" -o constellation1-mcp-server .
chmod +x constellation1-mcp-server
```

That's it! 🚀

