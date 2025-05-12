package mcp

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func registerBackendTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
	slog.Info("Registering HAProxy backend management tools...")

	// list_backends tool
	listBackends := mcp.NewTool("list_backends",
		mcp.WithDescription("Lists all configured HAProxy backends"),
	)
	s.AddTool(listBackends, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing list_backends")
		return callJSON(ctx, "list backends", "backends", func() (interface{}, error) {
			return client.GetBackends()
		})
	})

	// get_backend tool
	getBackend := mcp.NewTool("get_backend",
		mcp.WithDescription("Gets details of a specific HAProxy backend"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the backend to retrieve")),
	)
	s.AddTool(getBackend, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := getString(req, "name")
		slog.InfoContext(ctx, "Executing get_backend", "name", name)
		return callJSON(ctx, "get backend details", "backend", func() (interface{}, error) {
			return client.GetBackendDetails(name)
		})
	})

	// show_servers_state tool
	showServersState := mcp.NewTool("show_servers_state",
		mcp.WithDescription("Shows the state of servers including sessions and weight"),
		mcp.WithString("backend", mcp.Description("Optional backend name to filter servers")),
	)
	s.AddTool(showServersState, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		slog.InfoContext(ctx, "Executing show_servers_state", "backend", backend)
		return callJSON(ctx, "show servers state", "servers_state", func() (interface{}, error) {
			return client.ShowServersState(backend)
		})
	})

	slog.Info("Backend management tools registered")
}
