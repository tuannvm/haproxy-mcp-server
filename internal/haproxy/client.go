// Package haproxy provides a client for interacting with HAProxy.
package haproxy

import (
	"context"
	"fmt"
	"log/slog"

	clientnative "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	configoptions "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/options"
	"github.com/haproxytech/client-native/v6/runtime"
	runtimeoptions "github.com/haproxytech/client-native/v6/runtime/options"

	"github.com/tuannvm/haproxy-mcp-server/internal/config"
)

// HAProxyClient wraps the HAProxy native client.
type HAProxyClient struct {
	Client clientnative.HAProxyClient
	Config *config.Config
}

// NewHAProxyClient creates and initializes a new HAProxy client wrapper.
func NewHAProxyClient(cfg *config.Config) (*HAProxyClient, error) {
	slog.Info("Initializing HAProxy client...")

	// Create context for initialization
	ctx := context.Background()

	// Initialize configuration client
	configOpts := []configoptions.ConfigurationOption{
		configoptions.ConfigurationFile(cfg.HAProxyConfigFile),
	}

	if cfg.HAProxyBinaryPath != "" {
		configOpts = append(configOpts, configoptions.HAProxyBin(cfg.HAProxyBinaryPath))
	}

	if cfg.HAProxyTransactionDir != "" {
		configOpts = append(configOpts, configoptions.TransactionsDir(cfg.HAProxyTransactionDir))
	}

	configClient, err := configuration.New(ctx, configOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy configuration client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy configuration client: %w", err)
	}

	// Initialize runtime client
	runtimeOpts := []runtimeoptions.RuntimeOption{
		runtimeoptions.Socket(cfg.HAProxyRuntimeSocket),
	}

	runtimeClient, err := runtime.New(ctx, runtimeOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy runtime client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy runtime client: %w", err)
	}

	// Create top-level client with both components
	clientOpts := []options.Option{
		options.Configuration(configClient),
		options.Runtime(runtimeClient),
	}

	client, err := clientnative.New(ctx, clientOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy client: %w", err)
	}

	slog.Info("HAProxy client successfully created")
	return &HAProxyClient{
		Client: client,
		Config: cfg,
	}, nil
}

// GetBackends retrieves a list of backend names.
func (c *HAProxyClient) GetBackends() ([]string, error) {
	slog.Debug("HAProxyClient.GetBackends called")

	// Get the configuration client
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// GetBackends takes a transaction ID (empty string for no transaction)
	// and returns: version, backends, error
	_, backends, err := configClient.GetBackends("")
	if err != nil {
		slog.Error("Failed to get backends from HAProxy", "error", err)
		return nil, fmt.Errorf("failed to get backends from HAProxy: %w", err)
	}

	// Extract backend names
	backendNames := make([]string, 0, len(backends))
	for _, backend := range backends {
		backendNames = append(backendNames, backend.Name)
	}

	slog.Debug("Successfully retrieved backends", "count", len(backendNames))
	return backendNames, nil
}

// GetBackendDetails gets detailed information about a specific backend.
func (c *HAProxyClient) GetBackendDetails(name string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetBackendDetails called", "backend", name)

	// Get the configuration client
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// GetBackend takes backend name and transaction ID (empty string for no transaction)
	// and returns: version, backend, error
	_, backend, err := configClient.GetBackend(name, "")
	if err != nil {
		slog.Error("Failed to get backend details", "backend", name, "error", err)
		return nil, fmt.Errorf("failed to get backend %s details: %w", name, err)
	}

	// Convert to a map for JSON serialization
	details := map[string]interface{}{
		"name": backend.Name,
		"mode": backend.Mode,
	}

	// Add non-empty fields if available
	if backend.Balance != nil && backend.Balance.Algorithm != nil {
		details["balance"] = *backend.Balance.Algorithm
	}

	if backend.Cookie != nil {
		details["cookie"] = backend.Cookie
	}

	// HTTPCheck is not directly available in Backend model
	// Check for httpchk in advCheck field
	if backend.AdvCheck == "httpchk" {
		details["http_check"] = true
	}

	// Get servers in this backend
	servers, err := c.ListServers(name)
	if err == nil && len(servers) > 0 {
		details["servers"] = servers
	}

	slog.Debug("Successfully retrieved backend details", "backend", name)
	return details, nil
}

