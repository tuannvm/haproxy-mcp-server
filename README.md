# HAProxy MCP Server

[![Build](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/build.yml/badge.svg)](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/build.yml)
[![Release](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/release.yml/badge.svg)](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tuannvm/haproxy-mcp-server)](https://goreportcard.com/report/github.com/tuannvm/haproxy-mcp-server)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tuannvm/haproxy-mcp-server)
![License](https://img.shields.io/github/license/tuannvm/haproxy-mcp-server)
![Docker Pulls](https://img.shields.io/docker/pulls/tuannvm/haproxy-mcp-server)

A Model Context Protocol (MCP) server for HAProxy implemented in Go, leveraging HAProxy Runtime API and mcp-go.

## Overview

The HAProxy MCP Server provides a standardized way for LLMs to interact with HAProxy's runtime API via the Model Context Protocol (MCP). This enables LLMs to perform HAProxy administration tasks, monitor server status, manage backend servers, and analyze traffic patterns, all through natural language interfaces.

![Screenshot-1](https://github.com/user-attachments/assets/8d3bf7f5-be98-4997-b676-120891692f15)
![Screenshot-2](https://github.com/user-attachments/assets/a443b5d2-8d7d-4daf-a6c1-115912f704d1)
![Screenshot-3](https://github.com/user-attachments/assets/fa387604-4eb9-4456-adc3-5a8395e5ecc1)

## Features

- **Full HAProxy Runtime API Support**: Comprehensive coverage of HAProxy's runtime API commands
- **Context-Aware Operations**: All operations support proper timeout and cancellation handling
- **Stats Page Integration**: Support for HAProxy's web-based statistics page for enhanced metrics and visualization
- **Secure Authentication**: Support for secure connections to HAProxy runtime API
- **Multiple Transport Options**: Supports both stdio and HTTP transports for flexibility in different environments
- **Enterprise Ready**: Designed for production use in enterprise environments
- **Docker Support**: Pre-built Docker images for easy deployment

## Installation

### Homebrew

```bash
# Add the tap
brew tap tuannvm/tap

# Install the package
brew install haproxy-mcp-server
```

### From Binary

Download the latest binary for your platform from the [releases page](https://github.com/tuannvm/haproxy-mcp-server/releases).

### Using Go

```bash
go install github.com/tuannvm/haproxy-mcp-server/cmd/server@latest
```

### Using Docker

```bash
docker pull ghcr.io/tuannvm/haproxy-mcp-server:latest
docker run -it --rm \
  -e HAPROXY_HOST=your-haproxy-host \
  -e HAPROXY_PORT=9999 \
  ghcr.io/tuannvm/haproxy-mcp-server:latest
```

## MCP Integration

To use this server with MCP-compatible LLMs, configure the assistant with the following connection details:

### HAProxy Runtime API over TCP4:

```json
{
  "mcpServers": {
    "haproxy": {
      "command": "haproxy-mcp-server",
      "env": {
        "HAPROXY_HOST": "localhost",
        "HAPROXY_PORT": "9999",
        "HAPROXY_RUNTIME_MODE": "tcp4",
        "HAPROXY_RUNTIME_TIMEOUT": "10",
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

### HAProxy Runtime API over Unix Socket:

```json
{
  "mcpServers": {
    "haproxy": {
      "command": "haproxy-mcp-server",
      "env": {
        "HAPROXY_RUNTIME_MODE": "unix",
        "HAPROXY_RUNTIME_SOCKET": "/var/run/haproxy/admin.sock",
        "HAPROXY_RUNTIME_TIMEOUT": "10",
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

### HAProxy with Stats Page Support:

```json
{
  "mcpServers": {
    "haproxy": {
      "command": "haproxy-mcp-server",
      "env": {
        "HAPROXY_STATS_ENABLED": "true",
        "HAPROXY_STATS_URL": "http://localhost:8404/stats",
        "HAPROXY_STATS_TIMEOUT": "5",
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

When using only the stats page functionality, there's no need to define Runtime API parameters like host and port. You can use both Runtime API and Stats Page simultaneously for complementary capabilities, or use only one of them based on your environment's constraints.

> **Note:** For detailed instructions on how to configure HAProxy to expose the Runtime API and Statistics page, see the [HAProxy Configuration Guide](haproxy.md).

## Available MCP Tools

The HAProxy MCP Server exposes tools that map directly to HAProxy's Runtime API commands, organized into the following categories:

- **Statistics & Process Info**: Retrieve statistics, server information, and manage counters
- **Topology Discovery**: List frontends, backends, server states, and configuration details
- **Dynamic Pool Management**: Add, remove, enable/disable servers and adjust their properties
- **Session Control**: View and manage active sessions
- **Maps & ACLs**: Manage HAProxy maps and ACL files
- **Health Checks & Agents**: Control health checks and agent-based monitoring
- **Miscellaneous**: View errors, run echo tests, and get help information

For a complete list of all supported tools with their inputs, outputs, and corresponding HAProxy Runtime API commands, see the [tools.md](tools.md) documentation.

## Configuration

The server can be configured using the following environment variables:

| Variable | Description | Default |
| --- | --- | --- |
| HAPROXY_HOST | Host of the HAProxy instance (TCP4 mode only) | 127.0.0.1 |
| HAPROXY_PORT | Port for the HAProxy Runtime API (TCP4 mode only) | 9999 |
| HAPROXY_RUNTIME_MODE | Connection mode: "tcp4" or "unix" | tcp4 |
| HAPROXY_RUNTIME_SOCKET | Socket path (Unix mode only) | /var/run/haproxy/admin.sock |
| HAPROXY_RUNTIME_URL | Direct URL to Runtime API (optional, overrides other runtime settings) | |
| HAPROXY_RUNTIME_TIMEOUT | Timeout for runtime API operations in seconds | 10 |
| HAPROXY_STATS_ENABLED | Enable HAProxy stats page support | true |
| HAPROXY_STATS_URL | URL to HAProxy stats page (e.g., http://localhost:8404/stats) | http://127.0.0.1:8404/stats |
| HAPROXY_STATS_TIMEOUT | Timeout for stats page operations in seconds | 5 |
| MCP_TRANSPORT | MCP transport method (stdio/http) | stdio |
| MCP_PORT | Port for HTTP transport (when using http) | 8080 |
| LOG_LEVEL | Logging level (debug/info/warn/error) | info |

**Note:** You can use the Runtime API (TCP4 or Unix socket mode), the Stats API, or both simultaneously. At least one must be properly configured for the server to function.

## Security Considerations

- **Authentication**: Connect to HAProxy's Runtime API using secure methods
- **Network Security**: When using TCP4 mode, restrict connectivity to the Runtime API port
- **Unix Socket Permissions**: When using Unix socket mode, ensure proper socket file permissions
- **Input Validation**: All inputs are validated to prevent injection attacks

For comprehensive security best practices and configuration examples, see the [HAProxy Configuration Guide](haproxy.md#security-considerations).

## Development

### Testing

```bash
# Run all tests
go test ./...

# Run tests excluding integration tests
go test -short ./...

# Run integration tests with specific HAProxy instance
export HAPROXY_HOST="your-haproxy-host"
export HAPROXY_PORT="9999"
go test ./internal/haproxy -v -run Test
```

You can test the HAProxy MCP server locally in several ways:

#### Direct CLI Testing

Build and run the server directly with environment variables:

```bash
# Build the server
go build -o bin/haproxy-mcp-server cmd/server/main.go

# Option 1: Test with TCP connection mode
HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 HAPROXY_RUNTIME_TIMEOUT=10 LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 2: Test with Unix socket mode
HAPROXY_RUNTIME_MODE=unix HAPROXY_RUNTIME_SOCKET=/path/to/haproxy.sock HAPROXY_RUNTIME_TIMEOUT=10 LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 3: Test with Stats page integration
HAPROXY_STATS_ENABLED=true HAPROXY_STATS_URL="http://localhost:8404/stats" HAPROXY_STATS_TIMEOUT=5 LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 4: Test with both Runtime API and Stats page
HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 HAPROXY_RUNTIME_TIMEOUT=10 HAPROXY_STATS_ENABLED=true HAPROXY_STATS_URL="http://localhost:8404/stats" HAPROXY_STATS_TIMEOUT=5 LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server
```

#### Test Individual MCP Tools

You can test specific MCP tools with JSON-RPC calls:

```bash
# Test show_info tool
echo '{"jsonrpc":"2.0","id":1,"method":"callTool","params":{"name":"show_info","arguments":{}}}' | HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 LOG_LEVEL=debug ./bin/haproxy-mcp-server

# Test show_stat tool
echo '{"jsonrpc":"2.0","id":2,"method":"callTool","params":{"name":"show_stat","arguments":{"filter":""}}}' | HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 LOG_LEVEL=debug ./bin/haproxy-mcp-server
```

### Technical Implementation

The HAProxy MCP Server includes several technical improvements designed for reliability and robustness:

- **Context-Aware Operations**: All API calls support context-based timeout and cancellation, allowing graceful termination of long-running operations.
- **Fallback Mechanisms**: Automatic fallback to socat if direct connection fails, ensuring compatibility across different HAProxy deployments.
- **Unified Socket Handling**: Common code for both TCP and Unix socket connections, reducing duplication and improving maintainability.
- **Resilient Connection Management**: Dynamic buffer management for large responses and proper resource cleanup with deadline handling.
- **Comprehensive Error Handling**: Structured error handling and logging for easier troubleshooting.

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
