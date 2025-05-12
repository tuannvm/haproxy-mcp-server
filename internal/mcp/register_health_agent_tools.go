package mcp

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"

    "github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func registerHealthAgentTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
    slog.Info("Registering HAProxy health & agent check tools...")

    enableHealth := mcp.NewTool("enable_health",
        mcp.WithDescription("Enables health checks for a server in a backend"),
        mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
        mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable health checks for")),
    )
    s.AddTool(enableHealth, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        backend := getString(req, "backend")
        serverName := getString(req, "server")
        slog.InfoContext(ctx, "Executing enable_health", "backend", backend, "server", serverName)
        return callExec(ctx, "enable health checks", func() (string, error) {
            if err := client.EnableHealth(backend, serverName); err != nil {
                return "", err
            }
            return fmt.Sprintf("Health checks for server %s/%s enabled successfully", backend, serverName), nil
        })
    })

    disableHealth := mcp.NewTool("disable_health",
        mcp.WithDescription("Disables health checks for a server in a backend"),
        mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
        mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable health checks for")),
    )
    s.AddTool(disableHealth, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        backend := getString(req, "backend")
        serverName := getString(req, "server")
        slog.InfoContext(ctx, "Executing disable_health", "backend", backend, "server", serverName)
        return callExec(ctx, "disable health checks", func() (string, error) {
            if err := client.DisableHealth(backend, serverName); err != nil {
                return "", err
            }
            return fmt.Sprintf("Health checks for server %s/%s disabled successfully", backend, serverName), nil
        })
    })

    enableAgent := mcp.NewTool("enable_agent",
        mcp.WithDescription("Enables agent checks for a server in a backend"),
        mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
        mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable agent checks for")),
    )
    s.AddTool(enableAgent, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        backend := getString(req, "backend")
        serverName := getString(req, "server")
        slog.InfoContext(ctx, "Executing enable_agent", "backend", backend, "server", serverName)
        return callExec(ctx, "enable agent checks", func() (string, error) {
            if err := client.EnableAgent(backend, serverName); err != nil {
                return "", err
            }
            return fmt.Sprintf("Agent checks for server %s/%s enabled successfully", backend, serverName), nil
        })
    })

    disableAgent := mcp.NewTool("disable_agent",
        mcp.WithDescription("Disables agent checks for a server in a backend"),
        mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
        mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable agent checks for")),
    )
    s.AddTool(disableAgent, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        backend := getString(req, "backend")
        serverName := getString(req, "server")
        slog.InfoContext(ctx, "Executing disable_agent", "backend", backend, "server", serverName)
        return callExec(ctx, "disable agent checks", func() (string, error) {
            if err := client.DisableAgent(backend, serverName); err != nil {
                return "", err
            }
            return fmt.Sprintf("Agent checks for server %s/%s disabled successfully", backend, serverName), nil
        })
    })

    slog.Info("Health & agent check tools registered")
}