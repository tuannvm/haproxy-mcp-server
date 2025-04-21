package haproxy

import (
	"fmt"
	"log/slog"
	"strings"
)

// ListServers retrieves a list of servers for a specific backend.
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	slog.Debug("HAProxyClient.ListServers called", "backend", backend)

	// Use the native client's method to get server states
	serverStates, err := c.client.GetServersState(backend)
	if err != nil {
		slog.Error("Failed to list servers", "backend", backend, "error", err)
		return nil, fmt.Errorf("failed to list servers for backend %s: %w", backend, err)
	}

	// Extract server names
	serverNames := make([]string, 0, len(serverStates))
	for _, server := range serverStates {
		serverNames = append(serverNames, server.Name)
	}

	slog.Debug("Successfully retrieved servers", "backend", backend, "count", len(serverNames))
	return serverNames, nil
}

// GetServerDetails retrieves detailed information about a specific server.
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetServerDetails called", "backend", backend, "server", server)

	// Get server state from native client
	serverState, err := c.client.GetServerState(backend, server)
	if err != nil {
		slog.Error("Failed to get server state", "backend", backend, "server", server, "error", err)
		return nil, fmt.Errorf("failed to get server state for %s/%s: %w", backend, server, err)
	}

	// Build server details
	details := map[string]interface{}{
		"name":    server,
		"backend": backend,
		"address": serverState.Address,
		"status":  serverState.OperationalState,
	}

	// Add port if available
	if serverState.Port != nil {
		details["port"] = *serverState.Port
	}

	// Add admin state if available
	if serverState.AdminState != "" {
		details["admin_state"] = serverState.AdminState
	}

	// Get additional stats from stats command if available
	statsCmd := fmt.Sprintf("show stat %s %s", backend, server)
	statsOutput, err := c.ExecuteRuntimeCommand(statsCmd)
	if err == nil && len(statsOutput) > 0 {
		// Parse stats output for additional details
		lines := strings.Split(strings.TrimSpace(statsOutput), "\n")
		if len(lines) >= 2 {
			// Get headers from first line
			headers := strings.Split(lines[0], ",")
			// Get data from second line
			data := strings.Split(lines[1], ",")

			// Map data to headers
			for i := 0; i < len(headers) && i < len(data); i++ {
				if data[i] != "" {
					details[headers[i]] = data[i]
				}
			}
		}
	}

	slog.Debug("Successfully retrieved server details", "backend", backend, "server", server)
	return details, nil
}

