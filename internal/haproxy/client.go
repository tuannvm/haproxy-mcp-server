package haproxy

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	runtimeclient "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/runtime"
	statsclient "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// HAProxyClient is a combined client that can interact with HAProxy through both runtime API and stats API
type HAProxyClient struct {
	RuntimeClient RuntimeClient
	StatsClient   StatsClient
	StatsURL      string
}

// ===========================================================================
// Initialization and core functionality
// ===========================================================================

// ensureRuntime verifies the runtime client is initialized.
func (c *HAProxyClient) ensureRuntime() error {
	return c.EnsureRuntime()
}

// ensureStats verifies the stats client is initialized.
func (c *HAProxyClient) ensureStats() error {
	return c.EnsureStats()
}

// NewHAProxyClient creates a new HAProxy client using the provided configurations
func NewHAProxyClient(runtimeAPIURL string, statsURL string) (*HAProxyClient, error) {
	client := &HAProxyClient{
		StatsURL: statsURL,
	}

	// Initialize runtime client if URL is provided
	if runtimeAPIURL != "" {
		slog.Info("Initializing HAProxy Runtime API client", "url", runtimeAPIURL)
		runtimeClient, err := runtimeclient.NewHAProxyClient(runtimeAPIURL)
		if err != nil {
			slog.Warn("Failed to initialize HAProxy Runtime API client", "error", err, "url", runtimeAPIURL)
			// If stats URL is provided, continue without runtime client
			if statsURL == "" {
				return nil, fmt.Errorf("failed to initialize HAProxy Runtime API client: %w", err)
			}
			slog.Info("Continuing in stats-only mode")
		} else {
			client.RuntimeClient = runtimeClient
			slog.Info("HAProxy Runtime API client initialized successfully")
		}
	} else if statsURL != "" {
		slog.Info("Running in stats-only mode (no Runtime API URL provided)")
	}

	// Initialize stats client if URL is provided
	if statsURL != "" {
		slog.Info("Initializing HAProxy Stats client", "url", statsURL)
		statsClient, err := statsclient.NewStatsClient(statsURL)
		if err != nil {
			slog.Error("Failed to initialize HAProxy Stats client", "error", err)
			// If runtime client is already initialized, continue with only runtime client
			if client.RuntimeClient == nil {
				return nil, fmt.Errorf("failed to initialize HAProxy Stats client: %w", err)
			}
		} else {
			client.StatsClient = statsClient
			slog.Info("HAProxy Stats client initialized successfully")
		}
	}

	// Ensure at least one client is initialized
	if client.RuntimeClient == nil && client.StatsClient == nil {
		return nil, fmt.Errorf("at least one of Runtime API URL or Stats URL must be provided and successfully initialized")
	}

	return client, nil
}

// Close closes both runtime and stats client connections
func (c *HAProxyClient) Close() error {
	if c.RuntimeClient != nil {
		if err := c.RuntimeClient.Close(); err != nil {
			slog.Error("Error closing runtime client", "error", err)
		}
	}

	return nil
}

// ===========================================================================
// Methods supported by both Runtime and Stats APIs
// ===========================================================================

// GetBackends returns a list of all backends
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) GetBackends() ([]string, error) {
	return c.WithApiFallbackStringSlice(
		"get backends",
		"runtime",
		func() ([]string, error) {
			return c.RuntimeClient.ListBackends()
		},
		func() ([]string, error) {
			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to get backends: %w", err)
			}

			// Extract unique backend names from stats
			backendMap := make(map[string]bool)
			for _, item := range stats.Stats {
				// Only add actual backends (type=1) in the requested backend
				if item.GetType() == 1 && item.GetProxyName() != "" {
					backendMap[item.GetProxyName()] = true
				}
			}

			// Convert map to slice
			backendList := make([]string, 0, len(backendMap))
			for name := range backendMap {
				backendList = append(backendList, name)
			}

			return backendList, nil
		},
	)
}

