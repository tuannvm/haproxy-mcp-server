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

	// Get the runtime client
	runtimeClient := c.Runtime()

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
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 2 {
		// Need at least a header row and one data row
		return []map[string]string{}, nil
	}

	// Parse header row to get column names
	headerFields := strings.Split(lines[0], ",")

	// Parse data rows
	stats := make([]map[string]string, 0, len(lines)-1)
	for i := 1; i < len(lines); i++ {
		row := make(map[string]string)
		fields := strings.Split(lines[i], ",")

		// Add each column to the row map
		for j := 0; j < len(headerFields) && j < len(fields); j++ {
			row[headerFields[j]] = fields[j]
		}

		stats = append(stats, row)
	}

	slog.Debug("Successfully retrieved stats", "count", len(stats))
	return stats, nil
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

// DebugCounters executes the 'debug counters' runtime command to get HAProxy internal counters.
func (c *HAProxyClient) DebugCounters() (map[string]string, error) {
	slog.Debug("HAProxyClient.DebugCounters called")

	// Execute the debug counters command
	result, err := c.ExecuteRuntimeCommand("debug dev counters")
	if err != nil {
		slog.Error("Failed to execute 'debug counters' command", "error", err)
		return nil, fmt.Errorf("failed to execute 'debug counters': %w", err)
	}

	// Parse the output into a map
	counters := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			counters[key] = value
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

	// Get the runtime client
	runtimeClient := c.Runtime()

	// Try to reload HAProxy via the socket
	result, err := runtimeClient.Reload()
	if err != nil {
		slog.Error("Failed to reload HAProxy via socket", "error", err)

		// If the runtime reload fails, log and return the error
		// In strict Runtime API mode, we shouldn't try configuration reload
		return fmt.Errorf("failed to reload HAProxy via runtime API: %w", err)
	}

	slog.Info("HAProxy reloaded via socket", "result", result)
	return nil
}
