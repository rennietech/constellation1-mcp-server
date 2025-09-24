# RESO MCP Server Installation Guide

## Quick Start

1. **Download the binary**: Use the provided `constellation1-mcp-server-darwin-arm64` binary for macOS ARM64 systems.

2. **Make it executable**:
   ```bash
   chmod +x constellation1-mcp-server-darwin-arm64
   ```

3. **Configure your MCP client** (e.g., Claude Desktop):
   
   Edit your MCP client configuration file (usually `~/Library/Application Support/Claude/claude_desktop_config.json` for Claude Desktop):
   
   ```json
   {
     "mcpServers": {
       "reso": {
         "command": "/path/to/constellation1-mcp-server-darwin-arm64",
         "args": [
           "-client-id", "your_reso_client_id_here",
           "-client-secret", "your_reso_client_secret_here"
         ]
       }
     }
   }
   ```

4. **Restart your MCP client** to load the new server.

## Configuration Options

### Command Line Arguments (Recommended)
Pass credentials directly as command line arguments:
```json
{
  "mcpServers": {
    "reso": {
      "command": "/path/to/constellation1-mcp-server-darwin-arm64",
      "args": [
        "-client-id", "your_client_id",
        "-client-secret", "your_client_secret"
      ]
    }
  }
}
```

### Environment Variables (Alternative)
Set these environment variables before running the server:
```bash
export RESO_CLIENT_ID="your_client_id"
export RESO_CLIENT_SECRET="your_client_secret"
export RESO_AUTH_URL="https://authenticate.constellation1apis.com/oauth2/token"
export RESO_BASE_URL="https://listings.cdatalabs.com/odata"
```

## Obtaining RESO API Credentials

To use this MCP server, you need valid RESO API credentials:

1. Contact your RESO API provider (Constellation Data Labs or other RESO-compliant provider)
2. Request client credentials for API access
3. You will receive:
   - `client_id`: Your unique client identifier
   - `client_secret`: Your secret key for authentication

## Verification

Once configured, you can verify the installation by:

1. Starting your MCP client
2. Looking for the "reso_query" tool in available tools
3. Testing with a simple query like:
   ```json
   {
     "entity": "Property",
     "top": 1,
     "select": "ListingKey,StandardStatus"
   }
   ```

## Troubleshooting

### Common Issues

1. **"Server not initialized" error**:
   - Check that your client_id and client_secret are correct
   - Verify your RESO API credentials are active

2. **"Authentication failed" error**:
   - Confirm your credentials with your RESO API provider
   - Check that the authentication URL is correct

3. **"Permission denied" error**:
   - Make sure the binary is executable: `chmod +x constellation1-mcp-server-darwin-arm64`

4. **"Command not found" error**:
   - Use the full path to the binary in your MCP client configuration

### Debug Mode

For debugging, you can run the server manually to see detailed logs:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"settings":{"client_id":"your_id","client_secret":"your_secret"}},"clientInfo":{"name":"test","version":"1.0"}}}' | ./constellation1-mcp-server-darwin-arm64
```

## Security Notes

- Keep your client_secret secure and never share it
- The server only stores credentials in memory during execution
- All communication with RESO API uses HTTPS
- Credentials are Base64 encoded as required by the RESO API specification

## Support

For issues related to:
- **RESO API access**: Contact your RESO API provider
- **MCP integration**: Check your MCP client documentation
- **Server functionality**: Review error messages and ensure proper configuration

