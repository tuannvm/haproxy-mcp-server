package haproxy

import (
	"fmt"
	"log/slog"
	"strconv"
)

// ListBackends returns a list of all HAProxy backends.
func (c *HAProxyClient) ListBackends() ([]string, error) {
	slog.Debug("Listing all HAProxy backends")

	// Unfortunately, the client-native library doesn't have a direct method to list all backends
	// We'll use GetStats method to get all stats and extract backend names
	nativeStats := c.client.GetStats()
	if nativeStats.Error != "" {
		slog.Error("Failed to get stats for backends", "error", nativeStats.Error)
		return nil, fmt.Errorf("failed to get stats for backends: %s", nativeStats.Error)
	}

	// Extract backend names from stats
	backendSet := make(map[string]bool)
	for _, stat := range nativeStats.Stats {
		// Only process backend entries
		if stat.Type == "backend" {
			backendSet[stat.Name] = true
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

	// Get native stats directly
	nativeStats := c.client.GetStats()
	if nativeStats.Error != "" {
		slog.Error("Failed to get backend stats", "backend", backendName, "error", nativeStats.Error)
		return nil, fmt.Errorf("failed to get backend stats: %s", nativeStats.Error)
	}

	// Process data from native stats
	backendInfo := &BackendInfo{
		Name:    backendName,
		Status:  "UNKNOWN", // Default status
		Servers: []ServerInfo{},
		Stats:   make(map[string]string),
	}

	foundBackend := false
	for _, stat := range nativeStats.Stats {
		// Find the backend entry
		if stat.Type == "backend" && stat.Name == backendName {
			foundBackend = true

			// Save stat fields to our Stats map
			if stat.Stats != nil {
				// Extract status if available
				if stat.Stats.Status != "" {
					backendInfo.Status = stat.Stats.Status
				}

				// Extract sessions
				if stat.Stats.Scur != nil {
					backendInfo.Sessions = int(*stat.Stats.Scur)
				}

				// Extract other stats through reflection (limited support)
				statsFields := map[string]interface{}{
					"qcur":     stat.Stats.Qcur,
					"qmax":     stat.Stats.Qmax,
					"scur":     stat.Stats.Scur,
					"smax":     stat.Stats.Smax,
					"slim":     stat.Stats.Slim,
					"stot":     stat.Stats.Stot,
					"bin":      stat.Stats.Bin,
					"bout":     stat.Stats.Bout,
					"dreq":     stat.Stats.Dreq,
					"dresp":    stat.Stats.Dresp,
					"ereq":     stat.Stats.Ereq,
					"econ":     stat.Stats.Econ,
					"eresp":    stat.Stats.Eresp,
					"wretr":    stat.Stats.Wretr,
					"wredis":   stat.Stats.Wredis,
					"status":   stat.Stats.Status,
					"weight":   stat.Stats.Weight,
					"act":      stat.Stats.Act,
					"bck":      stat.Stats.Bck,
					"chkfail":  stat.Stats.Chkfail,
					"chkdown":  stat.Stats.Chkdown,
					"lastchg":  stat.Stats.Lastchg,
					"downtime": stat.Stats.Downtime,
					"pid":      stat.Stats.Pid,
					"iid":      stat.Stats.Iid,
					"sid":      stat.Stats.Sid,
					"throttle": stat.Stats.Throttle,
					"lbtot":    stat.Stats.Lbtot,
					"rate":     stat.Stats.Rate,
					"rate_lim": stat.Stats.RateLim,
					"rate_max": stat.Stats.RateMax,
				}

				for name, value := range statsFields {
					if value != nil {
						switch v := value.(type) {
						case *int64:
							if v != nil {
								backendInfo.Stats[name] = fmt.Sprintf("%d", *v)
							}
						case *int:
							if v != nil {
								backendInfo.Stats[name] = fmt.Sprintf("%d", *v)
							}
						case string:
							if v != "" {
								backendInfo.Stats[name] = v
							}
						}
					}
				}
			}
		} else if stat.Type == "server" && stat.BackendName == backendName {
			// This is a server in the backend we're looking for
			server := ServerInfo{
				Name: stat.Name,
			}

			if stat.Stats != nil {
				// Extract common server stats
				if stat.Stats.Addr != "" {
					// Use the utility function that returns the correct types
					address, port := parseAddressPort(stat.Stats.Addr)
					server.Address = address
					server.Port = strconv.Itoa(port)
				}

				if stat.Stats.Status != "" {
					server.Status = stat.Stats.Status
				}

				if stat.Stats.Weight != nil {
					server.Weight = int(*stat.Stats.Weight)
				}

				if stat.Stats.CheckStatus != "" {
					server.CheckStatus = stat.Stats.CheckStatus
				}

				if stat.Stats.Lastchg != nil {
					server.LastStatusChange = fmt.Sprintf("%d", *stat.Stats.Lastchg)
				}

				if stat.Stats.Scur != nil {
					server.ActiveConnections = int(*stat.Stats.Scur)
				}

				if stat.Stats.Stot != nil {
					server.TotalConnections = int(*stat.Stats.Stot)
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

	// Set the default-backend server state to ready
	// Native client doesn't have a direct method to enable backends,
	// but we can set the state of the default-backend server
	err := c.client.SetServerState(backendName, "default-backend", ServerStateReady)
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

	// Set the default-backend server state to maint
	// Native client doesn't have a direct method to disable backends,
	// but we can set the state of the default-backend server
	err := c.client.SetServerState(backendName, "default-backend", ServerStateMaint)
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
			"total_connections":  server.TotalConnections,
			"active_connections": server.ActiveConnections,
		}
		servers = append(servers, serverMap)
	}
	result["servers"] = servers

	return result, nil
}
