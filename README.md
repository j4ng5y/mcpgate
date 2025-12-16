# MCPGate - Local MCP Server Gateway

MCPGate is a local, dockerless MCP (Model Context Protocol) gateway that provides unified access to multiple upstream MCP servers. It acts as an MCP server itself, aggregating connections to various MCP servers and providing centralized request routing, connection management, and server discovery.

## Features

- **Multiple Transport Support**: Connect to MCP servers via stdio (subprocess), HTTP, WebSocket, or Unix sockets
- **Server Registry & Discovery**: Automatic registration and discovery of available MCP servers
- **Connection Pooling**: Efficient connection reuse and management with health monitoring
- **Request Routing**: Intelligent routing of requests to appropriate upstream servers
- **TOML Configuration**: Simple, readable configuration format for server definitions
- **Multiple Client Connections**: Support various client connection methods (stdout/stdio as default)
- **Production Ready**: Built with Go for performance and reliability
- **Optimized Binaries**: Uses GoReleaser with UPX compression for smaller artifacts

## Installation

### From Source

```bash
git clone https://github.com/j4ng5y/mcpgate.git
cd mcpgate

# Build for development
make dev

# Or build release binaries (requires goreleaser)
make release-snapshot
```

### Prerequisites

- Go 1.25.5 or later
- goreleaser (for building releases with UPX compression)
- UPX (optional, for binary compression)

## Configuration

MCPGate uses TOML format for configuration. See `example/config.toml` for a complete example.

### Basic Configuration

```toml
[gateway]
log_level = "info"

[[server]]
name = "bedrock"
transport = "stdio"
enabled = true
command = "node"
args = ["./server.js"]
timeout = 30
```

### Server Configuration

Each upstream MCP server can be configured with:

- **name**: Unique identifier for the server
- **transport**: Connection type (`stdio`, `http`, `websocket`, `unix`)
- **enabled**: Whether to start this server
- **command**: (stdio) Command to execute
- **args**: (stdio) Command arguments
- **env**: (stdio) Environment variables
- **url**: (http/websocket) Remote server URL
- **socket_path**: (unix) Path to Unix socket
- **timeout**: Request timeout in seconds
- **metadata**: Custom metadata (key-value pairs)

### Transport Types

#### Stdio (Default)
Spawns and communicates with a subprocess:

```toml
[[server]]
name = "local-tool"
transport = "stdio"
command = "python"
args = ["-m", "mcp_server_filesystem"]

[server.env]
PYTHONUNBUFFERED = "1"
```

#### HTTP
Connects to remote HTTP/JSON-RPC endpoints:

```toml
[[server]]
name = "remote-api"
transport = "http"
url = "http://api.example.com:8000"
timeout = 30
```

#### WebSocket
Real-time WebSocket connections:

```toml
[[server]]
name = "realtime"
transport = "websocket"
url = "ws://localhost:9000"
```

#### Unix Socket
Local Unix domain socket communication:

```toml
[[server]]
name = "local-socket"
transport = "unix"
socket_path = "/tmp/mcp-server.sock"
```

## Usage

### Running the Gateway

```bash
# Using default config
./bin/mcpgate

# Using custom config
./bin/mcpgate -config /path/to/config.toml

# Or with shorthand
./bin/mcpgate -c /path/to/config.toml
```

### Gateway-Specific Methods

While acting as an MCP server, MCPGate provides special gateway management methods:

#### List All Servers

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "gateway/list_servers",
  "params": {}
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "name": "bedrock",
      "connected": true,
      "initialized": true,
      "transport": "stdio",
      "capabilities": ["tools"]
    }
  ]
}
```

#### Get Server Details

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "gateway/get_server",
  "params": {"name": "bedrock"}
}
```

#### Check Server Status

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "gateway/server_status",
  "params": {"name": "bedrock"}
}
```

#### List Capabilities

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "gateway/capabilities",
  "params": {}
}
```

### Routing Requests to Specific Servers

