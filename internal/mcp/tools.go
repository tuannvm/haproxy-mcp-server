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

	// ============================================
	// Section: Statistics & Process Info
	// ============================================

	// --- show_stat tool ---
	showStatTool := mcp.NewTool("show_stat",
		mcp.WithDescription("Shows HAProxy statistics table (show stat command)"),
		mcp.WithString("filter", mcp.Description("Optional filter for proxy or server names")),
	)

	s.AddTool(showStatTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filter, _ := req.Params.Arguments["filter"].(string)
		slog.InfoContext(ctx, "Executing show_stat tool", "filter", filter)

		stats, err := client.ShowStat(filter)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get statistics", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get statistics: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"stats": stats})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal show_stat output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- show_info tool ---
	showInfoTool := mcp.NewTool("show_info",
		mcp.WithDescription("Shows HAProxy runtime information (version, uptime, limits, mode)"),
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

	// --- debug_counters tool ---
	debugCountersTool := mcp.NewTool("debug_counters",
		mcp.WithDescription("Shows HAProxy internal counters (allocations, events)"),
	)

	s.AddTool(debugCountersTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing debug_counters tool")

		counters, err := client.DebugCounters()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get debug counters", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get debug counters: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"counters": counters})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal debug_counters output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- clear_counters_all tool ---
	clearCountersAllTool := mcp.NewTool("clear_counters_all",
		mcp.WithDescription("Reset all HAProxy statistics counters"),
	)

	s.AddTool(clearCountersAllTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing clear_counters_all tool")

		err := client.ClearCountersAll()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to clear counters", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to clear counters: %v", err)), nil
		}

		return mcp.NewToolResultText("All statistics counters have been reset successfully"), nil
	})

	// --- dump_stats_file tool ---
	dumpStatsFileTool := mcp.NewTool("dump_stats_file",
		mcp.WithDescription("Dump HAProxy stats to a file"),
		mcp.WithString("filepath", mcp.Required(), mcp.Description("Path where stats file should be saved")),
	)

	s.AddTool(dumpStatsFileTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filepath, _ := req.Params.Arguments["filepath"].(string)
		slog.InfoContext(ctx, "Executing dump_stats_file tool", "filepath", filepath)

		result, err := client.DumpStatsFile(filepath)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to dump stats to file", "error", err, "filepath", filepath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to dump stats to file: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Statistics dumped successfully to %s", result)), nil
	})

	// ============================================
	// Section: Backend Management
	// ============================================

	// --- list_backends tool ---
	listBackendsTool := mcp.NewTool("list_backends",
		mcp.WithDescription("Lists all configured HAProxy backends"),
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
		mcp.WithDescription("Gets details of a specific HAProxy backend"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the backend to retrieve")),
	)

	s.AddTool(getBackendTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := req.Params.Arguments["name"].(string)
		slog.InfoContext(ctx, "Executing get_backend tool", "backend_name", name)

		backendDetails, err := client.GetBackendDetails(name)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get backend details", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get backend details: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"backend": backendDetails})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal get_backend output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- show_servers_state tool ---
	showServersStateTool := mcp.NewTool("show_servers_state",
		mcp.WithDescription("Shows the state of servers including sessions and weight"),
		mcp.WithString("backend", mcp.Description("Optional backend name to filter servers")),
	)

	s.AddTool(showServersStateTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		slog.InfoContext(ctx, "Executing show_servers_state tool", "backend", backend)

		serversState, err := client.ShowServersState(backend)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to show servers state", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to show servers state: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"servers_state": serversState})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal show_servers_state output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// ============================================
	// Section: Server Management
	// ============================================

	// --- list_servers tool ---
	listServersTool := mcp.NewTool("list_servers",
		mcp.WithDescription("Lists servers within a specific HAProxy backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the servers")),
	)

	s.AddTool(listServersTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		slog.InfoContext(ctx, "Executing list_servers tool", "backend", backend)

		servers, err := client.ListServers(backend)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to list servers", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list servers: %v", err)), nil
		}

		// Marshal result to JSON
		output := map[string]interface{}{
			"backend": backend,
			"servers": servers,
		}
		jsonData, err := json.Marshal(output)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal list_servers output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- get_server tool ---
	getServerTool := mcp.NewTool("get_server",
		mcp.WithDescription("Gets details of a specific server within an HAProxy backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to retrieve")),
	)

	s.AddTool(getServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing get_server tool", "backend", backend, "server", server)

		serverDetails, err := client.GetServerDetails(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get server details", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get server details: %v", err)), nil
		}

		// Marshal result to JSON
		jsonData, err := json.Marshal(map[string]interface{}{"server": serverDetails})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal get_server output", "error", err)
			return mcp.NewToolResultError("Internal server error: failed to marshal results"), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// --- add_server tool ---
	addServerTool := mcp.NewTool("add_server",
		mcp.WithDescription("Adds a new server to a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend to add the server to")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name for the new server")),
		mcp.WithString("addr", mcp.Required(), mcp.Description("Address for the new server")),
		mcp.WithNumber("port", mcp.Description("Port for the new server")),
		mcp.WithNumber("weight", mcp.Description("Weight for the new server")),
	)

	s.AddTool(addServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		name, _ := req.Params.Arguments["name"].(string)
		addr, _ := req.Params.Arguments["addr"].(string)
		portVal, portExists := req.Params.Arguments["port"]
		weightVal, weightExists := req.Params.Arguments["weight"]

		var port, weight int
		if portExists {
			port = int(portVal.(float64))
		}
		if weightExists {
			weight = int(weightVal.(float64))
		}

		slog.InfoContext(ctx, "Executing add_server tool",
			"backend", backend,
			"name", name,
			"addr", addr,
			"port", port,
			"weight", weight)

		err := client.AddServer(backend, name, addr, port, weight)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to add server", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add server: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Server %s added successfully to backend %s", name, backend)), nil
	})

	// --- del_server tool ---
	delServerTool := mcp.NewTool("del_server",
		mcp.WithDescription("Deletes a server from a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the server to delete")),
	)

	s.AddTool(delServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		name, _ := req.Params.Arguments["name"].(string)
		slog.InfoContext(ctx, "Executing del_server tool", "backend", backend, "name", name)

		err := client.DelServer(backend, name)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to delete server", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete server: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Server %s deleted successfully from backend %s", name, backend)), nil
	})

	// --- enable_server tool ---
	enableServerTool := mcp.NewTool("enable_server",
		mcp.WithDescription("Enables a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable")),
	)

	s.AddTool(enableServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing enable_server tool", "backend", backend, "server", server)

		err := client.EnableServer(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to enable server", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to enable server: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Server %s/%s enabled successfully", backend, server)), nil
	})

	// --- disable_server tool ---
	disableServerTool := mcp.NewTool("disable_server",
		mcp.WithDescription("Disables a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable")),
	)

	s.AddTool(disableServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing disable_server tool", "backend", backend, "server", server)

		err := client.DisableServer(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to disable server", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to disable server: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Server %s/%s disabled successfully", backend, server)), nil
	})

	// --- set_weight tool ---
	setWeightTool := mcp.NewTool("set_weight",
		mcp.WithDescription("Sets server weight in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to modify")),
		mcp.WithNumber("weight", mcp.Required(), mcp.Description("New weight value to set")),
	)

	s.AddTool(setWeightTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		weightVal, _ := req.Params.Arguments["weight"].(float64)
		weight := int(weightVal)

		slog.InfoContext(ctx, "Executing set_weight tool", "backend", backend, "server", server, "weight", weight)

		result, err := client.SetWeight(backend, server, weight)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to set weight", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set weight: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	})

	// --- set_maxconn_server tool ---
	setMaxconnServerTool := mcp.NewTool("set_maxconn_server",
		mcp.WithDescription("Sets maximum connections for a server"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to modify")),
		mcp.WithNumber("maxconn", mcp.Required(), mcp.Description("New maxconn value to set")),
	)

	s.AddTool(setMaxconnServerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		maxconnVal, _ := req.Params.Arguments["maxconn"].(float64)
		maxconn := int(maxconnVal)

		slog.InfoContext(ctx, "Executing set_maxconn_server tool", "backend", backend, "server", server, "maxconn", maxconn)

		err := client.SetMaxConnServer(backend, server, maxconn)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to set maxconn", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set maxconn: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Maxconn for server %s/%s set to %d", backend, server, maxconn)), nil
	})

	// ============================================
	// Section: Health Checks & Agents
	// ============================================

	// --- enable_health tool ---
	enableHealthTool := mcp.NewTool("enable_health",
		mcp.WithDescription("Enables health checks for a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable health checks for")),
	)

	s.AddTool(enableHealthTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing enable_health tool", "backend", backend, "server", server)

		err := client.EnableHealth(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to enable health checks", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to enable health checks: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Health checks for server %s/%s enabled successfully", backend, server)), nil
	})

	// --- disable_health tool ---
	disableHealthTool := mcp.NewTool("disable_health",
		mcp.WithDescription("Disables health checks for a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable health checks for")),
	)

	s.AddTool(disableHealthTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing disable_health tool", "backend", backend, "server", server)

		err := client.DisableHealth(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to disable health checks", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to disable health checks: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Health checks for server %s/%s disabled successfully", backend, server)), nil
	})

	// --- enable_agent tool ---
	enableAgentTool := mcp.NewTool("enable_agent",
		mcp.WithDescription("Enables agent checks for a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable agent checks for")),
	)

	s.AddTool(enableAgentTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing enable_agent tool", "backend", backend, "server", server)

		err := client.EnableAgent(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to enable agent checks", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to enable agent checks: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Agent checks for server %s/%s enabled successfully", backend, server)), nil
	})

	// --- disable_agent tool ---
	disableAgentTool := mcp.NewTool("disable_agent",
		mcp.WithDescription("Disables agent checks for a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable agent checks for")),
	)

	s.AddTool(disableAgentTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend, _ := req.Params.Arguments["backend"].(string)
		server, _ := req.Params.Arguments["server"].(string)
		slog.InfoContext(ctx, "Executing disable_agent tool", "backend", backend, "server", server)

		err := client.DisableAgent(backend, server)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to disable agent checks", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to disable agent checks: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Agent checks for server %s/%s disabled successfully", backend, server)), nil
	})

	// --- reload_haproxy tool ---
	reloadHAProxyTool := mcp.NewTool("reload_haproxy",
		mcp.WithDescription("Triggers a reload of the HAProxy configuration"),
	)

	s.AddTool(reloadHAProxyTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.InfoContext(ctx, "Executing reload_haproxy tool")

		err := client.ReloadHAProxy()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to reload HAProxy configuration", "error", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reload HAProxy: %v", err)), nil
		}

		return mcp.NewToolResultText("HAProxy configuration reloaded successfully"), nil
	})

	slog.Info("All HAProxy MCP tools registered successfully")
}