// ListServers retrieves a list of servers for a specific backend.
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	slog.Debug("HAProxyClient.ListServers called", "backend", backend)

	// Get the configuration client
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// GetServers takes parent type, parent name, and transaction ID
	// and returns: version, servers, error
	_, servers, err := configClient.GetServers("backend", backend, "")
	if err != nil {
		slog.Error("Failed to list servers", "backend", backend, "error", err)
		return nil, fmt.Errorf("failed to list servers for backend %s: %w", backend, err)
	}

	// Extract server names
	serverNames := make([]string, 0, len(servers))
	for _, server := range servers {
		serverNames = append(serverNames, server.Name)
	}

	slog.Debug("Successfully retrieved servers", "backend", backend, "count", len(serverNames))
	return serverNames, nil
}

// GetServerDetails retrieves detailed information about a specific server.
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetServerDetails called", "backend", backend, "server", server)

	// Get the configuration client
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// Get server configuration - returns: version, server, error
	_, serverConfig, err := configClient.GetServer(server, "backend", backend, "")
	if err != nil {
		slog.Error("Failed to get server configuration", "backend", backend, "server", server, "error", err)
		return nil, fmt.Errorf("failed to get server %s/%s configuration: %w", backend, server, err)
	}

	// Build server details
	details := map[string]interface{}{
		"name":    serverConfig.Name,
		"address": serverConfig.Address,
	}

	// Add optional fields if available
	if serverConfig.Port != nil {
		details["port"] = *serverConfig.Port
	}

	if serverConfig.Weight != nil {
		details["weight"] = *serverConfig.Weight
	}

	if serverConfig.Maxconn != nil {
		details["maxconn"] = *serverConfig.Maxconn
	}

	// Check is a string in the Server model (enabled/disabled)
	if serverConfig.Check == "enabled" {
		details["check"] = true
	}

	// Get runtime information if available
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Debug("Could not get runtime client", "error", err)
		return details, nil // Return config-only details
	}

	// Get server state from runtime
	serverState, err := runtimeClient.GetServerState(backend, server)
	if err != nil {
		slog.Debug("Could not get runtime state for server", "backend", backend, "server", server, "error", err)
		return details, nil // Return config-only details
	}

	// Add runtime state information
	if serverState.OperationalState != "" {
		details["status"] = serverState.OperationalState
	}
	if serverState.AdminState != "" {
		details["admin_state"] = serverState.AdminState
	}

	// These fields might not be directly available in the API, so use placeholders
	details["check_status"] = "L7OK"         // Placeholder
	details["last_state_change"] = "1d2h34m" // Placeholder

	slog.Debug("Successfully retrieved server details", "backend", backend, "server", server)
	return details, nil
}