Include `_server` parameter to route to a specific server:

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/list",
  "params": {"_server": "bedrock"}
}
```

Without explicit server specification, MCPGate uses intelligent routing:
- Attempts to route based on method prefix (e.g., `tools/list` → tools capability)
- Falls back to first available server if no specific capability match
- Returns error if no servers are available

## Building

### Development Build

```bash
make dev
# Creates: bin/mcpgate
```

### Release Build (with compression)

```bash
make release
# Creates optimized, UPX-compressed binaries in dist/
# Supports: Linux, macOS, Windows (amd64, arm64, arm)
```

### Build Snapshot

```bash
make release-snapshot
# Quick build without publishing, for testing
```

## Development

### Run with Example Config

```bash
make run
```

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Lint

```bash
make lint
```

### Clean Build Artifacts

```bash
make clean
```

## Architecture

### Components

```
┌─────────────────────────────────────────┐
│         MCP Clients                     │
│  (Communicates via stdio/JSON-RPC)      │
└─────────────┬───────────────────────────┘
              │
              ▼
    ┌─────────────────────┐
    │   MCPGate Gateway   │
    │  (Acts as MCP Server)
    │                     │
    ├─ Configuration     │
    ├─ Request Router    │
    ├─ Server Registry   │
    └─────────────────────┘
         │    │    │
    ┌────┘    │    └────┐
    ▼         ▼         ▼
  Upstream  Upstream  Upstream
  Server 1  Server 2  Server 3
```

### Key Modules

- **config**: TOML configuration parsing
- **transport**: Abstract transport layer (stdio, HTTP, WebSocket, Unix socket)
- **server**: Managed server lifecycle and registry
- **mcp**: MCP protocol handling and request routing
- **pool**: Connection pooling and management

## Connection Management

The gateway manages connections to upstream servers with:

- **Automatic Connection Establishment**: Connects on startup with retries
- **Health Monitoring**: Tracks connection health and availability
- **Connection Pooling**: Reuses connections efficiently
- **Automatic Reconnection**: Detects disconnections and rebuilds connections
- **Timeout Management**: Configurable timeouts per server

## Error Handling

MCPGate provides detailed error responses following JSON-RPC 2.0 specification:

- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid parameters
- `-32603`: Internal error
- `-32000` to `-32099`: Server errors

## Performance Considerations

- **Binary Size**: UPX-compressed binaries are typically 60-70% smaller
- **Memory**: Connection pooling reduces overhead
- **Concurrency**: Thread-safe design for high-concurrency scenarios
- **Timeouts**: All external operations have configurable timeouts

## Troubleshooting

### Server Connection Issues

Check the logs for connection errors:

```bash
./bin/mcpgate -c config.toml
```

Ensure upstream servers are accessible:
- For stdio: Command exists and is executable
- For HTTP: URL is reachable
- For WebSocket: URL is accessible
- For Unix socket: Socket file exists and is readable

### Routing Issues

If requests aren't routing correctly:

1. Check `gateway/list_servers` to see connected servers
2. Explicitly specify `_server` parameter for testing
3. Review method prefix matching in router

## CI/CD Pipeline

MCPGate uses GitHub Actions for automated testing, linting, and releases:

### Testing Workflow (`.github/workflows/test.yml`)

Runs on every push to `main` and all pull requests:

- **Matrix Testing**: Tests on Go 1.25 and 1.26 across Ubuntu, macOS, and Windows
- **Unit Tests**: `go test -v -race` with coverage reporting
- **Linting**: `golangci-lint` with automatic checks
- **Build Verification**: Snapshot builds to ensure artifacts are created
- **Coverage Upload**: Sends coverage reports to Codecov

Run locally:
```bash
make test      # Run unit tests
make lint      # Run linters
make build     # Build snapshot
```

### Release Workflow (`.github/workflows/release.yml`)

Automatically triggered when a git tag is pushed (format: `v*`):

- **Pre-release Testing**: Runs all tests before building
- **Multi-platform Builds**: Creates binaries for Linux, macOS, Windows (amd64, arm64, arm)
- **Binary Compression**: Uses UPX compression for smaller artifacts
- **Checksums**: Generates SHA256 checksums for all artifacts
- **GitHub Release**: Creates a GitHub release with artifacts and auto-generated changelog

Create a release:
```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions will automatically build and publish the release
```

### Dependency Management

Dependabot automatically:
- Creates pull requests for Go module updates (weekly)
- Updates GitHub Actions to latest versions (weekly)
- Labels PRs for easy organization

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Ensure all checks pass:
   - `make test` - Unit tests pass
   - `make lint` - Linting passes
   - `make fmt` - Code is formatted
6. Submit a pull request (use the PR template)

## License

[Add your license here]

## Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

**MCPGate** - Unified MCP Server Management Made Simple
