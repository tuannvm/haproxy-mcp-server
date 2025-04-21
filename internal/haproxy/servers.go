package haproxy

import (
	"fmt"
	"log/slog"
	"strings"
)

// ListServers retrieves a list of servers for a specific backend.
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	slog.Debug("HAProxyClient.ListServers called", "backend", backend)

	// Use direct command to get server state
	cmd := fmt.Sprintf("show servers state %s", backend)
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to list servers", "backend", backend, "error", err)
		return nil, fmt.Errorf("failed to list servers for backend %s: %w", backend, err)
	}

	// Parse the output to extract server names
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 2 {
		// Return empty list if not enough lines (need header + data)
		return []string{}, nil
	}

	// Find the server name column index (assuming second line contains column headers)
	headerLine := 1
	if strings.HasPrefix(lines[0], "#") {
		headerLine = 1
	}

	if headerLine >= len(lines) {
		return []string{}, nil
	}

	headers := strings.Fields(lines[headerLine])
	nameIndex := -1
	for i, header := range headers {
		if header == "srv_name" {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		slog.Error("Failed to find server name column", "backend", backend)
		return nil, fmt.Errorf("failed to find server name column for backend %s", backend)
	}

	// Extract server names
	serverNames := make([]string, 0)
	for i := headerLine + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(lines[i], "#") {
			continue
		}

		fields := strings.Fields(lines[i])
		if nameIndex < len(fields) {
			serverNames = append(serverNames, fields[nameIndex])
		}
	}

	slog.Debug("Successfully retrieved servers", "backend", backend, "count", len(serverNames))
	return serverNames, nil
}

