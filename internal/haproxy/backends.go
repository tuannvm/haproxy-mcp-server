package haproxy

import (
	"fmt"
	"log/slog"
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
	_, statsData, err := parseCSVStats(result)
	if err != nil {
		slog.Error("Failed to parse stats", "backend", backendName, "error", err)
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	// Process data lines
	backendInfo := &BackendInfo{
		Name:    backendName,
		Status:  "UNKNOWN", // Default status
		Servers: []ServerInfo{},
		Stats:   make(map[string]string),
	}

	foundBackend := false
	sessions := 0

	for _, fieldMap := range statsData {
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

			sessions = safeParseInt(fieldMap["scur"], 0)
			backendInfo.Sessions = sessions
		} else if svname != "FRONTEND" {
			// This is a server entry
			server := ServerInfo{
				Name: svname,
			}

			// Extract server address and port if available
			if addr, ok := fieldMap["addr"]; ok {
				server.Address, server.Port = parseAddressPort(addr)
			}

			// Extract other server info
			if status, ok := fieldMap["status"]; ok {
				server.Status = status
			}

			server.Weight = safeParseInt(fieldMap["weight"], 0)

			if checkStatus, ok := fieldMap["check_status"]; ok {
				server.CheckStatus = checkStatus
			}

			if lastChgStr, ok := fieldMap["lastchg"]; ok {
				server.LastStatusChange = lastChgStr
			}

			server.ActiveConnections = safeParseInt(fieldMap["scur"], 0)
			server.TotalConnections = safeParseInt(fieldMap["stot"], 0)

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
