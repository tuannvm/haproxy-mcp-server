# HAProxy MCP Server Project Plan

## Overview

This document outlines the development plan for the `haproxy-mcp-server` Go project, leveraging the `haproxytech/client-native` SDK. The server will follow the Model Context Protocol (MCP) standard, similar to other MCP server implementations.

## Project Goals

- Develop a Go-based MCP server for HAProxy that allows interaction with HAProxy's runtime API
- Provide a standardized way for AI assistants to manage HAProxy configurations
- Support both TCP4 and Unix socket connections to HAProxy
- Ensure enterprise-grade security and reliability

## Key Design Principles

1. **API Focus**: The server will strictly interact with HAProxy through the Runtime API only.
2. **Standard Logging**: The project will use Go's standard library `log/slog` for logging.
3. **Configuration Flexibility**: Support for both TCP4 and Unix socket connections to HAProxy.
4. **Security**: Implement secure handling of connections and proper input validation.

## Core Components

### 1. Project Structure

The project will follow a standard Go project layout:
- `cmd/server`: Entry point for the application
- `internal/haproxy`: HAProxy client wrapper and implementation
- `internal/mcp/tools`: MCP tool definitions and implementations
- `internal/config`: Configuration management

### 2. HAProxy Client Wrapper

A client wrapper will abstract the HAProxy Runtime API communications:
- Support for both TCP4 connections and Unix socket access
- Error handling and retry mechanisms
- Method wrappers for all needed Runtime API operations

### 3. MCP Tools Implementation

MCP tools will expose HAProxy operations through the Model Context Protocol:

#### 3.1 Statistics & Process Info
- Runtime statistics retrieval
- Version and process information
- Counters management

#### 3.2 Topology Discovery
- Frontend, backend, and server discovery
- Configuration exploration
- Map and table inspection

#### 3.3 Dynamic Pool Management
- Server addition and removal
- Server enabling and disabling
- Weight and connection limit adjustments

#### 3.4 Session Control
- Session listing and management
- Session termination capabilities

#### 3.5 Maps & ACLs
- Map file management
- ACL list operations

#### 3.6 Health Checks & Agents
- Health check control
- Agent-based monitoring configuration

#### 3.7 Miscellaneous
- Error reporting
- Echo testing
- Help documentation

### 4. Server Implementation

The MCP server will handle the protocol communications:
- Support for both stdio and HTTP transports
- Tool registration and dispatch
- Error handling and logging

### 5. Configuration

Configuration will be handled through environment variables:
- `HAPROXY_HOST`: Host of the HAProxy instance
- `HAPROXY_PORT`: Port for the HAProxy Runtime API
- `HAPROXY_RUNTIME_MODE`: "tcp4" or "unix"
- `HAPROXY_RUNTIME_SOCKET`: Unix socket path
- `MCP_TRANSPORT`: Transport method (stdio/http)
- `MCP_PORT`: HTTP port for server
- `LOG_LEVEL`: Logging verbosity

### 6. Build and Packaging

The project will include:
- Makefile for common development operations
- Docker support for containerized deployment
- CI/CD configuration for automated builds and testing

### 7. Testing Strategy

Testing will focus on:
- Unit tests for all components
- Integration tests for the HAProxy client
- End-to-end tests for the MCP server

### 8. Documentation

Documentation will include:
- README with installation and usage instructions
- Configuration documentation
- MCP tools reference
- Integration examples

### 9. CI/CD Pipeline

The CI/CD pipeline will automate:
- Code linting and testing
- Security scanning
- Binary builds for multiple platforms
- Docker image building and publishing
- Release management

## Release Strategy

The project will follow semantic versioning, with releases published to:
- GitHub Releases
- Docker Hub/GitHub Container Registry
- Homebrew tap
