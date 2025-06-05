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

// doRequest performs an HTTP request and processes the response
func (c *StatsClient) doRequest(url string, description string) ([]byte, error) {
	slog.Info(fmt.Sprintf("Fetching %s", description), "url", url)

	// Make HTTP request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", description, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("Error closing response body", "error", closeErr)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s request failed with status code: %d", description, resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", description, err)
	}

	// Log raw response data for debugging
	slog.Debug(fmt.Sprintf("%s raw response", description),
		"content_type", resp.Header.Get("Content-Type"),
		"content_length", len(body),
		"response_start", string(body[:min(100, len(body))]))

	return body, nil
}

// GetStats fetches statistics from HAProxy stats page
func (c *StatsClient) GetStats() (*HAProxyStats, error) {
	// Construct URL for JSON stats
	statsURL := c.StatsURL
	// Make sure we're requesting JSON format
	if !containsSubstring(statsURL, ";json") && !containsSubstring(statsURL, "/stats;json") {
		statsURL = appendPath(statsURL, ";json")
	}

	// Get response body
	body, err := c.doRequest(statsURL, "HAProxy stats")
	if err != nil {
		return nil, err
	}

	// First, try to unmarshal as a generic interface (could be array or object)
	var rawData interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		slog.Error("Failed to parse HAProxy stats response",
			"error", err,
			"content_type", "application/json",
			"response_start", string(body[:min(200, len(body))]))
		return nil, fmt.Errorf("failed to parse HAProxy stats response: %w", err)
	}

	// Convert the generic interface data to our StatsItem objects
	stats := &HAProxyStats{
		Stats: []StatsItem{},
	}

	// Process based on the type we received
	switch data := rawData.(type) {
	case []interface{}:
		// Handle array response
		slog.Info("Processing HAProxy stats as array", "count", len(data))
		for _, item := range data {
			if mapItem, ok := item.(map[string]interface{}); ok {
				stats.Stats = append(stats.Stats, NewStatsItem(mapItem))
			}
		}
	case map[string]interface{}:
		// Handle object response
		slog.Info("Processing HAProxy stats as object")
		// Check if there's a "stats" field containing an array
		if statsArray, ok := data["stats"].([]interface{}); ok {
			for _, item := range statsArray {
				if mapItem, ok := item.(map[string]interface{}); ok {
					stats.Stats = append(stats.Stats, NewStatsItem(mapItem))
				}
			}
		} else {
			// Treat the whole object as a single stats item
			stats.Stats = append(stats.Stats, NewStatsItem(data))
		}
	default:
		slog.Warn("Unexpected HAProxy stats response format",
			"type", fmt.Sprintf("%T", rawData),
			"data_preview", fmt.Sprintf("%v", rawData)[:min(100, len(fmt.Sprintf("%v", rawData)))])
		return nil, fmt.Errorf("unexpected HAProxy stats response format: %T", rawData)
	}

	if len(stats.Stats) == 0 {
		slog.Warn("No stats items found in HAProxy response")
	} else {
		slog.Info("Successfully parsed HAProxy stats", "count", len(stats.Stats))
	}

	return stats, nil
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

	// Get response body
	body, err := c.doRequest(schemaURL, "HAProxy stats schema")
	if err != nil {
		return nil, err
	}

	var schema StatsSchema
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse HAProxy stats schema: %w", err)
	}

	return &schema, nil
}

// filterStatsByType returns stats items matching the specified type
func (c *StatsClient) filterStatsByType(stats *HAProxyStats, itemType int) []common.StatItem {
	var result []common.StatItem

	for _, item := range stats.Stats {
		if item.GetType() == itemType {
			result = append(result, common.StatItem{
				ProxyName:   item.GetProxyName(),
				ServiceName: item.GetServiceName(),
				Type:        item.GetType(),
				Status:      item.GetStatus(),
				Weight:      item.GetWeight(),
			})
		}
	}

	return result
}

// GetFrontends returns all frontend stats
func (c *StatsClient) GetFrontends(stats *HAProxyStats) []common.StatItem {
	return c.filterStatsByType(stats, 0) // Type 0 is frontend
}

// GetBackends returns all backend stats
func (c *StatsClient) GetBackends(stats *HAProxyStats) []common.StatItem {
	return c.filterStatsByType(stats, 1) // Type 1 is backend
}

// GetServers returns all server stats
func (c *StatsClient) GetServers(stats *HAProxyStats) []common.StatItem {
	return c.filterStatsByType(stats, 2) // Type 2 is server
}

// GetServersByBackend returns all server stats for a specific backend
func (c *StatsClient) GetServersByBackend(stats *HAProxyStats, backendName string) []common.StatItem {
	var servers []common.StatItem

	for _, item := range stats.Stats {
		if item.GetType() == 2 && item.GetProxyName() == backendName { // Type 2 is server
			servers = append(servers, common.StatItem{
				ProxyName:   item.GetProxyName(),
				ServiceName: item.GetServiceName(),
				Type:        item.GetType(),
				Status:      item.GetStatus(),
				Weight:      item.GetWeight(),
			})
		}
	}

	return servers
}

// Min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}

// Helper function to append a path to a URL, handling trailing slashes properly
func appendPath(baseURL, path string) string {
	if baseURL == "" {
		return path
	}

	if path == "" {
		return baseURL
	}

	// If path starts with a slash and baseURL ends with a slash, remove one
	if baseURL[len(baseURL)-1] == '/' && path[0] == '/' {
		return baseURL + path[1:]
	}

	// If neither has a slash, add one
	if baseURL[len(baseURL)-1] != '/' && path[0] != '/' {
		return baseURL + "/" + path
	}

	// Otherwise, just concatenate
	return baseURL + path
}

// FilterStats filters the stats by proxy name and/or service name
func (c *StatsClient) FilterStats(stats *HAProxyStats, proxyName, serviceName string) []common.StatItem {
	var filtered []common.StatItem

	for _, item := range stats.Stats {
		// Apply proxy name filter if provided
		if proxyName != "" && item.GetProxyName() != proxyName {
			continue
		}

		// Apply service name filter if provided
		if serviceName != "" && item.GetServiceName() != serviceName {
			continue
		}

		filtered = append(filtered, common.StatItem{
			ProxyName:   item.GetProxyName(),
			ServiceName: item.GetServiceName(),
			Type:        item.GetType(),
			Status:      item.GetStatus(),
			Weight:      item.GetWeight(),
		})
	}

	return filtered
}
