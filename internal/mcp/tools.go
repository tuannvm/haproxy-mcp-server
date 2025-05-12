package mcp

import (
    "log/slog"

    "github.com/mark3labs/mcp-go/server"
    "github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func RegisterTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
    slog.Info("Registering HAProxy MCP tools...")
    registerStatTools(s, client)
    registerBackendTools(s, client)
    registerServerTools(s, client)
    registerHealthAgentTools(s, client)
    registerReloadTool(s, client)
    slog.Info("All HAProxy MCP tools registered successfully")
}