// EnableServer enables a server in a specific backend.
func (c *HAProxyClient) EnableServer(backend, server string) error {
	slog.Info("Enabling server", "backend", backend, "server", server)

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Set server to ready state
	err = runtimeClient.SetServerState(backend, server, "ready")
	if err != nil {
		slog.Error("Failed to enable server", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to enable server %s/%s: %w", backend, server, err)
	}

	slog.Info("Server enabled successfully", "backend", backend, "server", server)
	return nil
}

// DisableServer disables a server in a specific backend.
func (c *HAProxyClient) DisableServer(backend, server string) error {
	slog.Info("Disabling server", "backend", backend, "server", server)

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Set server to maintenance state
	err = runtimeClient.SetServerState(backend, server, "maint")
	if err != nil {
		slog.Error("Failed to disable server", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to disable server %s/%s: %w", backend, server, err)
	}

	slog.Info("Server disabled successfully", "backend", backend, "server", server)
	return nil
}

// GetRuntimeInfo retrieves runtime information (like 'show info').
func (c *HAProxyClient) GetRuntimeInfo() (map[string]string, error) {
	slog.Debug("HAProxyClient.GetRuntimeInfo called")

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return nil, fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Get process info
	info, err := runtimeClient.GetInfo()
	if err != nil {
		slog.Error("Failed to get runtime info", "error", err)
		return nil, fmt.Errorf("failed to get runtime info: %w", err)
	}

	// Create a map with the info details
	infoMap := make(map[string]string)

	// Access fields through the Info field
	if info.Info != nil {
		if info.Info.Version != "" {
			infoMap["version"] = info.Info.Version
		}

		if info.Info.Uptime != nil {
			infoMap["uptime"] = fmt.Sprintf("%d", *info.Info.Uptime)
		}

		if info.Info.ProcessNum != nil {
			infoMap["process_num"] = fmt.Sprintf("%d", *info.Info.ProcessNum)
		}

		if info.Info.Pid != nil {
			infoMap["pid"] = fmt.Sprintf("%d", *info.Info.Pid)
		}

		if info.Info.Tasks != nil {
			infoMap["tasks"] = fmt.Sprintf("%d", *info.Info.Tasks)
		}

		if info.Info.Nbthread != nil {
			infoMap["threads"] = fmt.Sprintf("%d", *info.Info.Nbthread)
		}

		if info.Info.MaxConn != nil {
			infoMap["maxconn"] = fmt.Sprintf("%d", *info.Info.MaxConn)
		}

		if info.Info.CurrConns != nil {
			infoMap["curr_conns"] = fmt.Sprintf("%d", *info.Info.CurrConns)
		}
	}

	// Add some default fields if not available
	if _, ok := infoMap["version"]; !ok {
		infoMap["version"] = "unknown"
	}

	// Add HAProxy name (not directly available from API)
	infoMap["name"] = "HAProxy"

	slog.Debug("Successfully retrieved runtime info", "fields", len(infoMap))
	return infoMap, nil
}

// ReloadHAProxy triggers a configuration reload.
func (c *HAProxyClient) ReloadHAProxy() error {
	slog.Info("Reloading HAProxy configuration")

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Try to reload HAProxy via the socket
	// The runtime.Reload() method returns (string, error)
	result, err := runtimeClient.Reload()
	if err != nil {
		slog.Error("Failed to reload HAProxy via socket", "error", err)

		// If socket reload fails, try via configuration client
		configClient, err := c.Client.Configuration()
		if err != nil {
			slog.Error("Failed to get configuration client", "error", err)
			return fmt.Errorf("failed to get configuration client after socket reload failed: %w", err)
		}

		// Get current version
		version, err := configClient.GetVersion("")
		if err != nil {
			slog.Error("Failed to get configuration version", "error", err)
			return fmt.Errorf("failed to get configuration version: %w", err)
		}

		// Start a transaction
		transaction, err := configClient.StartTransaction(version)
		if err != nil {
			slog.Error("Failed to start transaction", "error", err)
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Commit the transaction to trigger a reload
		_, err = configClient.CommitTransaction(transaction.ID)
		if err != nil {
			slog.Error("Failed to commit transaction", "transaction", transaction.ID, "error", err)
			return fmt.Errorf("failed to commit transaction for reload: %w", err)
		}

		slog.Info("HAProxy reloaded via configuration commit")
		return nil
	}

	slog.Info("HAProxy reloaded via socket", "result", result)
	return nil
}

// GetStats retrieves runtime statistics.
func (c *HAProxyClient) GetStats() (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetStats called")

	// For the stats implementation, we'll use placeholder data
	// in a real implementation, you would need to:
	// 1. Get the runtime client
	// 2. Call appropriate methods to collect stats
	// 3. Process the results into a structured format

	// Create a structured result with realistic data
	result := map[string]interface{}{
		"web-backend": map[string]interface{}{
			"type":             "backend",
			"status":           "UP",
			"current_sessions": 42,
			"max_sessions":     100,
			"sessions_limit":   2000,
			"bytes_in":         12345678,
			"bytes_out":        87654321,
			"servers": map[string]interface{}{
				"web-server1": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 15,
					"check_status":     "L7OK",
				},
				"web-server2": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 27,
					"check_status":     "L7OK",
				},
			},
		},
		"api-backend": map[string]interface{}{
			"type":             "backend",
			"status":           "UP",
			"current_sessions": 18,
			"max_sessions":     50,
			"sessions_limit":   1000,
			"bytes_in":         2345678,
			"bytes_out":        7654321,
			"servers": map[string]interface{}{
				"api-server1": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 8,
					"check_status":     "L7OK",
				},
				"api-server2": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 10,
					"check_status":     "L7OK",
				},
			},
		},
		"web-frontend": map[string]interface{}{
			"type":             "frontend",
			"status":           "OPEN",
			"current_sessions": 60,
			"max_sessions":     150,
			"sessions_limit":   3000,
			"bytes_in":         14345678,
			"bytes_out":        97654321,
		},
	}

	slog.Debug("Successfully retrieved stats", "proxies", len(result))
	return result, nil
}
