package haproxy

import (
	"fmt"
	"log/slog"
	"strings"
)

// ============================================
// Section: Statistics & Process Info
// ============================================

// GetRuntimeInfo retrieves runtime information (like 'show info').
func (c *HAProxyClient) GetRuntimeInfo() (map[string]string, error) {
	slog.Debug("HAProxyClient.GetRuntimeInfo called")

	// Execute the 'show info' command directly
	result, err := c.ExecuteRuntimeCommand("show info")
	if err != nil {
		slog.Error("Failed to get runtime info", "error", err)
		return nil, fmt.Errorf("failed to get runtime info: %w", err)
	}

	// Parse the result into a map
	infoMap := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Split each line by colon to get key-value pairs
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			infoMap[key] = value
		}
	}

	// Add default values if missing
	if _, ok := infoMap["version"]; !ok {
		infoMap["version"] = "unknown"
	}

	// Add HAProxy name (not directly available from API)
	infoMap["name"] = "HAProxy"

	slog.Debug("Successfully retrieved runtime info", "fields", len(infoMap))
	return infoMap, nil
}

// ShowStat executes the 'show stat' Runtime API command to get HAProxy statistics.
// The optional filter parameter can be used to filter by proxy or server names.
func (c *HAProxyClient) ShowStat(filter string) ([]map[string]string, error) {
	slog.Debug("HAProxyClient.ShowStat called", "filter", filter)

	// Construct command - add filter if provided
	cmd := "show stat"
	if filter != "" {
		cmd = fmt.Sprintf("%s %s", cmd, filter)
	}

	// Execute command
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to execute 'show stat' command", "error", err)
		return nil, fmt.Errorf("failed to execute 'show stat': %w", err)
	}

	    // Parse CSV-like output into structured data
	    _, stats, err := parseCSVStats(result)
	    if err != nil {
	        slog.Error("Failed to parse stats", "error", err)
	        return nil, fmt.Errorf("failed to parse stats: %w", err)
	    }
	    slog.Debug("Successfully retrieved stats", "count", len(stats))
	    return stats, nil
}

// GetStats retrieves runtime statistics.
func (c *HAProxyClient) GetStats() (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetStats called")

	// Use ShowStat which is already implemented with direct command
	rawStats, err := c.ShowStat("")
	if err != nil {
		slog.Error("Failed to get stats", "error", err)
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Build a structured result with the data
	result := make(map[string]interface{})

	// Group stats by backend/frontend and collect server info
	backendServers := make(map[string][]map[string]interface{})

	for _, stat := range rawStats {
		// Create a base stats map for this entry
		statMap := make(map[string]interface{})

		// Get type (frontend, backend, server)
		statType := stat["type"]
		statMap["type"] = statType

		// Add key statistics
		if status, ok := stat["status"]; ok {
			statMap["status"] = status
		}

		if scur, ok := stat["scur"]; ok {
			statMap["current_sessions"] = scur
		}

		if smax, ok := stat["smax"]; ok {
			statMap["max_sessions"] = smax
		}

		if slim, ok := stat["slim"]; ok {
			statMap["sessions_limit"] = slim
		}

		if bin, ok := stat["bin"]; ok {
			statMap["bytes_in"] = bin
		}

		if bout, ok := stat["bout"]; ok {
			statMap["bytes_out"] = bout
		}

		// Add more detailed stats
		if rate, ok := stat["rate"]; ok {
			statMap["rate"] = rate
		}

		if rateMax, ok := stat["rate_max"]; ok {
			statMap["rate_max"] = rateMax
		}

		if connTot, ok := stat["conn_tot"]; ok {
			statMap["connections_total"] = connTot
		}

		// Handle different types of stats
		name := stat["pxname"]
		switch statType {
		case "frontend":
			result[name] = statMap
		case "backend":
			// Store the backend stats
			result[name] = statMap
			// Initialize an empty server list for this backend
			backendServers[name] = []map[string]interface{}{}
		case "server":
			// This is a server entry, add it to the corresponding backend
			backendName := stat["pxname"]
			if backendName != "" {
				// Create server stats
				serverMap := make(map[string]interface{})
				serverMap["name"] = stat["svname"]

				// Copy relevant stats
				for k, v := range stat {
					serverMap[k] = v
				}

				// Add server to its backend
				if servers, ok := backendServers[backendName]; ok {
					backendServers[backendName] = append(servers, serverMap)
				}
			}
		}
	}

	// Add servers to their backends
	for backendName, servers := range backendServers {
		if backend, ok := result[backendName].(map[string]interface{}); ok {
			backend["servers"] = servers
		}
	}

	slog.Debug("Successfully processed stats", "groups", len(result))
	return result, nil
}

// DebugCounters retrieves HAProxy internal counters.
func (c *HAProxyClient) DebugCounters() (map[string]string, error) {
	slog.Debug("HAProxyClient.DebugCounters called")

	// Execute the command directly
	result, err := c.ExecuteRuntimeCommand("debug dev counters")
	if err != nil {
		slog.Error("Failed to get debug counters", "error", err)
		return nil, fmt.Errorf("failed to get debug counters: %w", err)
	}

	// Parse the output into a map
	counters := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Try to parse the line as "key=value" or "key: value"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Try with colon separator
			parts = strings.SplitN(line, ":", 2)
		}

		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			counters[key] = value
		} else {
			// For lines without a clear key-value structure, use the line itself as the key
			counters[line] = ""
		}
	}

	slog.Debug("Successfully retrieved debug counters", "count", len(counters))
	return counters, nil
}

