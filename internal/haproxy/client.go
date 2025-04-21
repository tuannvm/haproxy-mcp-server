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
			return nil, fmt.Errorf("failed to initialize HAProxy Runtime API client: %w", err)
		}
		client.RuntimeClient = runtimeClient
		slog.Info("HAProxy Runtime API client initialized successfully")
	}

	// Initialize stats client if URL is provided
	if statsURL != "" {
		slog.Info("Initializing HAProxy Stats client", "url", statsURL)
		statsClient, err := statsclient.NewStatsClient(statsURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize HAProxy Stats client: %w", err)
		}
		client.StatsClient = statsClient
		slog.Info("HAProxy Stats client initialized successfully")
	}

	// Ensure at least one client is initialized
	if client.RuntimeClient == nil && client.StatsClient == nil {
		return nil, fmt.Errorf("at least one of Runtime API URL or Stats URL must be provided")
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

// ExecuteRuntimeCommand executes a command on HAProxy's Runtime API
func (c *HAProxyClient) ExecuteRuntimeCommand(command string) (string, error) {
	return c.ExecuteRuntimeCommandWithContext(context.Background(), command)
}

// ExecuteRuntimeCommandWithContext executes a command on HAProxy's Runtime API with context
func (c *HAProxyClient) ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error) {
	if c.RuntimeClient == nil {
		return "", fmt.Errorf("runtime client is not initialized")
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
func (c *HAProxyClient) GetRuntimeInfo() (map[string]string, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.GetProcessInfo()
}

// GetStats retrieves HAProxy statistics from stats page
func (c *HAProxyClient) GetStats() (*statsclient.HAProxyStats, error) {
	if c.StatsClient == nil {
		return nil, fmt.Errorf("stats client is not initialized")
	}
	return c.StatsClient.GetStats()
}

// GetBackends returns a list of all backends
func (c *HAProxyClient) GetBackends() ([]string, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.ListBackends()
}

// GetBackendDetails returns detailed information about a backend
func (c *HAProxyClient) GetBackendDetails(name string) (map[string]interface{}, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
	}
	info, err := c.RuntimeClient.GetBackendInfo(name)
	if err != nil {
		return nil, err
	}

	// Convert to map format - handle correctly based on GetBackendInfo return type
	result := make(map[string]interface{})
	// Conversion logic depends on actual return type of GetBackendInfo
	// This is a simplified approach
	result["name"] = name
	result["info"] = info

	return result, nil
}

// ListServers returns a list of servers for a backend
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.ListServers(backend)
}

// GetServerDetails returns detailed information about a server
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
	}
	serverInfo, err := c.RuntimeClient.GetServerDetails(backend, server)
	if err != nil {
		return nil, err
	}

	// Convert to map format directly based on the returned structure
	// This assumes the runtime client returns a map[string]interface{}
	return serverInfo, nil
}

// EnableServer enables a server in a backend
func (c *HAProxyClient) EnableServer(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.EnableServer(backend, server)
}

// DisableServer disables a server in a backend
func (c *HAProxyClient) DisableServer(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.DisableServer(backend, server)
}

// SetWeight sets the weight for a server in a backend
func (c *HAProxyClient) SetWeight(backend, server string, weight int) (string, error) {
	if c.RuntimeClient == nil {
		return "", fmt.Errorf("runtime client is not initialized")
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
func (c *HAProxyClient) SetServerMaxconn(backend, server string, maxconn int) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	return c.RuntimeClient.SetServerMaxconn(backend, server, maxconn)
}

// EnableHealth enables health checks for a server
func (c *HAProxyClient) EnableHealth(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	// Use the correct method in the runtime client
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(fmt.Sprintf("enable health %s/%s", backend, server))
	return err
}

// DisableHealth disables health checks for a server
func (c *HAProxyClient) DisableHealth(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	// Use the correct method in the runtime client
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(fmt.Sprintf("disable health %s/%s", backend, server))
	return err
}

// EnableAgent enables agent checks for a server
func (c *HAProxyClient) EnableAgent(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	// Use the correct method in the runtime client
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(fmt.Sprintf("enable agent %s/%s", backend, server))
	return err
}

// DisableAgent disables agent checks for a server
func (c *HAProxyClient) DisableAgent(backend, server string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}
	// Use the correct method in the runtime client
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(fmt.Sprintf("disable agent %s/%s", backend, server))
	return err
}

// ShowStat executes the show stat command
func (c *HAProxyClient) ShowStat(filter string) ([]map[string]string, error) {
	// Try stats client first if available
	if c.StatsClient != nil {
		stats, err := c.StatsClient.GetStats()
		if err == nil {
			result := []map[string]string{}
			for _, item := range stats.Stats {
				if filter == "" || strings.Contains(item.PxName, filter) || strings.Contains(item.SvName, filter) {
					row := map[string]string{
						"pxname": item.PxName,
						"svname": item.SvName,
						"status": item.Status,
						"weight": fmt.Sprintf("%d", item.Weight),
					}
					result = append(result, row)
				}
			}
			return result, nil
		}
		// Fall back to runtime client if stats client failed
		slog.Warn("Stats client failed, falling back to runtime client", "error", err)
	}

	// Use runtime client as fallback or primary if stats client not available
	if c.RuntimeClient != nil {
		cmd := "show stat"
		if filter != "" {
			cmd = fmt.Sprintf("%s %s", cmd, filter)
		}
		response, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
		if err != nil {
			return nil, err
		}

		// Parse CSV-like output (simplified implementation)
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
	}

	return nil, fmt.Errorf("neither stats client nor runtime client is initialized")
}

// ShowServersState returns server state information
func (c *HAProxyClient) ShowServersState(backend string) ([]map[string]string, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
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
}

// Other methods like DumpStatsFile, DebugCounters, etc. would be added here
// to support the full tool set in tools.go

// DumpStatsFile dumps stats to a file
func (c *HAProxyClient) DumpStatsFile(filepath string) (string, error) {
	if c.RuntimeClient == nil {
		return "", fmt.Errorf("runtime client is not initialized")
	}

	cmd := fmt.Sprintf("show stat > %s", filepath)
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	if err != nil {
		return "", err
	}

	return filepath, nil
}

// DebugCounters returns debug counters
func (c *HAProxyClient) DebugCounters() (map[string]interface{}, error) {
	if c.RuntimeClient == nil {
		return nil, fmt.Errorf("runtime client is not initialized")
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
func (c *HAProxyClient) ClearCountersAll() error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}

	_, err := c.RuntimeClient.ExecuteRuntimeCommand("clear counters all")
	return err
}

// AddServer adds a server to a backend
func (c *HAProxyClient) AddServer(backend, name, addr string, port, weight int) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
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
func (c *HAProxyClient) DelServer(backend, name string) error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}

	cmd := fmt.Sprintf("del server %s/%s", backend, name)
	_, err := c.RuntimeClient.ExecuteRuntimeCommand(cmd)
	return err
}

// ReloadHAProxy reloads the HAProxy configuration
func (c *HAProxyClient) ReloadHAProxy() error {
	if c.RuntimeClient == nil {
		return fmt.Errorf("runtime client is not initialized")
	}

	// Note: This is a simplified implementation
	// In a real-world scenario, this would likely call a system command
	// to trigger a reload of HAProxy
	_, err := c.RuntimeClient.ExecuteRuntimeCommand("reload")
	return err
}