// GetBackendDetails returns detailed information about a backend
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) GetBackendDetails(name string) (map[string]interface{}, error) {
	return c.WithApiFallbackMap(
		"get backend details",
		"runtime",
		func() (map[string]interface{}, error) {
			info, err := c.RuntimeClient.GetBackendInfo(name)
			if err != nil {
				return nil, err
			}
			// Convert to map format - handle correctly based on GetBackendInfo return type
			result := make(map[string]interface{})
			// Conversion logic depends on actual return type of GetBackendInfo
			result["name"] = name
			result["info"] = info
			return result, nil
		},
		func() (map[string]interface{}, error) {
			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to get backend details: %w", err)
			}

			// Find the backend stats
			var backendStats statsclient.StatsItem
			var foundBackend bool
			for _, item := range stats.Stats {
				if item.GetProxyName() == name && item.GetServiceName() == "BACKEND" {
					backendStats = item
					foundBackend = true
					break
				}
			}

			if !foundBackend {
				return nil, fmt.Errorf("backend %s not found", name)
			}

			// Get server stats for this backend
			var serverStats []statsclient.StatsItem
			for _, item := range stats.Stats {
				if item.GetProxyName() == name && item.GetType() == 2 && item.GetServiceName() != "BACKEND" {
					serverStats = append(serverStats, item)
				}
			}

			// Create the backend details map
			details := map[string]interface{}{
				"name":   name,
				"status": backendStats.GetStatus(),
			}

			// Add session info
			statsclient.AddStatsValueToMap(details, "current_sessions", backendStats, "scur")

			// Add all servers
			servers := make([]map[string]interface{}, 0, len(serverStats))
			for _, server := range serverStats {
				serverInfo := map[string]interface{}{
					"name":   server.GetServiceName(),
					"status": server.GetStatus(),
				}

				// Add weight
				statsclient.AddStatsValueToMap(serverInfo, "weight", server, "weight")

				servers = append(servers, serverInfo)
			}
			details["servers"] = servers

			return details, nil
		},
	)
}

// ListServers returns a list of servers for a backend
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	return c.WithApiFallbackStringSlice(
		"list servers",
		"runtime",
		func() ([]string, error) {
			return c.RuntimeClient.ListServers(backend)
		},
		func() ([]string, error) {
			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to list servers: %w", err)
			}

			// Extract server names for the specified backend
			serverMap := make(map[string]bool)
			for _, item := range stats.Stats {
				// Look for server entries (type=2) in the requested backend
				if item.GetType() == 2 && item.GetProxyName() == backend && item.GetServiceName() != "BACKEND" && item.GetServiceName() != "FRONTEND" {
					serverMap[item.GetServiceName()] = true
				}
			}

			// Convert map to slice
			serverList := make([]string, 0, len(serverMap))
			for name := range serverMap {
				serverList = append(serverList, name)
			}

			return serverList, nil
		},
	)
}

// GetServerDetails returns detailed information about a server
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	return c.WithApiFallbackMap(
		"get server details",
		"runtime",
		func() (map[string]interface{}, error) {
			return c.RuntimeClient.GetServerDetails(backend, server)
		},
		func() (map[string]interface{}, error) {
			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to get server details: %w", err)
			}

			// Find the matching server in stats
			for _, item := range stats.Stats {
				if item.GetProxyName() == backend && item.GetServiceName() == server {
					// Convert to map format
					details := map[string]interface{}{
						"name":    server,
						"backend": backend,
						"status":  item.GetStatus(),
					}

					// Use helper functions for data extraction
					statsclient.AddStatsValueToMap(details, "type", item, "type")
					statsclient.AddStatsValueToMap(details, "weight", item, "weight")
					statsclient.AddStatsValueToMap(details, "current_sessions", item, "scur")
					statsclient.AddStatsValueToMap(details, "max_sessions", item, "smax")
					statsclient.AddStatsInt64ValueToMap(details, "bytes_in", item, "bin")
					statsclient.AddStatsInt64ValueToMap(details, "bytes_out", item, "bout")

					return details, nil
				}
			}

			return nil, fmt.Errorf("server %s not found in backend %s", server, backend)
		},
	)
}

