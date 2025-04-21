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

## Features

- **Full HAProxy Runtime API Support**: Comprehensive coverage of HAProxy's runtime API commands
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
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

When using only the stats page functionality, there's no need to define Runtime API parameters like host and port. You can use both Runtime API and Stats Page simultaneously for complementary capabilities, or use only one of them based on your environment's constraints.

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
| HAPROXY_STATS_ENABLED | Enable HAProxy stats page support | true |
| HAPROXY_STATS_URL | URL to HAProxy stats page (e.g., http://localhost:8404/stats) | http://127.0.0.1:8404/stats |
| MCP_TRANSPORT | MCP transport method (stdio/http) | stdio |
| MCP_PORT | Port for HTTP transport (when using http) | 8080 |
| LOG_LEVEL | Logging level (debug/info/warn/error) | info |

**Note:** You can use the Runtime API (TCP4 or Unix socket mode), the Stats API, or both simultaneously. At least one must be properly configured for the server to function.

## Security Considerations

- **Authentication**: Connect to HAProxy's Runtime API using secure methods
- **Network Security**: When using TCP4 mode, restrict connectivity to the Runtime API port
- **Unix Socket Permissions**: When using Unix socket mode, ensure proper socket file permissions
- **Input Validation**: All inputs are validated to prevent injection attacks

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

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Testing Locally

You can test the HAProxy MCP server locally in several ways:

### Direct CLI Testing

Build and run the server directly with environment variables:

```bash
# Build the server
go build -o bin/haproxy-mcp-server cmd/server/main.go

# Option 1: Test with TCP connection mode
HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 2: Test with Unix socket mode
HAPROXY_RUNTIME_MODE=unix HAPROXY_RUNTIME_SOCKET=/path/to/haproxy.sock LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 3: Test with Stats page integration
HAPROXY_STATS_ENABLED=true HAPROXY_STATS_URL="http://localhost:8404/stats" LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server

# Option 4: Test with both Runtime API and Stats page
HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 HAPROXY_STATS_ENABLED=true HAPROXY_STATS_URL="http://localhost:8404/stats" LOG_LEVEL=debug MCP_TRANSPORT=stdio ./bin/haproxy-mcp-server
```

### Test Individual MCP Tools

You can test specific MCP tools with JSON-RPC calls:

```bash
# Test show_info tool
echo '{"jsonrpc":"2.0","id":1,"method":"callTool","params":{"name":"show_info","arguments":{}}}' | HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 LOG_LEVEL=debug ./bin/haproxy-mcp-server

# Test show_stat tool
echo '{"jsonrpc":"2.0","id":2,"method":"callTool","params":{"name":"show_stat","arguments":{"filter":""}}}' | HAPROXY_HOST=<your-haproxy-host> HAPROXY_PORT=9999 HAPROXY_RUNTIME_MODE=tcp4 LOG_LEVEL=debug ./bin/haproxy-mcp-server
```

### Test with Claude or Other LLMs

Create a configuration file:

```bash
cat > haproxy-mcp-config.json << EOF
{
  "mcpServers": {
    "haproxy": {
      "command": "$(pwd)/bin/haproxy-mcp-server",
      "env": {
        "HAPROXY_HOST": "<your-haproxy-host>",
        "HAPROXY_PORT": "9999",
        "HAPROXY_RUNTIME_MODE": "tcp4",
        "MCP_TRANSPORT": "stdio",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
EOF
```

Then connect using a Claude-compatible client:

```bash
claude --mcp-servers-file ./haproxy-mcp-config.json
```

### Verify HAProxy Connection Separately

Before testing the MCP server, verify that your HAProxy setup is working correctly:

```bash
# For TCP connections
echo "show info" | socat tcp-connect:<your-haproxy-host>:9999 stdio

# For Unix socket connections
echo "show info" | socat unix-connect:/path/to/haproxy.sock stdio
```

This helps to confirm that your HAProxy instance is correctly configured to accept Runtime API commands.

## Stats Page Integration

The HAProxy MCP Server supports integration with HAProxy's statistics page, which provides a wealth of information about your HAProxy instance through a web interface.

### Benefits of Stats Page Integration

- **Enhanced Metrics**: Access to detailed metrics that may not be available through the Runtime API
- **Visualization**: Web-based view of HAProxy metrics for easier analysis
- **Fallback Mechanism**: Provides an alternative data source if the Runtime API is unavailable
- **Complementary Data**: Combines with Runtime API data for a more complete picture of your HAProxy instance

### Configuring the Stats Page in HAProxy

To enable the stats page in your HAProxy configuration:

```
frontend stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats admin if LOCALHOST
```

Then configure the HAProxy MCP Server to use this stats page by setting the appropriate environment variables:

```
HAPROXY_STATS_ENABLED=true
HAPROXY_STATS_URL=http://localhost:8404/stats
```
