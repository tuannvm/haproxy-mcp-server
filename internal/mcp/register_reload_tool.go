package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func registerReloadTool(s *server.MCPServer, client *haproxy.HAProxyClient) {
	slog.Info("Registering HAProxy reload tool...")

	reloadTool := mcp.NewTool("reload_haproxy",
		mcp.WithDescription("Triggers a reload of the HAProxy configuration"),
	)
	s.AddTool(reloadTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing reload_haproxy")
		return callExec(ctx, "reload haproxy", func() (string, error) {
			if err := client.ReloadHAProxy(); err != nil {
				return "", err
			}
			return "HAProxy configuration reloaded successfully", nil
		})
	})

	slog.Info("Reload tool registered")
}
