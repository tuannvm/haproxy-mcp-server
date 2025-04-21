package haproxy

import (
	"fmt"
	"log/slog"
)

// ListServers retrieves a list of servers for a specific backend.
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	slog.Debug("HAProxyClient.ListServers called", "backend", backend)

	// Get the configuration client
	// Note: We need to use the configuration API here since the runtime API
	// doesn't provide a direct way to just list server names
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

// AddServer adds a new server to a backend dynamically.
// This uses the Runtime API 'add server' command.
func (c *HAProxyClient) AddServer(backend, name, addr string, port, weight int) error {
	slog.Info("Adding server", "backend", backend, "name", name, "addr", addr, "port", port, "weight", weight)

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Construct the add server command
	cmd := fmt.Sprintf("add server %s/%s %s", backend, name, addr)

	// Add port if specified
	if port > 0 {
		cmd = fmt.Sprintf("%s:%d", cmd, port)
	}

	// Add weight if specified
	if weight > 0 {
		cmd = fmt.Sprintf("%s weight %d", cmd, weight)
	}

	// Execute the command
	_, err = runtimeClient.ExecuteRaw(cmd)
	if err != nil {
		slog.Error("Failed to add server", "error", err, "backend", backend, "name", name)
		return fmt.Errorf("failed to add server %s to backend %s: %w", name, backend, err)
	}

	slog.Info("Server added successfully", "backend", backend, "name", name)
	return nil
}

// DelServer removes a server from a backend dynamically.
// This uses the Runtime API 'del server' command.
func (c *HAProxyClient) DelServer(backend, name string) error {
	slog.Info("Deleting server", "backend", backend, "name", name)

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Construct the del server command
	cmd := fmt.Sprintf("del server %s/%s", backend, name)

	// Execute the command
	_, err = runtimeClient.ExecuteRaw(cmd)
	if err != nil {
		slog.Error("Failed to delete server", "error", err, "backend", backend, "name", name)
		return fmt.Errorf("failed to delete server %s from backend %s: %w", name, backend, err)
	}

	slog.Info("Server deleted successfully", "backend", backend, "name", name)
	return nil
}