// ShowStat executes the show stat command
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) ShowStat(filter string) ([]map[string]string, error) {
	return c.WithApiFallbackStringMapSlice(
		"show stat",
		"stats", // Try stats first, then runtime as fallback
		func() ([]map[string]string, error) {
			if err := c.ensureStats(); err != nil {
				return nil, err
			}

			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to get statistics: %w", err)
			}

			result := []map[string]string{}
			for _, item := range stats.Stats {
				if filter == "" || strings.Contains(item.GetProxyName(), filter) || strings.Contains(item.GetServiceName(), filter) {
					row := map[string]string{
						"pxname": item.GetProxyName(),
						"svname": item.GetServiceName(),
						"status": item.GetStatus(),
					}

					// Use helper functions for value extraction
					weight := statsclient.GetStatsItemWeight(item)
					if weight != 0 {
						row["weight"] = fmt.Sprintf("%d", weight)
					}

					typeVal := statsclient.GetStatsItemType(item)
					if typeVal != 0 {
						row["type"] = fmt.Sprintf("%d", typeVal)
					}

					sessions := statsclient.GetStatsItemSessions(item)
					if sessions != 0 {
						row["scur"] = fmt.Sprintf("%d", sessions)
					}

					maxSessions := statsclient.GetStatsItemMaxSessions(item)
					if maxSessions != 0 {
						row["smax"] = fmt.Sprintf("%d", maxSessions)
					}

					bytesIn := statsclient.GetStatsItemBytesIn(item)
					if bytesIn != 0 {
						row["bin"] = fmt.Sprintf("%d", bytesIn)
					}

					bytesOut := statsclient.GetStatsItemBytesOut(item)
					if bytesOut != 0 {
						row["bout"] = fmt.Sprintf("%d", bytesOut)
					}

					result = append(result, row)
				}
			}
			return result, nil
		},
		func() ([]map[string]string, error) {
			if err := c.ensureRuntime(); err != nil {
				return nil, err
			}

			cmd := "show stat"
			if filter != "" {
				cmd = fmt.Sprintf("%s %s", cmd, filter)
			}

			response, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
			if err != nil {
				return nil, fmt.Errorf("failed to execute runtime command: %w", err)
			}

			// Parse CSV-like output
			result := []map[string]string{}
			lines := strings.Split(response, "\n")
			if len(lines) > 0 {
				headers := strings.Split(lines[0], ",")
				for i := 1; i < len(lines); i++ {
					if lines[i] == "" {
						continue
					}
					values := strings.Split(lines[i], ",")
					row := make(map[string]string)
					for j := 0; j < len(headers) && j < len(values); j++ {
						row[headers[j]] = values[j]
					}
					result = append(result, row)
				}
			}
			return result, nil
		},
	)
}

// ShowServersState returns server state information
// Supported by both Runtime and Stats APIs
func (c *HAProxyClient) ShowServersState(backend string) ([]map[string]string, error) {
	return c.WithApiFallbackStringMapSlice(
		"show servers state",
		"runtime",
		func() ([]map[string]string, error) {
			if err := c.ensureRuntime(); err != nil {
				return nil, err
			}
			cmd := "show servers state"
			if backend != "" {
				cmd = fmt.Sprintf("%s %s", cmd, backend)
			}

			response, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
			if err != nil {
				return nil, err
			}

			// Parse output (simplified implementation)
			result := []map[string]string{}
			lines := strings.Split(response, "\n")
			if len(lines) > 0 {
				headers := strings.Fields(lines[0])
				for i := 1; i < len(lines); i++ {
					if lines[i] == "" {
						continue
					}
					values := strings.Fields(lines[i])
					row := make(map[string]string)
					for j := 0; j < len(headers) && j < len(values); j++ {
						row[headers[j]] = values[j]
					}
					result = append(result, row)
				}
			}
			return result, nil
		},
		func() ([]map[string]string, error) {
			stats, err := c.StatsClient.GetStats()
			if err != nil {
				return nil, fmt.Errorf("failed to show servers state: %w", err)
			}

			// Convert stats to server state format
			result := []map[string]string{}
			for _, item := range stats.Stats {
				// Filter by backend if specified
				if backend != "" && item.GetProxyName() != backend {
					continue
				}

				// Skip backend entries, only include servers
				if item.GetServiceName() == "BACKEND" || item.GetServiceName() == "FRONTEND" {
					continue
				}

				// Create a server state entry
				row := map[string]string{
					"be_name":  item.GetProxyName(),
					"srv_name": item.GetServiceName(),
					"status":   item.GetStatus(),
				}

				// Add additional fields if available
				addr, ok := item.GetString("addr")
				if ok && addr != "" {
					row["srv_addr"] = addr
				}

				// Add weight
				weight := statsclient.GetStatsItemWeight(item)
				if weight != 0 {
					row["weight"] = fmt.Sprintf("%d", weight)
				}

				// Handle last state change
				lastChg, ok := item.GetInt("lastchg")
				if ok && lastChg != 0 {
					row["last_change"] = fmt.Sprintf("%d", lastChg)
				}

				// Handle server state
				status := item.GetStatus()
				switch status {
				case "UP":
					row["srv_op_state"] = "active"
				case "DOWN":
					row["srv_op_state"] = "down"
				case "MAINT":
					row["srv_op_state"] = "maint"
				default:
					row["srv_op_state"] = "unknown"
				}

				result = append(result, row)
			}

			if len(result) == 0 && backend != "" {
				slog.Warn("No servers found for backend in stats response", "backend", backend)
			}

			return result, nil
		},
	)
}

