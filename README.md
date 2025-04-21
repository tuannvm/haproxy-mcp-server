# HAProxy MCP Server

[![Build](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/build.yml/badge.svg)](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/build.yml)
[![Release](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/release.yml/badge.svg)](https://github.com/tuannvm/haproxy-mcp-server/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tuannvm/haproxy-mcp-server)](https://goreportcard.com/report/github.com/tuannvm/haproxy-mcp-server)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tuannvm/haproxy-mcp-server)
![License](https://img.shields.io/github/license/tuannvm/haproxy-mcp-server)
![Docker Pulls](https://img.shields.io/docker/pulls/tuannvm/haproxy-mcp-server)

A Model Context Protocol (MCP) server for HAProxy implemented in Go, leveraging HAProxytech/client-native and mcp-go.

## Overview

The HAProxy MCP Server provides a standardized way for AI assistants to interact with HAProxy's runtime API via the Model Context Protocol (MCP). This enables AI assistants to perform HAProxy administration tasks, monitor server status, manage backend servers, and analyze traffic patterns, all through natural language interfaces.

## Features

- **Full HAProxy Runtime API Support**: Comprehensive coverage of HAProxy's runtime API commands
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

To use this server with MCP-compatible AI assistants, configure the assistant with the following connection details:

### For HAProxy Runtime API over TCP4:

```json
{
  "tools": [
    {
      "type": "mcp",
      "mcp": {
        "transport": "stdio",
        "binary": "/path/to/haproxy-mcp-server",
        "env": {
          "HAPROXY_HOST": "your-haproxy-host",
          "HAPROXY_PORT": "9999",
          "HAPROXY_RUNTIME_MODE": "tcp4"
        }
      }
    }
  ]
}
```

### For HAProxy Runtime API over Unix Socket:

```json
{
  "tools": [
    {
      "type": "mcp",
      "mcp": {
        "transport": "stdio",
        "binary": "/path/to/haproxy-mcp-server",
        "env": {
          "HAPROXY_RUNTIME_MODE": "unix",
          "HAPROXY_RUNTIME_SOCKET": "/var/run/haproxy/admin.sock"
        }
      }
    }
  ]
}
```

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
| HAPROXY_HOST | Host of the HAProxy instance | 127.0.0.1 |
| HAPROXY_PORT | Port for the HAProxy Runtime API | 9999 |
| HAPROXY_RUNTIME_MODE | Connection mode: "tcp4" or "unix" | tcp4 |
| HAPROXY_RUNTIME_SOCKET | Socket path (used only with unix mode) | /var/run/haproxy/admin.sock |
| MCP_TRANSPORT | MCP transport method (stdio/http) | stdio |
| MCP_PORT | Port for HTTP transport (when using http) | 8080 |
| LOG_LEVEL | Logging level (debug/info/warn/error) | info |

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
