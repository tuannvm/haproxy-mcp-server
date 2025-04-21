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

	// Get backends by executing a simple command
	result, err := c.client.ExecuteRaw("show backends")
	if err != nil {
		slog.Error("Failed to list backends", "error", err)
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}

	// Split the result by newlines to get backend names
	backends := splitAndTrim(result)
	slog.Debug("Successfully listed backends", "count", len(backends))
	return backends, nil
}

// BackendInfo contains information about a HAProxy backend.
type BackendInfo struct {
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Servers  []ServerInfo      `json:"servers,omitempty"`
	Sessions int               `json:"sessions"`
	Stats    map[string]string `json:"stats,omitempty"`
}

// ServerInfo contains information about a HAProxy server.
type ServerInfo struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	Port              int    `json:"port,omitempty"`
	Status            string `json:"status,omitempty"`
	Weight            int    `json:"weight,omitempty"`
	CheckStatus       string `json:"check_status,omitempty"`
	LastStatusChange  string `json:"last_status_change,omitempty"`
	TotalConnections  int    `json:"total_connections,omitempty"`
	ActiveConnections int    `json:"active_connections,omitempty"`
}

// GetBackendInfo returns detailed information about a specific backend.
func (c *HAProxyClient) GetBackendInfo(backendName string) (*BackendInfo, error) {
	slog.Debug("Getting backend info", "backend", backendName)

	// Get stats for this backend
	result, err := c.ExecuteRuntimeCommand("show stat")
	if err != nil {
		slog.Error("Failed to get backend stats", "backend", backendName, "error", err)
		return nil, fmt.Errorf("failed to get backend stats: %w", err)
	}

	// Parse the CSV-like output
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 2 {
		slog.Error("Invalid stats output format", "backend", backendName)
		return nil, fmt.Errorf("invalid stats output format")
	}

	// Get headers from first line
	headers := strings.Split(lines[0], ",")

	// Process data lines
	backendInfo := &BackendInfo{
		Name:    backendName,
		Status:  "UNKNOWN", // Default status
		Servers: []ServerInfo{},
		Stats:   make(map[string]string),
	}

	foundBackend := false
	sessions := 0

	for i := 1; i < len(lines); i++ {
		data := strings.Split(lines[i], ",")
		if len(data) < len(headers) {
			continue // Skip incomplete lines
		}

		// Create a map of field name to value
		fieldMap := make(map[string]string)
		for j := 0; j < len(headers) && j < len(data); j++ {
			fieldMap[headers[j]] = data[j]
		}

		// Check if this is our backend or a server in our backend
		pxname, hasPxname := fieldMap["pxname"]
		svname, hasSvname := fieldMap["svname"]
		if !hasPxname || !hasSvname || pxname != backendName {
			continue
		}

		// Handle backend entry
		if svname == "BACKEND" {
			foundBackend = true

			// Save all stats
			for k, v := range fieldMap {
				if v != "" {
					backendInfo.Stats[k] = v
				}
			}

			// Extract specific fields
			if status, ok := fieldMap["status"]; ok {
				backendInfo.Status = status
			}

			if sessionStr, ok := fieldMap["scur"]; ok {
				if sess, err := strconv.Atoi(sessionStr); err == nil {
					sessions = sess
				}
			}

			backendInfo.Sessions = sessions
		} else if svname != "FRONTEND" {
			// This is a server entry
			server := ServerInfo{
				Name: svname,
			}

			// Extract server address and port if available
			if addr, ok := fieldMap["addr"]; ok {
				server.Address = addr

				// Some versions might include port in the address
				if strings.Contains(addr, ":") {
					parts := strings.Split(addr, ":")
					server.Address = parts[0]
					if len(parts) > 1 {
						if port, err := strconv.Atoi(parts[1]); err == nil {
							server.Port = port
						}
					}
				}
			}

			// Extract other server info
			if status, ok := fieldMap["status"]; ok {
				server.Status = status
			}

			if weightStr, ok := fieldMap["weight"]; ok {
				if weight, err := strconv.Atoi(weightStr); err == nil {
					server.Weight = weight
				}
			}

			if checkStatus, ok := fieldMap["check_status"]; ok {
				server.CheckStatus = checkStatus
			}

			if lastChgStr, ok := fieldMap["lastchg"]; ok {
				server.LastStatusChange = lastChgStr
			}

			if connsStr, ok := fieldMap["scur"]; ok {
				if conns, err := strconv.Atoi(connsStr); err == nil {
					server.ActiveConnections = conns
				}
			}

			if totConnsStr, ok := fieldMap["stot"]; ok {
				if conns, err := strconv.Atoi(totConnsStr); err == nil {
					server.TotalConnections = conns
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

	// Enable the backend by setting its state to ready
	command := fmt.Sprintf("backend %s", backendName)
	result, err := c.client.ExecuteRaw(command)
	if err != nil {
		slog.Error("Failed to enable backend", "backend", backendName, "error", err)
		return fmt.Errorf("failed to enable backend %s: %w", backendName, err)
	}

	// Check if the result indicates success
	if result != "" {
		slog.Debug("Command output", "output", result)
	}

	// Execute additional command to set state
	stateCmd := fmt.Sprintf("set server %s/default-backend state ready", backendName)
	_, err = c.client.ExecuteRaw(stateCmd)
	if err != nil {
		slog.Error("Failed to set backend state", "backend", backendName, "error", err)
		return fmt.Errorf("failed to set backend %s state: %w", backendName, err)
	}

	slog.Debug("Successfully enabled backend", "backend", backendName)
	return nil
}

// DisableBackend disables a backend.
func (c *HAProxyClient) DisableBackend(backendName string) error {
	slog.Debug("Disabling backend", "backend", backendName)

	// Disable the backend by setting its state to maint
	command := fmt.Sprintf("backend %s", backendName)
	result, err := c.client.ExecuteRaw(command)
	if err != nil {
		slog.Error("Failed to disable backend", "backend", backendName, "error", err)
		return fmt.Errorf("failed to disable backend %s: %w", backendName, err)
	}

	// Check if the result indicates success
	if result != "" {
		slog.Debug("Command output", "output", result)
	}

	// Execute additional command to set state
	stateCmd := fmt.Sprintf("set server %s/default-backend state maint", backendName)
	_, err = c.client.ExecuteRaw(stateCmd)
	if err != nil {
		slog.Error("Failed to set backend state", "backend", backendName, "error", err)
		return fmt.Errorf("failed to set backend %s state: %w", backendName, err)
	}

	slog.Debug("Successfully disabled backend", "backend", backendName)
	return nil
}

// splitAndTrim splits a string by newlines and trims each line, returning only non-empty lines
func splitAndTrim(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
