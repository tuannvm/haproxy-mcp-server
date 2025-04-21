package haproxy

import (
	"fmt"
	"log/slog"
)

// GetBackends retrieves a list of all backends from HAProxy.
func (c *HAProxyClient) GetBackends() ([]string, error) {
	slog.Debug("HAProxyClient.GetBackends called")

	// Get the configuration client to retrieve backends
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// GetBackends takes a transaction ID (empty string for no transaction)
	// and returns: version string, backends array, error
	_, backends, err := configClient.GetBackends("")
	if err != nil {
		slog.Error("Failed to get backends", "error", err)
		return nil, fmt.Errorf("failed to list backends: %w", err)
	}

	// Process the backends data
	backendNames := make([]string, 0, len(backends))
	for _, backend := range backends {
		backendNames = append(backendNames, backend.Name)
	}

	slog.Debug("Successfully retrieved backends", "count", len(backendNames))
	return backendNames, nil
}

// GetBackendDetails retrieves detailed information about a specific backend.
func (c *HAProxyClient) GetBackendDetails(backendName string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetBackendDetails called", "backend", backendName)

	// First verify the backend exists
	configClient, err := c.Client.Configuration()
	if err != nil {
		slog.Error("Failed to get configuration client", "error", err)
		return nil, fmt.Errorf("failed to get configuration client: %w", err)
	}

	// Check if backend exists with GetBackend
	_, backend, err := configClient.GetBackend(backendName, "")
	if err != nil {
		slog.Error("Backend not found", "backend", backendName, "error", err)
		return nil, fmt.Errorf("backend '%s' not found: %w", backendName, err)
	}

	// Create backend details from configuration
	backendDetails := map[string]interface{}{
		"name": backend.Name,
	}

	// Add optional fields if available
	if backend.Mode != "" {
		backendDetails["mode"] = backend.Mode
	}

	if backend.Balance != nil && backend.Balance.Algorithm != nil && *backend.Balance.Algorithm != "" {
		backendDetails["balance"] = *backend.Balance.Algorithm
	}

	// Try to get additional stats from our GetStats method
	stats, err := c.GetStats()
	if err == nil {
		// Check if backend has stats in our implementation
		if backendStats, ok := stats[backendName].(map[string]interface{}); ok {
			// Merge stats into details
			for k, v := range backendStats {
				backendDetails[k] = v
			}
		}
	}

	slog.Debug("Successfully retrieved backend details", "backend", backendName)
	return backendDetails, nil
}

// SetMaxConnServer sets the maxconn value for a server within a backend.
func (c *HAProxyClient) SetMaxConnServer(backend, server string, maxconn int) error {
	slog.Debug("HAProxyClient.SetMaxConnServer called", "backend", backend, "server", server, "maxconn", maxconn)

	// Get the runtime client
	runtimeClient, err := c.Client.Runtime()
	if err != nil {
		slog.Error("Failed to get runtime client", "error", err)
		return fmt.Errorf("failed to get runtime client: %w", err)
	}

	// Construct command
	cmd := fmt.Sprintf("set maxconn server %s/%s %d", backend, server, maxconn)

	// Execute the command
	_, err = runtimeClient.ExecuteRaw(cmd)
	if err != nil {
		slog.Error("Failed to set server maxconn", "error", err, "backend", backend, "server", server)
		return fmt.Errorf("failed to set maxconn for server %s/%s: %w", backend, server, err)
	}

	slog.Debug("Successfully set server maxconn", "backend", backend, "server", server, "maxconn", maxconn)
	return nil
}