// ===========================================================================
// Methods supported by Runtime API only
// ===========================================================================

// ExecuteRuntimeCommand executes a command on HAProxy's Runtime API
// Requires Runtime API
func (c *HAProxyClient) ExecuteRuntimeCommand(command string) (string, error) {
	return c.ExecuteRuntimeCommandWithContext(context.Background(), command)
}

// ExecuteRuntimeCommandWithContext executes a command on HAProxy's Runtime API with context
// Requires Runtime API
func (c *HAProxyClient) ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error) {
	if c.RuntimeClient == nil {
		return "", fmt.Errorf("runtime client is not initialized (HAPROXY_RUNTIME_ENABLED=false or runtime connection failed)")
	}

	// Use context-aware version if available
	if ctxClient, ok := c.RuntimeClient.(interface {
		ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error)
	}); ok {
		return ctxClient.ExecuteRuntimeCommandWithContext(ctx, command)
	}

	// Fall back to non-context version
	return c.RuntimeClient.ExecuteRuntimeCommand(command)
}

// GetRuntimeInfo retrieves HAProxy process information from runtime API
// Supported by both Runtime and Stats APIs with different capabilities
func (c *HAProxyClient) GetRuntimeInfo() (map[string]string, error) {
	// If Runtime API is available, use it for complete info
	if c.RuntimeClient != nil {
		return c.RuntimeClient.GetProcessInfo()
	}

	// If only Stats API is available, provide limited information
	if c.StatsClient != nil {
		// Create a basic info map with limited data available from stats
		info := map[string]string{
			"mode":                "stats-only",
			"stats_url":           c.StatsURL,
			"runtime_api_enabled": "false",
			"note":                "Limited information available in stats-only mode",
		}

		// Try to get some version or uptime info from stats if possible
		stats, err := c.StatsClient.GetStats()
		if err == nil && len(stats.Stats) > 0 {
			// Check if we can find any useful info in stats
			for _, item := range stats.Stats {
				// Backend typically has more metadata
				if item.GetServiceName() == "BACKEND" || item.GetServiceName() == "FRONTEND" {
					mode, ok := item.GetString("mode")
					if ok && mode != "" {
						info["proxy_mode"] = mode
					}
					break
				}
			}
			info["backends_count"] = fmt.Sprintf("%d", len(c.getUniqueBackendsFromStats(stats)))
			info["status"] = "running"
		} else {
			info["status"] = "unknown"
			info["connection_status"] = "Stats API accessible but no data available"
		}

		return info, nil
	}

	return nil, fmt.Errorf("no available API clients (neither Runtime nor Stats API initialized)")
}

// getUniqueBackendsFromStats is a helper method to count unique backends in stats
func (c *HAProxyClient) getUniqueBackendsFromStats(stats *statsclient.HAProxyStats) []string {
	backendMap := make(map[string]bool)

	for _, item := range stats.Stats {
		if item.GetType() == 1 && item.GetProxyName() != "" {
			backendMap[item.GetProxyName()] = true
		}
	}

	backends := make([]string, 0, len(backendMap))
	for name := range backendMap {
		backends = append(backends, name)
	}

	return backends
}

// EnableServer enables a server in a backend
// Requires Runtime API
func (c *HAProxyClient) EnableServer(backend, server string) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}
	return c.RuntimeClient.EnableServer(backend, server)
}

// DisableServer disables a server in a backend
// Requires Runtime API
func (c *HAProxyClient) DisableServer(backend, server string) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}
	return c.RuntimeClient.DisableServer(backend, server)
}

