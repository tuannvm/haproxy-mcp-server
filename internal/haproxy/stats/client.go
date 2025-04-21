package stats

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// NewStatsClient creates a new HAProxy stats client
func NewStatsClient(statsURL string) (*StatsClient, error) {
	// Validate URL
	_, err := url.Parse(statsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid stats URL: %w", err)
	}

	return &StatsClient{
		StatsURL: statsURL,
	}, nil
}

// GetStats fetches statistics from HAProxy stats page
func (c *StatsClient) GetStats() (*HAProxyStats, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct URL for JSON stats
	statsURL := c.StatsURL
	if statsURL[len(statsURL)-1] != '/' {
		statsURL += "/"
	}
	if statsURL[len(statsURL)-6:] != ";json" {
		statsURL += ";json"
	}

	// Make HTTP request
	resp, err := client.Get(statsURL)
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
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Construct URL for schema
	schemaURL := c.StatsURL
	if schemaURL[len(schemaURL)-1] != '/' {
		schemaURL += "/"
	}
	// Replace ;json with ;json-schema if present, otherwise append it
	if schemaURL[len(schemaURL)-6:] == ";json" {
		schemaURL = schemaURL[:len(schemaURL)-5] + "-schema"
	} else {
		schemaURL += ";json-schema"
	}

	// Make HTTP request
	resp, err := client.Get(schemaURL)
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
func (c *StatsClient) FilterStats(stats *HAProxyStats, proxyName, serviceName string) []StatsItem {
	var filtered []StatsItem

	for _, item := range stats.Stats {
		// Apply proxy name filter if provided
		if proxyName != "" && item.PxName != proxyName {
			continue
		}

		// Apply service name filter if provided
		if serviceName != "" && item.SvName != serviceName {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// GetFrontends returns all frontend stats
func (c *StatsClient) GetFrontends(stats *HAProxyStats) []StatsItem {
	var frontends []StatsItem

	for _, item := range stats.Stats {
		if item.Type == 0 { // Type 0 is frontend
			frontends = append(frontends, item)
		}
	}

	return frontends
}

// GetBackends returns all backend stats
func (c *StatsClient) GetBackends(stats *HAProxyStats) []StatsItem {
	var backends []StatsItem

	for _, item := range stats.Stats {
		if item.Type == 1 { // Type 1 is backend
			backends = append(backends, item)
		}
	}

	return backends
}

// GetServers returns all server stats
func (c *StatsClient) GetServers(stats *HAProxyStats) []StatsItem {
	var servers []StatsItem

	for _, item := range stats.Stats {
		if item.Type == 2 { // Type 2 is server
			servers = append(servers, item)
		}
	}

	return servers
}

// GetServersByBackend returns all server stats for a specific backend
func (c *StatsClient) GetServersByBackend(stats *HAProxyStats, backendName string) []StatsItem {
	var servers []StatsItem

	for _, item := range stats.Stats {
		if item.Type == 2 && item.PxName == backendName { // Type 2 is server
			servers = append(servers, item)
		}
	}

	return servers
}