// GetServerDetails retrieves detailed information about a specific server.
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetServerDetails called", "backend", backend, "server", server)

	// Get server state from direct command
	stateCmd := fmt.Sprintf("show servers state %s %s", backend, server)
	stateOutput, err := c.ExecuteRuntimeCommand(stateCmd)
	if err != nil {
		slog.Error("Failed to get server state", "backend", backend, "server", server, "error", err)
		return nil, fmt.Errorf("failed to get server state for %s/%s: %w", backend, server, err)
	}

	// Build basic server details
	details := map[string]interface{}{
		"name":    server,
		"backend": backend,
	}

	// Parse server state output
	lines := strings.Split(strings.TrimSpace(stateOutput), "\n")
	if len(lines) >= 3 { // Need at least comment, header, and data line
		// Find the header line
		headerLine := 1
		if strings.HasPrefix(lines[0], "#") {
			headerLine = 1
		}

		// Get headers and data
		headers := strings.Fields(lines[headerLine])
		dataLine := headerLine + 1

		if dataLine < len(lines) {
			data := strings.Fields(lines[dataLine])

			// Map headers to data
			for i := 0; i < len(headers) && i < len(data); i++ {
				details[headers[i]] = data[i]

				// Special handling for common fields
				switch headers[i] {
				case "srv_addr":
					details["address"] = data[i]
				case "srv_op_state":
					details["status"] = data[i]
				}
			}
		}
	}

	// Get additional stats from stats command if available
	statsCmd := fmt.Sprintf("show stat %s %s", backend, server)
	statsOutput, err := c.ExecuteRuntimeCommand(statsCmd)
	if err == nil && len(statsOutput) > 0 {
		// Parse stats output for additional details
		_, statsData, parseErr := parseCSVStats(statsOutput)
		if parseErr == nil && len(statsData) > 0 {
			// Add all fields from the first row of stats
			for key, value := range statsData[0] {
				if value != "" {
					details[key] = value
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

	// Use direct command
	cmd := fmt.Sprintf("set server %s/%s state ready", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("set server %s/%s state maint", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("set server %s/%s weight %d", backend, server, weight)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("show servers state %s %s", backend, server)
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to get server state", "backend", backend, "server", server, "error", err)
		return "", fmt.Errorf("failed to get server state for %s/%s: %w", backend, server, err)
	}

	// Parse the output to find operational state
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 3 { // Need at least comment, header, and data line
		return "", fmt.Errorf("insufficient data in server state output for %s/%s", backend, server)
	}

	// Find the header line
	headerLine := 1
	if strings.HasPrefix(lines[0], "#") {
		headerLine = 1
	}

	// Get headers and data
	headers := strings.Fields(lines[headerLine])
	dataLine := headerLine + 1

	if dataLine >= len(lines) {
		return "", fmt.Errorf("missing data line in server state output for %s/%s", backend, server)
	}

	// Find the srv_op_state column
	stateIdx := -1
	for i, h := range headers {
		if h == "srv_op_state" {
			stateIdx = i
			break
		}
	}

	if stateIdx == -1 {
		return "", fmt.Errorf("srv_op_state column not found in server state output for %s/%s", backend, server)
	}

	// Extract state value
	data := strings.Fields(lines[dataLine])
	if stateIdx >= len(data) {
		return "", fmt.Errorf("srv_op_state value not found in server state output for %s/%s", backend, server)
	}

	state := data[stateIdx]
	slog.Debug("Successfully got server state", "backend", backend, "server", server, "state", state)

	return state, nil
}

// GetServersState retrieves the state of all servers in a backend.
func (c *HAProxyClient) GetServersState(backend string) ([]map[string]string, error) {
	slog.Debug("Getting servers state", "backend", backend)

	// Use direct command
	cmd := fmt.Sprintf("show servers state %s", backend)
	output, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to get servers state", "backend", backend, "error", err)
		return nil, fmt.Errorf("failed to get servers state for backend %s: %w", backend, err)
	}

	// Parse the output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 { // Need at least header and data line
		return []map[string]string{}, nil
	}

	// Find the header line
	headerLine := 1
	if strings.HasPrefix(lines[0], "#") {
		headerLine = 1
	}

	if headerLine >= len(lines) {
		return []map[string]string{}, nil
	}

	// Get headers
	headers := strings.Fields(lines[headerLine])

	// Convert to a more generic format
	servers := make([]map[string]string, 0)
	for i := headerLine + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(lines[i], "#") {
			continue
		}

		fields := strings.Fields(lines[i])
		serverMap := make(map[string]string)

		// Map fields to headers
		for j := 0; j < len(headers) && j < len(fields); j++ {
			serverMap[headers[j]] = fields[j]
		}

		// Add standard fields if present
		for _, key := range headers {
			if key == "be_name" || key == "srv_name" || key == "srv_addr" || key == "srv_op_state" {
				idx := -1
				for i, h := range headers {
					if h == key {
						idx = i
						break
					}
				}

				if idx >= 0 && idx < len(fields) {
					// Map to standard names
					switch key {
					case "be_name":
						serverMap["backend"] = fields[idx]
					case "srv_name":
						serverMap["name"] = fields[idx]
					case "srv_addr":
						serverMap["address"] = fields[idx]
					case "srv_op_state":
						serverMap["state"] = fields[idx]
					}
				}
			}
		}

		// Add to result if it has name and backend
		if _, hasName := serverMap["name"]; hasName {
			servers = append(servers, serverMap)
		}
	}

	slog.Debug("Successfully got servers state", "backend", backend, "count", len(servers))
	return servers, nil
}

// EnableAgentCheck enables the agent check for a server.
func (c *HAProxyClient) EnableAgentCheck(backend, server string) error {
	slog.Debug("Enabling agent check", "backend", backend, "server", server)

	// Use direct command
	cmd := fmt.Sprintf("enable agent %s/%s", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("disable agent %s/%s", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("enable health %s/%s", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
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

	// Use direct command
	cmd := fmt.Sprintf("disable health %s/%s", backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to disable health check", "backend", backend, "server", server, "error", err)
		return fmt.Errorf("failed to disable health check for %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully disabled health check", "backend", backend, "server", server)
	return nil
}

// AddServer adds a new server to a backend.
func (c *HAProxyClient) AddServer(backend, name, addr string, port, weight int) error {
	slog.Debug("Adding server", "backend", backend, "server", name, "address", addr, "port", port, "weight", weight)

	// Form the add server command
	cmd := fmt.Sprintf("add server %s/%s %s:%d weight %d", backend, name, addr, port, weight)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to add server", "backend", backend, "server", name, "error", err)
		return fmt.Errorf("failed to add server %s/%s: %w", backend, name, err)
	}

	slog.Debug("Successfully added server", "backend", backend, "server", name)
	return nil
}

// DelServer removes a server from a backend.
func (c *HAProxyClient) DelServer(backend, name string) error {
	slog.Debug("Deleting server", "backend", backend, "server", name)

	// Form the delete server command
	cmd := fmt.Sprintf("del server %s/%s", backend, name)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to delete server", "backend", backend, "server", name, "error", err)
		return fmt.Errorf("failed to delete server %s/%s: %w", backend, name, err)
	}

	slog.Debug("Successfully deleted server", "backend", backend, "server", name)
	return nil
}