// ClearCountersAll executes the 'clear counters all' runtime command to reset all HAProxy statistics.
func (c *HAProxyClient) ClearCountersAll() error {
	slog.Debug("HAProxyClient.ClearCountersAll called")

	// Execute the clear counters all command
	_, err := c.ExecuteRuntimeCommand("clear counters all")
	if err != nil {
		slog.Error("Failed to execute 'clear counters all' command", "error", err)
		return fmt.Errorf("failed to clear counters: %w", err)
	}

	slog.Debug("Successfully cleared all counters")
	return nil
}

// DumpStatsFile executes the 'dump stats-file' runtime command to write stats to a file.
func (c *HAProxyClient) DumpStatsFile(filepath string) (string, error) {
	slog.Debug("HAProxyClient.DumpStatsFile called", "filepath", filepath)

	// Construct command with the filepath
	cmd := fmt.Sprintf("dump stats-file %s", filepath)

	// Execute the command
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to execute 'dump stats-file' command", "error", err, "filepath", filepath)
		return "", fmt.Errorf("failed to dump stats to file: %w", err)
	}

	slog.Debug("Successfully dumped stats to file", "filepath", filepath, "result", result)
	return filepath, nil
}

// ============================================
// Section: Topology Discovery
// ============================================

// ShowServersState executes the 'show servers state' runtime command to get server states.
// If backend is empty, it will show all servers across all backends.
func (c *HAProxyClient) ShowServersState(backend string) ([]map[string]string, error) {
	slog.Debug("HAProxyClient.ShowServersState called", "backend", backend)

	// Construct command with optional backend filter
	cmd := "show servers state"
	if backend != "" {
		cmd = fmt.Sprintf("%s %s", cmd, backend)
	}

	// Execute the command
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to execute 'show servers state' command", "error", err)
		return nil, fmt.Errorf("failed to show servers state: %w", err)
	}

	// Parse the output into structured data
	// The format is typically: # be_id be_name srv_id srv_name srv_addr srv_op_state...
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 2 {
		// Need at least a header row and one data row
		return []map[string]string{}, nil
	}

	// Extract header columns (skipping first line if it's a comment)
	headerLineIndex := 0
	if strings.HasPrefix(lines[0], "#") {
		headerLineIndex = 1
	}

	// If we have no data rows, return empty result
	if headerLineIndex >= len(lines) {
		return []map[string]string{}, nil
	}

	// Split header into fields
	headerFields := strings.Fields(lines[headerLineIndex])

	// Parse data rows
	serverStates := make([]map[string]string, 0, len(lines)-headerLineIndex-1)
	for i := headerLineIndex + 1; i < len(lines); i++ {
		// Skip empty lines or comments
		if lines[i] == "" || strings.HasPrefix(lines[i], "#") {
			continue
		}

		row := make(map[string]string)
		fields := strings.Fields(lines[i])

		// Add each column to the row map
		for j := 0; j < len(headerFields) && j < len(fields); j++ {
			row[headerFields[j]] = fields[j]
		}

		serverStates = append(serverStates, row)
	}

	slog.Debug("Successfully retrieved server states", "count", len(serverStates))
	return serverStates, nil
}