// SetWeight sets the weight for a server in a backend
// Requires Runtime API
func (c *HAProxyClient) SetWeight(backend, server string, weight int) (string, error) {
	if err := c.ensureRuntime(); err != nil {
		return "", err
	}

	// Directly execute the command since it might be different across versions
	cmd := fmt.Sprintf("set weight %s/%s %d", backend, server, weight)
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Weight for %s/%s set to %d", backend, server, weight), nil
}

// SetServerMaxconn sets the maximum connections for a server
// Requires Runtime API
func (c *HAProxyClient) SetServerMaxconn(backend, server string, maxconn int) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}
	return c.RuntimeClient.SetServerMaxconn(backend, server, maxconn)
}

// EnableHealth enables health checks for a server
// Requires Runtime API
func (c *HAProxyClient) EnableHealth(backend, server string) error {
	return c.toggleCheck("enable", "health", backend, server)
}

// DisableHealth disables health checks for a server
// Requires Runtime API
func (c *HAProxyClient) DisableHealth(backend, server string) error {
	return c.toggleCheck("disable", "health", backend, server)
}

// EnableAgent enables agent checks for a server
// Requires Runtime API
func (c *HAProxyClient) EnableAgent(backend, server string) error {
	return c.toggleCheck("enable", "agent", backend, server)
}

// DisableAgent disables agent checks for a server
// Requires Runtime API
func (c *HAProxyClient) DisableAgent(backend, server string) error {
	return c.toggleCheck("disable", "agent", backend, server)
}

// DumpStatsFile dumps stats to a file
// Requires Runtime API
func (c *HAProxyClient) DumpStatsFile(filepath string) (string, error) {
	if err := c.ensureRuntime(); err != nil {
		return "", err
	}

	cmd := fmt.Sprintf("show stat > %s", filepath)
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	if err != nil {
		return "", err
	}

	return filepath, nil
}

// DebugCounters returns debug counters
// Requires Runtime API
func (c *HAProxyClient) DebugCounters() (map[string]interface{}, error) {
	if err := c.ensureRuntime(); err != nil {
		return nil, err
	}

	response, err := c.RuntimeClient.ExecuteRuntimeCommand("debug dev state")
	if err != nil {
		return nil, err
	}

	// Parse output into structured format
	counters := make(map[string]interface{})
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			counters[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return counters, nil
}

// ClearCountersAll clears all counters
// Requires Runtime API
func (c *HAProxyClient) ClearCountersAll() error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}

	_, err := c.RuntimeClient.ExecuteRuntimeCommand("clear counters all")
	return err
}

// AddServer adds a server to a backend
// Requires Runtime API
func (c *HAProxyClient) AddServer(backend, name, addr string, port, weight int) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}

	cmd := fmt.Sprintf("add server %s/%s %s", backend, name, addr)
	if port > 0 {
		cmd = fmt.Sprintf("%s:%d", cmd, port)
	}
	if weight > 0 {
		cmd = fmt.Sprintf("%s weight %d", cmd, weight)
	}

	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	return err
}

// DelServer removes a server from a backend
// Requires Runtime API
func (c *HAProxyClient) DelServer(backend, name string) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}

	cmd := fmt.Sprintf("del server %s/%s", backend, name)
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	return err
}

// ReloadHAProxy reloads the HAProxy configuration
// Requires Runtime API
func (c *HAProxyClient) ReloadHAProxy() error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}

	_, err := c.RuntimeClient.ExecuteRuntimeCommand("reload")
	return err
}

// toggleCheck performs enable/disable for health or agent checks
// This is a private helper for Runtime API methods
func (c *HAProxyClient) toggleCheck(action, checkType, backend, server string) error {
	if err := c.ensureRuntime(); err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s %s %s/%s", action, checkType, backend, server)
	_, err := c.ExecuteRuntimeCommand(cmd)
	return err
}

// ===========================================================================
// Methods supported by Stats API only
// ===========================================================================

// GetStats retrieves HAProxy statistics from stats page
// Requires Stats API
func (c *HAProxyClient) GetStats() (*statsclient.HAProxyStats, error) {
	if err := c.ensureStats(); err != nil {
		return nil, err
	}
	return c.StatsClient.GetStats()
}