// EnableServer enables a server in a backend.
func (c *HAProxyClient) EnableServer(backend, server string) error {
	slog.Debug("Enabling server", "backend", backend, "server", server)

	// Use native client's method
	err := c.client.EnableServer(backend, server)
	if err != nil {
		slog.Error("Failed to enable server", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to enable server %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully enabled server", "backend", backend, "server", server)
	return nil
}

// DisableServer disables a server in a backend.
func (c *HAProxyClient) DisableServer(backend, server string) error {
	slog.Debug("Disabling server", "backend", backend, "server", server)

	// Use native client's method
	err := c.client.DisableServer(backend, server)
	if err != nil {
		slog.Error("Failed to disable server", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to disable server %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully disabled server", "backend", backend, "server", server)
	return nil
}

// SetServerWeight sets the weight of a server in a backend.
func (c *HAProxyClient) SetServerWeight(backend, server string, weight int) error {
	slog.Debug("Setting server weight", "backend", backend, "server", server, "weight", weight)

	// Validate weight
	if weight < 0 || weight > 256 {
		return fmt.Errorf("invalid weight %d (must be between 0 and 256)", weight)
	}

	// Use native client's method
	err := c.client.SetServerWeight(backend, server, fmt.Sprintf("%d", weight))
	if err != nil {
		slog.Error("Failed to set server weight", "backend", backend, "server", server, "weight", weight, "error", err)
		return fmt.Errorf("failed to set weight for server %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully set server weight", "backend", backend, "server", server, "weight", weight)
	return nil
}

// SetServerMaxconn sets the maximum connections for a server.
func (c *HAProxyClient) SetServerMaxconn(backend, server string, maxconn int) error {
	slog.Debug("Setting server maxconn", "backend", backend, "server", server, "maxconn", maxconn)

	// Validate maxconn
	if maxconn < 0 {
		return fmt.Errorf("invalid maxconn %d (must be >= 0)", maxconn)
	}

	// Execute the set maxconn command
	cmd := fmt.Sprintf("set maxconn server %s/%s %d", backend, server, maxconn)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to set server maxconn", "backend", backend, "server", server, "maxconn", maxconn, "error", err)
		return fmt.Errorf("failed to set maxconn for server %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully set server maxconn", "backend", backend, "server", server, "maxconn", maxconn)
	return nil
}

// GetServerState retrieves the state of a server in a backend.
func (c *HAProxyClient) GetServerState(backend, server string) (string, error) {
	slog.Debug("Getting server state", "backend", backend, "server", server)

	// Use native client's method
	serverState, err := c.client.GetServerState(backend, server)
	if err != nil {
		slog.Error("Failed to get server state", "backend", backend, "server", server, "error", err)
		return "", fmt.Errorf("failed to get server state for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully got server state",
		"backend", backend,
		"server", server,
		"state", serverState.OperationalState)

	return serverState.OperationalState, nil
}

// GetServersState retrieves the state of all servers in a backend.
func (c *HAProxyClient) GetServersState(backend string) ([]map[string]string, error) {
	slog.Debug("Getting servers state", "backend", backend)

	// Use native client's method
	serverStates, err := c.client.GetServersState(backend)
	if err != nil {
		slog.Error("Failed to get servers state", "backend", backend, "error", err)
		return nil, fmt.Errorf("failed to get servers state for backend %s: %w", backend, err)
	}

	// Convert to a more generic format
	result := make([]map[string]string, 0, len(serverStates))
	for _, state := range serverStates {
		serverMap := map[string]string{
			"name":    state.Name,
			"address": state.Address,
			"state":   state.OperationalState,
		}

		// Add additional fields if available
		if state.AdminState != "" {
			serverMap["admin_state"] = state.AdminState
		}

		// Get weight using stats command for this specific server
		statsCmd := fmt.Sprintf("show stat %s %s", backend, state.Name)
		statsOutput, err := c.ExecuteRuntimeCommand(statsCmd)
		if err == nil && len(statsOutput) > 0 {
			// Parse stats output for weight
			lines := strings.Split(strings.TrimSpace(statsOutput), "\n")
			if len(lines) >= 2 {
				// Get headers from first line
				headers := strings.Split(lines[0], ",")
				// Get data from second line
				data := strings.Split(lines[1], ",")

				// Look for weight field
				for i := 0; i < len(headers) && i < len(data); i++ {
					if headers[i] == "weight" && data[i] != "" {
						serverMap["weight"] = data[i]
						break
					}
				}
			}
		}

		result = append(result, serverMap)
	}

	slog.Debug("Successfully got servers state", "backend", backend, "count", len(result))
	return result, nil
}

// EnableAgentCheck enables the agent check for a server.
func (c *HAProxyClient) EnableAgentCheck(backend, server string) error {
	slog.Debug("Enabling agent check", "backend", backend, "server", server)

	// Use native client's method
	err := c.client.EnableAgentCheck(backend, server)
	if err != nil {
		slog.Error("Failed to enable agent check", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to enable agent check for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully enabled agent check", "backend", backend, "server", server)
	return nil
}

// DisableAgentCheck disables the agent check for a server.
func (c *HAProxyClient) DisableAgentCheck(backend, server string) error {
	slog.Debug("Disabling agent check", "backend", backend, "server", server)

	// Use native client's method
	err := c.client.DisableAgentCheck(backend, server)
	if err != nil {
		slog.Error("Failed to disable agent check", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to disable agent check for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully disabled agent check", "backend", backend, "server", server)
	return nil
}

// EnableHealthCheck enables the health check for a server.
func (c *HAProxyClient) EnableHealthCheck(backend, server string) error {
	slog.Debug("Enabling health check", "backend", backend, "server", server)

	// Use native client's method
	err := c.client.EnableServerHealth(backend, server)
	if err != nil {
		slog.Error("Failed to enable health check", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to enable health check for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully enabled health check", "backend", backend, "server", server)
	return nil
}

// DisableHealthCheck disables the health check for a server.
func (c *HAProxyClient) DisableHealthCheck(backend, server string) error {
	slog.Debug("Disabling health check", "backend", backend, "server", server)

	// Execute the disable health command - native client doesn't have this method
	cmd := fmt.Sprintf("disable health %s/%s", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to disable health check", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to disable health check for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully disabled health check", "backend", backend, "server", server)
	return nil
}

// AddServer adds a new server to a backend dynamically.
func (c *HAProxyClient) AddServer(backend, name, addr string, port, weight int) error {
	slog.Info("Adding server", "backend", backend, "name", name, "addr", addr, "port", port, "weight", weight)

	// Build the attributes string
	attributes := addr
	if port > 0 {
		attributes = fmt.Sprintf("%s:%d", attributes, port)
	}
	if weight > 0 {
		attributes = fmt.Sprintf("%s weight %d", attributes, weight)
	}

	// Use native client's method
	err := c.client.AddServer(backend, name, attributes)
	if err != nil {
		slog.Error("Failed to add server", "error", err, "backend", backend, "name", name)
		return fmt.Errorf("failed to add server %s to backend %s: %w", name, backend, err)
	}

	slog.Info("Server added successfully", "backend", backend, "name", name)
	return nil
}

// DelServer removes a server from a backend dynamically.
func (c *HAProxyClient) DelServer(backend, name string) error {
	slog.Info("Deleting server", "backend", backend, "name", name)

	// Use native client's method
	err := c.client.DeleteServer(backend, name)
	if err != nil {
		slog.Error("Failed to delete server", "error", err, "backend", backend, "name", name)
		return fmt.Errorf("failed to delete server %s from backend %s: %w", name, backend, err)
	}

	slog.Info("Server deleted successfully", "backend", backend, "name", name)
	return nil
}
