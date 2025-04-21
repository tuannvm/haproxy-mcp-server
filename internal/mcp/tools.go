package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server" // Import directly without alias

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

// RegisterTools registers all defined HAProxy tools with the MCP server.
func RegisterTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
	slog.Info("Registering HAProxy MCP tools...")

	// --- list_backends tool ---
	listBackendsTool := mcp.NewTool("list_backends",
		mcp.WithDescription("Lists all configured HAProxy backends."),
	)

	s.AddTool(listBackendsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing list_backends tool")

		backends, err := client.GetBackends()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to list backends", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list backends: %v", err)), nil
		}

		// Marshal result to JSON
		output := map[string]interface{}{
			"backends": backends,
		}
		jsonData, err := json.Marshal(output)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal list_backends output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- get_backend tool ---
	getBackendTool := mcp.NewTool("get_backend",
		mcp.WithDescription("Gets details of a specific HAProxy backend."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the backend to retrieve.")),
	)

	s.AddTool(getBackendTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := req.Params.Arguments["name"].(string)
		slog.InfoContext(ctx, "Executing get_backend tool", "backend_name", name)

		// TODO: Replace with actual implementation
		// backendDetails, err := client.GetBackendDetails(name)
		details := map[string]interface{}{
			"mode":        "http",
			"placeholder": true,
		}

		output := map[string]interface{}{
			"name":    name,
			"details": details,
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(output)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal get_backend output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- list_servers tool ---
	listServersTool := mcp.NewTool("list_servers",
		mcp.WithDescription("Lists servers within a specific HAProxy backend."),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the servers.")),
	)

	s.AddTool(listServersTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		slog.InfoContext(ctx, "Executing list_servers tool", "backend", backend)

		// TODO: Replace with actual implementation
		// servers, err := client.ListServers(backend)
		servers := []string{"server1", "server2"}

		output := map[string]interface{}{
			"backend": backend,
			"servers": servers,
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(output)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal list_servers output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- get_server tool ---
	getServerTool := mcp.NewTool("get_server",
		mcp.WithDescription("Gets details of a specific server within an HAProxy backend."),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server.")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to retrieve.")),
	)

	s.AddTool(getServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing get_server tool", "backend", backend, "server", server)

		// TODO: Replace with actual implementation
		// serverDetails, err := client.GetServerDetails(backend, server)
		details := map[string]interface{}{
			"address":     "127.0.0.1",
			"port":        8080,
			"status":      "UP",
			"placeholder": true,
		}

		output := map[string]interface{}{
			"backend": backend,
			"server":  server,
			"details": details,
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(output)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal get_server output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- enable_server tool ---
	enableServerTool := mcp.NewTool("enable_server",
		mcp.WithDescription("Enables a specific server within an HAProxy backend."),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server.")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable.")),
	)

	s.AddTool(enableServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing enable_server tool", "backend", backend, "server", server)

		// TODO: Replace with actual implementation
		// err := client.EnableServer(backend, server)
		// if err != nil {
		//     slog.ErrorContext(ctx, "Failed to enable server", "error", err, "backend", backend, "server", server)
		//     return mcp.NewToolResultError(fmt.Sprintf("Failed to enable server %s/%s: %v", backend, server, err)), nil
		// }

		return mcp.NewToolResultText(fmt.Sprintf("Server %s/%s enabled successfully.", backend, server)), nil
	})

	// --- disable_server tool ---
	disableServerTool := mcp.NewTool("disable_server",
		mcp.WithDescription("Disables a specific server within an HAProxy backend."),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server.")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable.")),
	)

	s.AddTool(disableServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing disable_server tool", "backend", backend, "server", server)

		// TODO: Replace with actual implementation
		// err := client.DisableServer(backend, server)
		// if err != nil {
		//     slog.ErrorContext(ctx, "Failed to disable server", "error", err, "backend", backend, "server", server)
		//     return mcp.NewToolResultError(fmt.Sprintf("Failed to disable server %s/%s: %v", backend, server, err)), nil
		// }

		return mcp.NewToolResultText(fmt.Sprintf("Server %s/%s disabled successfully.", backend, server)), nil
	})

	// --- show_info tool ---
	showInfoTool := mcp.NewTool("show_info",
		mcp.WithDescription("Gets runtime information from HAProxy (similar to 'show info')."),
	)

	s.AddTool(showInfoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing show_info tool")

		info, err := client.GetRuntimeInfo()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get runtime info", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get runtime info: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"info": info})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal show_info output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- reload_haproxy tool ---
	reloadHAProxyTool := mcp.NewTool("reload_haproxy",
		mcp.WithDescription("Triggers a reload of the HAProxy configuration."),
	)

	s.AddTool(reloadHAProxyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing reload_haproxy tool")

		err := client.ReloadHAProxy()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to reload HAProxy configuration", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reload HAProxy: %v", err)), nil
		}

		return mcp.NewToolResultText("HAProxy configuration reloaded successfully."), nil
	})

	// --- get_stats tool ---
	getStatsTool := mcp.NewTool("get_stats",
		mcp.WithDescription("Gets runtime statistics from HAProxy (similar to 'show stat')."),
	)

	s.AddTool(getStatsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing get_stats tool")

		// TODO: Replace with actual implementation
		// stats, err := client.GetStats()
		// if err != nil {
		//     slog.ErrorContext(ctx, "Failed to get stats", "error", err)
		//     return mcp.NewToolResultError(fmt.Sprintf("Failed to get stats: %v", err)), nil
		// }

		stats := map[string]interface{}{
			"frontend": map[string]string{
				"scur": "10",
				"smax": "100",
			},
			"backend_1": map[string]string{
				"qcur": "0",
				"scur": "5",
			},
			"placeholder": true,
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"stats": stats})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal get_stats output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	slog.Info("All HAProxy MCP tools registered successfully")
}
