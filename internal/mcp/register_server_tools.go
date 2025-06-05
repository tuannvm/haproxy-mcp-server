package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
)

func registerServerTools(s *server.MCPServer, client *haproxy.HAProxyClient) {
	slog.Info("Registering HAProxy server management tools...")

	// list_servers tool
	listServers := mcp.NewTool("list_servers",
		mcp.WithDescription("Lists servers within a specific HAProxy backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the servers")),
	)
	s.AddTool(listServers, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		slog.InfoContext(ctx, "Executing list_servers", "backend", backend)
		return callJSON(ctx, "list servers", "servers", func() (interface{}, error) {
			return client.ListServers(backend)
		})
	})

	// get_server tool
	getServer := mcp.NewTool("get_server",
		mcp.WithDescription("Gets details of a specific server within an HAProxy backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to retrieve")),
	)
	s.AddTool(getServer, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		serverName := getString(req, "server")
		slog.InfoContext(ctx, "Executing get_server", "backend", backend, "server", serverName)
		return callJSON(ctx, "get server details", "server", func() (interface{}, error) {
			return client.GetServerDetails(backend, serverName)
		})
	})

	// add_server tool
	addServer := mcp.NewTool("add_server",
		mcp.WithDescription("Adds a new server to a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend to add the server to")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name for the new server")),
		mcp.WithString("addr", mcp.Required(), mcp.Description("Address for the new server")),
		mcp.WithNumber("port", mcp.Description("Port for the new server")),
		mcp.WithNumber("weight", mcp.Description("Weight for the new server")),
	)
	s.AddTool(addServer, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		name := getString(req, "name")
		addr := getString(req, "addr")
		port := getInt(req, "port")
		weight := getInt(req, "weight")
		slog.InfoContext(ctx, "Executing add_server", "backend", backend, "name", name, "addr", addr, "port", port, "weight", weight)
		return callExec(ctx, "add server", func() (string, error) {
			if err := client.AddServer(backend, name, addr, port, weight); err != nil {
				return "", err
			}
			return fmt.Sprintf("Server %s added successfully to backend %s", name, backend), nil
		})
	})

	// del_server tool
	delServer := mcp.NewTool("del_server",
		mcp.WithDescription("Deletes a server from a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the server to delete")),
	)
	s.AddTool(delServer, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		name := getString(req, "name")
		slog.InfoContext(ctx, "Executing del_server", "backend", backend, "name", name)
		return callExec(ctx, "delete server", func() (string, error) {
			if err := client.DelServer(backend, name); err != nil {
				return "", err
			}
			return fmt.Sprintf("Server %s deleted successfully from backend %s", name, backend), nil
		})
	})

	// enable_server tool
	enableServer := mcp.NewTool("enable_server",
		mcp.WithDescription("Enables a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to enable")),
	)
	s.AddTool(enableServer, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		serverName := getString(req, "server")
		slog.InfoContext(ctx, "Executing enable_server", "backend", backend, "server", serverName)
		return callExec(ctx, "enable server", func() (string, error) {
			if err := client.EnableServer(backend, serverName); err != nil {
				return "", err
			}
			return fmt.Sprintf("Server %s/%s enabled successfully", backend, serverName), nil
		})
	})

	// disable_server tool
	disableServer := mcp.NewTool("disable_server",
		mcp.WithDescription("Disables a server in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to disable")),
	)
	s.AddTool(disableServer, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		serverName := getString(req, "server")
		slog.InfoContext(ctx, "Executing disable_server", "backend", backend, "server", serverName)
		return callExec(ctx, "disable server", func() (string, error) {
			if err := client.DisableServer(backend, serverName); err != nil {
				return "", err
			}
			return fmt.Sprintf("Server %s/%s disabled successfully", backend, serverName), nil
		})
	})

	// set_weight tool
	setWeight := mcp.NewTool("set_weight",
		mcp.WithDescription("Sets server weight in a backend"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to modify")),
		mcp.WithNumber("weight", mcp.Required(), mcp.Description("New weight value to set")),
	)
	s.AddTool(setWeight, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		serverName := getString(req, "server")
		weight := getInt(req, "weight")
		slog.InfoContext(ctx, "Executing set_weight", "backend", backend, "server", serverName, "weight", weight)
		return callExec(ctx, "set weight", func() (string, error) {
			return client.SetWeight(backend, serverName, weight)
		})
	})

	// set_maxconn_server tool
	setMaxconn := mcp.NewTool("set_maxconn_server",
		mcp.WithDescription("Sets maximum connections for a server"),
		mcp.WithString("backend", mcp.Required(), mcp.Description("Name of the backend containing the server")),
		mcp.WithString("server", mcp.Required(), mcp.Description("Name of the server to modify")),
		mcp.WithNumber("maxconn", mcp.Required(), mcp.Description("New maxconn value to set")),
	)
	s.AddTool(setMaxconn, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		backend := getString(req, "backend")
		serverName := getString(req, "server")
		maxconn := getInt(req, "maxconn")
		slog.InfoContext(ctx, "Executing set_maxconn_server", "backend", backend, "server", serverName, "maxconn", maxconn)
		return callExec(ctx, "set maxconn", func() (string, error) {
			if err := client.SetServerMaxconn(backend, serverName, maxconn); err != nil {
				return "", err
			}
			return fmt.Sprintf("Maxconn for server %s/%s set to %d", backend, serverName, maxconn), nil
		})
	})

	slog.Info("Server management tools registered")
}
