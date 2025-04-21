# HAProxy MCP Server

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

The HAProxy MCP Server exposes the following tools, grouped by feature area:

### Statistics & Process Info

#### show_stat
Retrieves the full statistics table for HAProxy.

**Arguments:**
- `filter` (optional): Filter by proxy or server names

**Example Response:**
```json
{
  "stats": [
    {
      "pxname": "web-frontend",
      "svname": "FRONTEND",
      "status": "OPEN",
      "sessions_current": 12,
      "bytes_in": 24560,
      "bytes_out": 145890
    },
    {
      "pxname": "web-backend",
      "svname": "server1",
      "status": "UP",
      "sessions_current": 5,
      "bytes_in": 12500,
      "bytes_out": 72945
    }
  ]
}
```

#### show_info
Displays HAProxy version, uptime, and process information.

**Arguments:** None

**Example Response:**
```json
{
  "info": {
    "version": "HAProxy version 2.6.6-9e7a86c 2023/05/28",
    "uptime": "0d 2h 35m 12s",
    "maxconn": 4000,
    "process_num": 1,
    "pid": 12345,
    "mode": "http"
  }
}
```

### Topology Discovery

#### show_frontend
Lists all frontends with their bind addresses, modes, and states.

**Arguments:** None

**Example Response:**
```json
{
  "frontends": [
    {
      "name": "web-frontend",
      "bind": "*:80",
      "mode": "http",
      "state": "OPEN"
    },
    {
      "name": "api-frontend",
      "bind": "*:8080",
      "mode": "http",
      "state": "OPEN"
    }
  ]
}
```

#### show_backend
Lists all backends and their configuration.

**Arguments:** None

**Example Response:**
```json
{
  "backends": [
    {
      "name": "web-backend",
      "mode": "http",
      "balance": "roundrobin",
      "server_count": 3
    },
    {
      "name": "api-backend",
      "mode": "http",
      "balance": "leastconn",
      "server_count": 2
    }
  ]
}
```

#### show_servers_state
Displays per-server state, current sessions, and weight.

**Arguments:**
- `backend` (optional): Backend name to filter results

**Example Response:**
```json
{
  "servers": [
    {
      "backend": "web-backend",
      "name": "server1",
      "address": "192.168.1.10:8080",
      "state": "UP",
      "sessions_current": 12,
      "weight": 100
    },
    {
      "backend": "web-backend",
      "name": "server2",
      "address": "192.168.1.11:8080",
      "state": "UP",
      "sessions_current": 8,
      "weight": 100
    }
  ]
}
```

### Dynamic Pool Management

#### add_server
Dynamically registers a new server in a backend.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name
- `address` (required): Server address (IP:port)
- `weight` (optional): Server weight

**Example Response:**
```json
{
  "success": true,
  "message": "Server 'server3' successfully added to backend 'web-backend'"
}
```

#### del_server
Removes a dynamic server from a backend.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name

**Example Response:**
```json
{
  "success": true,
  "message": "Server 'server3' successfully removed from backend 'web-backend'"
}
```

#### enable_server
Takes a server out of maintenance mode.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name

**Example Response:**
```json
{
  "success": true,
  "message": "Server 'server1' in backend 'web-backend' has been enabled"
}
```

#### disable_server
Puts a server into maintenance mode.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name

**Example Response:**
```json
{
  "success": true,
  "message": "Server 'server1' in backend 'web-backend' has been disabled"
}
```

#### set_weight
Changes a server's load-balancing weight.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name
- `weight` (required): New weight value

**Example Response:**
```json
{
  "success": true,
  "message": "Server 'server1' in backend 'web-backend' weight changed from 100 to 50"
}
```

### Session Control

#### show_sess
Lists all active sessions.

**Arguments:**
- `backend` (optional): Backend filter

**Example Response:**
```json
{
  "sessions": [
    {
      "id": "123456",
      "frontend": "web-frontend",
      "backend": "web-backend",
      "server": "server1",
      "age": "10s",
      "client_ip": "10.0.0.5"
    },
    {
      "id": "123457",
      "frontend": "api-frontend",
      "backend": "api-backend",
      "server": "server2",
      "age": "5s",
      "client_ip": "10.0.0.6"
    }
  ]
}
```

#### shutdown_session
Terminates a specific client session by ID.

**Arguments:**
- `session_id` (required): Session ID

**Example Response:**
```json
{
  "success": true,
  "message": "Session 123456 has been terminated"
}
```

### Health Checks & Agents

#### enable_health
Enables active health checks on a server.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name

**Example Response:**
```json
{
  "success": true,
  "message": "Health checks enabled for server 'server1' in backend 'web-backend'"
}
```

#### disable_health
Disables active health checks on a server.

**Arguments:**
- `backend` (required): Backend name
- `server_name` (required): Server name

**Example Response:**
```json
{
  "success": true,
  "message": "Health checks disabled for server 'server1' in backend 'web-backend'"
}
```

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
