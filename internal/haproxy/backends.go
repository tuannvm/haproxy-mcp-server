package haproxy

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// ListBackends returns a list of all HAProxy backends.
func (c *HAProxyClient) ListBackends() ([]string, error) {
	slog.Debug("Listing all HAProxy backends")

	// Use show stat command to get all stats and extract backend names
	result, err := c.ExecuteRuntimeCommand("show stat")
	if err != nil {
		slog.Error("Failed to get stats for backends", "error", err)
		return nil, fmt.Errorf("failed to get stats for backends: %w", err)
	}

	// Parse the result to extract backend names
	_, stats, err := parseCSVStats(result)
	if err != nil {
		slog.Error("Failed to parse stats", "error", err)
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	// Extract backend names
	backendSet := make(map[string]bool)
	for _, stat := range stats {
		// Only process backend entries
		if statType, ok := stat["type"]; ok && statType == "backend" {
			if name, ok := stat["pxname"]; ok && name != "" {
				backendSet[name] = true
			}
		}
	}

	// Convert to a slice
	backends := make([]string, 0, len(backendSet))
	for backend := range backendSet {
		backends = append(backends, backend)
	}

	slog.Debug("Successfully listed backends", "count", len(backends))
	return backends, nil
}

// GetBackendInfo returns detailed information about a specific backend.
func (c *HAProxyClient) GetBackendInfo(backendName string) (*BackendInfo, error) {
	slog.Debug("Getting backend info", "backend", backendName)

	// Use show stat to get stats for this backend
	cmd := fmt.Sprintf("show stat %s", backendName)
	result, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to get backend stats", "backend", backendName, "error", err)
		return nil, fmt.Errorf("failed to get backend stats: %w", err)
	}

	// Parse the result
	_, stats, err := parseCSVStats(result)
	if err != nil {
		slog.Error("Failed to parse stats", "error", err)
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	// Process data from stats
	backendInfo := &BackendInfo{
		Name:    backendName,
		Status:  "UNKNOWN", // Default status
		Servers: []ServerInfo{},
		Stats:   make(map[string]string),
	}

	foundBackend := false
	for _, stat := range stats {
		// Get the type and name
		statType, hasType := stat["type"]
		name, hasName := stat["pxname"]

		// Find the backend entry
		if hasType && hasName && statType == "backend" && name == backendName {
			foundBackend = true

			// Extract status if available
			if status, ok := stat["status"]; ok {
				backendInfo.Status = status
			}

			// Extract sessions
			if scur, ok := stat["scur"]; ok {
				if sessions, err := strconv.Atoi(scur); err == nil {
					backendInfo.Sessions = sessions
				}
			}

			// Copy all stats
			for k, v := range stat {
				backendInfo.Stats[k] = v
			}
		} else if hasType && hasName && statType == "server" && name == backendName {
			// This is a server in the backend we're looking for
			serverName, hasServerName := stat["svname"]
			if !hasServerName {
				continue
			}

			server := ServerInfo{
				Name: serverName,
			}

			// Extract server info
			if addr, ok := stat["addr"]; ok {
				parts := strings.Split(addr, ":")
				if len(parts) > 0 {
					server.Address = parts[0]
					if len(parts) > 1 {
						server.Port = parts[1]
					}
				} else {
					server.Address = addr
				}
			}

			if status, ok := stat["status"]; ok {
				server.Status = status
			}

			if weight, ok := stat["weight"]; ok {
				if w, err := strconv.Atoi(weight); err == nil {
					server.Weight = w
				}
			}

			if check, ok := stat["check_status"]; ok {
				server.CheckStatus = check
			}

			if lastchg, ok := stat["lastchg"]; ok {
				server.LastStatusChange = lastchg
			}

			if scur, ok := stat["scur"]; ok {
				if conn, err := strconv.Atoi(scur); err == nil {
					server.ActiveConnections = conn
				}
			}

			if stot, ok := stat["stot"]; ok {
				if conn, err := strconv.Atoi(stot); err == nil {
					server.TotalConnections = conn
				}
			}

			backendInfo.Servers = append(backendInfo.Servers, server)
		}
	}

	if !foundBackend {
		slog.Error("Backend not found", "backend", backendName)
		return nil, fmt.Errorf("backend not found: %s", backendName)
	}

	slog.Debug("Successfully retrieved backend info", "backend", backendName, "servers", len(backendInfo.Servers))
	return backendInfo, nil
}

// EnableBackend enables a backend.
func (c *HAProxyClient) EnableBackend(backendName string) error {
	slog.Debug("Enabling backend", "backend", backendName)

	// Set the backend state to ready using direct command
	cmd := fmt.Sprintf("set server %s/default-backend state ready", backendName)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to enable backend", "backend", backendName, "error", err)
		return fmt.Errorf("failed to enable backend %s: %w", backendName, err)
	}

	slog.Debug("Successfully enabled backend", "backend", backendName)
	return nil
}

// DisableBackend disables a backend.
func (c *HAProxyClient) DisableBackend(backendName string) error {
	slog.Debug("Disabling backend", "backend", backendName)

	// Set the backend state to maint using direct command
	cmd := fmt.Sprintf("set server %s/default-backend state maint", backendName)
	_, err := c.ExecuteRuntimeCommand(cmd)
	if err != nil {
		slog.Error("Failed to disable backend", "backend", backendName, "error", err)
		return fmt.Errorf("failed to disable backend %s: %w", backendName, err)
	}

	slog.Debug("Successfully disabled backend", "backend", backendName)
	return nil
}

// Methods to support tools.go

// GetBackends returns a list of all HAProxy backends
// This is an alias for ListBackends for API consistency
func (c *HAProxyClient) GetBackends() ([]string, error) {
	return c.ListBackends()
}

// GetBackendDetails retrieves details of a specific backend in a map format
// for tools.go compatibility
func (c *HAProxyClient) GetBackendDetails(backend string) (map[string]interface{}, error) {
	backendInfo, err := c.GetBackendInfo(backend)
	if err != nil {
		return nil, err
	}

	// Convert BackendInfo to map for tools.go
	result := map[string]interface{}{
		"name":     backendInfo.Name,
		"status":   backendInfo.Status,
		"sessions": backendInfo.Sessions,
		"stats":    backendInfo.Stats,
	}

	// Convert servers to map for consistent interface
	servers := make([]map[string]interface{}, 0, len(backendInfo.Servers))
	for _, server := range backendInfo.Servers {
		serverMap := map[string]interface{}{
			"name":               server.Name,
			"address":            server.Address,
			"port":               server.Port,
			"status":             server.Status,
			"weight":             server.Weight,
			"check_status":       server.CheckStatus,
			"last_status_change": server.LastStatusChange,
			"active_connections": server.ActiveConnections,
			"total_connections":  server.TotalConnections,
		}
		servers = append(servers, serverMap)
	}

	result["servers"] = servers
	return result, nil
}