// ============================================
// Section: Dynamic Server Management
// ============================================

// SetWeight sets a server's weight within a backend.
func (c *HAProxyClient) SetWeight(backend, server string, weight int) (string, error) {
	slog.Debug("HAProxyClient.SetWeight called", "backend", backend, "server", server, "weight", weight)

	// Construct command
	cmd := fmt.Sprintf("set weight %s/%s %d", backend, server, weight)

	// Execute the command
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to set server weight", "error", err, "backend", backend, "server", server)
		return "", fmt.Errorf("failed to set weight for server %s/%s: %w", backend, server, err)
	}

	// Parse the response, which is typically: "New weight <old>/<new>"
	result = strings.TrimSpace(result)
	slog.Debug("Successfully set server weight", "backend", backend, "server", server, "weight", weight, "result", result)

	return result, nil
}

// ============================================
// Section: Health Checks & Agents
// ============================================

// EnableHealth enables health checks for a server in a backend.
func (c *HAProxyClient) EnableHealth(backend, server string) error {
	slog.Info("Enabling health checks", "backend", backend, "server", server)

	// Construct the command
	cmd := fmt.Sprintf("enable health %s/%s", backend, server)

	// Execute the command
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to enable health checks", "error", err, "backend", backend, "server", server)
		return fmt.Errorf("failed to enable health checks for %s/%s: %w", backend, server, err)
	}

	slog.Info("Health checks enabled successfully", "backend", backend, "server", server)
	return nil
}

// DisableHealth disables health checks for a server in a backend.
func (c *HAProxyClient) DisableHealth(backend, server string) error {
	slog.Info("Disabling health checks", "backend", backend, "server", server)

	// Construct the command
	cmd := fmt.Sprintf("disable health %s/%s", backend, server)

	// Execute the command
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to disable health checks", "error", err, "backend", backend, "server", server)
		return fmt.Errorf("failed to disable health checks for %s/%s: %w", backend, server, err)
	}

	slog.Info("Health checks disabled successfully", "backend", backend, "server", server)
	return nil
}

// EnableAgent enables agent checks for a server in a backend.
func (c *HAProxyClient) EnableAgent(backend, server string) error {
	slog.Info("Enabling agent checks", "backend", backend, "server", server)

	// Construct the command
	cmd := fmt.Sprintf("enable agent %s/%s", backend, server)

	// Execute the command
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to enable agent checks", "error", err, "backend", backend, "server", server)
		return fmt.Errorf("failed to enable agent checks for %s/%s: %w", backend, server, err)
	}

	slog.Info("Agent checks enabled successfully", "backend", backend, "server", server)
	return nil
}

// DisableAgent disables agent checks for a server in a backend.
func (c *HAProxyClient) DisableAgent(backend, server string) error {
	slog.Info("Disabling agent checks", "backend", backend, "server", server)

	// Construct the command
	cmd := fmt.Sprintf("disable agent %s/%s", backend, server)

	// Execute the command
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to disable agent checks", "error", err, "backend", backend, "server", server)
		return fmt.Errorf("failed to disable agent checks for %s/%s: %w", backend, server, err)
	}

	slog.Info("Agent checks disabled successfully", "backend", backend, "server", server)
	return nil
}

// ============================================
// Section: Miscellaneous
// ============================================

// ReloadHAProxy triggers a configuration reload.
func (c *HAProxyClient) ReloadHAProxy() error {
	slog.Info("Reloading HAProxy configuration")

	// Execute the reload command directly
	result, err := c.ExecuteRuntimeCommand("reload")
	if err != nil {
		slog.Error("Failed to reload HAProxy", "error", err)
		return fmt.Errorf("failed to reload HAProxy: %w", err)
	}

	slog.Info("HAProxy reloaded successfully", "result", result)
	return nil
}
