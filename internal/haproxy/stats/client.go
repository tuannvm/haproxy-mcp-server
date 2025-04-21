package stats

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
)

// StatsClient is a client for fetching HAProxy stats from the stats page
type StatsClient struct {
	StatsURL   string       // URL to HAProxy stats page (e.g., http://127.0.0.1:1936/;json)
	httpClient *http.Client // Shared HTTP client
}

// NewStatsClient creates a new HAProxy stats client
func NewStatsClient(statsURL string) (*StatsClient, error) {
	// Validate URL
	_, err := url.Parse(statsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid stats URL: %w", err)
	}

	return &StatsClient{
		StatsURL: statsURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// buildURL builds a URL with the given suffix
func (c *StatsClient) buildURL(suffix string) string {
	baseURL := c.StatsURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	return baseURL + suffix
}

// GetStats fetches statistics from HAProxy stats page
func (c *StatsClient) GetStats() (*HAProxyStats, error) {
	// Construct URL for JSON stats
	statsURL := c.buildURL(";json")

	// Make HTTP request
	resp, err := c.httpClient.Get(statsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HAProxy stats: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("Error closing response body", "error", closeErr)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HAProxy stats request failed with status code: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HAProxy stats response: %w", err)
	}

	var stats HAProxyStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse HAProxy stats: %w", err)
	}

	return &stats, nil
}

// GetSchema fetches the JSON schema for HAProxy stats
func (c *StatsClient) GetSchema() (*StatsSchema, error) {
	// Construct URL for schema
	schemaURL := c.StatsURL
	if schemaURL[len(schemaURL)-6:] == ";json" {
		schemaURL = schemaURL[:len(schemaURL)-5] + "-schema"
	} else {
		schemaURL = c.buildURL(";json-schema")
	}

	// Make HTTP request
	resp, err := c.httpClient.Get(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HAProxy stats schema: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("Error closing response body", "error", closeErr)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HAProxy stats schema request failed with status code: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HAProxy stats schema response: %w", err)
	}

	var schema StatsSchema
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse HAProxy stats schema: %w", err)
	}

	return &schema, nil
}

// FilterStats filters the stats by proxy name and/or service name
func (c *StatsClient) FilterStats(stats *HAProxyStats, proxyName, serviceName string) []common.StatItem {
	var filtered []common.StatItem

	for _, item := range stats.Stats {
		// Apply proxy name filter if provided
		if proxyName != "" && item.PxName != proxyName {
			continue
		}

		// Apply service name filter if provided
		if serviceName != "" && item.SvName != serviceName {
			continue
		}

		filtered = append(filtered, common.StatItem{
			ProxyName:   item.PxName,
			ServiceName: item.SvName,
			Type:        item.Type,
			Status:      item.Status,
			Weight:      item.Weight,
		})
	}

	return filtered
}

// GetFrontends returns all frontend stats
func (c *StatsClient) GetFrontends(stats *HAProxyStats) []common.StatItem {
	var frontends []common.StatItem

	for _, item := range stats.Stats {
		if item.Type == 0 { // Type 0 is frontend
			frontends = append(frontends, common.StatItem{
				ProxyName:   item.PxName,
				ServiceName: item.SvName,
				Type:        item.Type,
				Status:      item.Status,
				Weight:      item.Weight,
			})
		}
	}

	return frontends
}

// GetBackends returns all backend stats
func (c *StatsClient) GetBackends(stats *HAProxyStats) []common.StatItem {
	var backends []common.StatItem

	for _, item := range stats.Stats {
		if item.Type == 1 { // Type 1 is backend
			backends = append(backends, common.StatItem{
				ProxyName:   item.PxName,
				ServiceName: item.SvName,
				Type:        item.Type,
				Status:      item.Status,
				Weight:      item.Weight,
			})
		}
	}

	return backends
}

// GetServers returns all server stats
func (c *StatsClient) GetServers(stats *HAProxyStats) []common.StatItem {
	var servers []common.StatItem

	for _, item := range stats.Stats {
		if item.Type == 2 { // Type 2 is server
			servers = append(servers, common.StatItem{
				ProxyName:   item.PxName,
				ServiceName: item.SvName,
				Type:        item.Type,
				Status:      item.Status,
				Weight:      item.Weight,
			})
		}
	}

	return servers
}

// GetServersByBackend returns all server stats for a specific backend
func (c *StatsClient) GetServersByBackend(stats *HAProxyStats, backendName string) []common.StatItem {
	var servers []common.StatItem

	for _, item := range stats.Stats {
		if item.Type == 2 && item.PxName == backendName { // Type 2 is server
			servers = append(servers, common.StatItem{
				ProxyName:   item.PxName,
				ServiceName: item.SvName,
				Type:        item.Type,
				Status:      item.Status,
				Weight:      item.Weight,
			})
		}
	}

	return servers
}
