package haproxy

import (
	"fmt"
	"log/slog"
)

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
