package testing

import (
	"fmt"

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// MockStatsClient implements the haproxy.StatsClient interface for testing
type MockStatsClient struct {
	haproxy.StatsClient // Embedded interface for proper implementation

	// Configuration for mocking behavior
	FailGetStats  bool
	FailGetSchema bool

	// Mocked return values
	Stats  *stats.HAProxyStats
	Schema *stats.StatsSchema
}

// NewMockStatsClient creates a new mock stats client with default settings
func NewMockStatsClient() *MockStatsClient {
	return &MockStatsClient{
		Stats: &stats.HAProxyStats{
			Stats: []stats.StatsItem{
				{
					PxName: "backend1",
					SvName: "BACKEND",
					Type:   1, // Backend type
					Status: "UP",
					Weight: 100,
				},
				{
					PxName: "backend1",
					SvName: "server1",
					Type:   2, // Server type
					Status: "UP",
					Weight: 100,
				},
			},
		},
		Schema: &stats.StatsSchema{
			Title:       "HAProxy Stats Schema",
			Description: "Schema for HAProxy statistics",
			Type:        "object",
			Properties:  map[string]stats.Property{},
		},
	}
}

// GetStats implements StatsClient.GetStats
func (m *MockStatsClient) GetStats() (*stats.HAProxyStats, error) {
	if m.FailGetStats {
		return nil, fmt.Errorf("mock error getting stats")
	}
	return m.Stats, nil
}

// GetSchema implements StatsClient.GetSchema
func (m *MockStatsClient) GetSchema() (*stats.StatsSchema, error) {
	if m.FailGetSchema {
		return nil, fmt.Errorf("mock error getting schema")
	}
	return m.Schema, nil
}

// FilterStats implements StatsClient.FilterStats
func (m *MockStatsClient) FilterStats(stats *stats.HAProxyStats, proxyName, serviceName string) []common.StatItem {
	var filtered []common.StatItem

	for _, item := range stats.Stats {
		if (proxyName == "" || item.PxName == proxyName) &&
			(serviceName == "" || item.SvName == serviceName) {
			filtered = append(filtered, common.StatItem{
				ProxyName:   item.PxName,
				ServiceName: item.SvName,
				Type:        item.Type,
				Status:      item.Status,
				Weight:      item.Weight,
			})
		}
	}

	return filtered
}

// GetFrontends implements StatsClient.GetFrontends
func (m *MockStatsClient) GetFrontends(stats *stats.HAProxyStats) []common.StatItem {
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

// GetBackends implements StatsClient.GetBackends
func (m *MockStatsClient) GetBackends(stats *stats.HAProxyStats) []common.StatItem {
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

// GetServers implements StatsClient.GetServers
func (m *MockStatsClient) GetServers(stats *stats.HAProxyStats) []common.StatItem {
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

// GetServersByBackend implements StatsClient.GetServersByBackend
func (m *MockStatsClient) GetServersByBackend(stats *stats.HAProxyStats, backendName string) []common.StatItem {
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

// AsMockStatsClient returns the mock as an explicitly typed haproxy.StatsClient interface
func (m *MockStatsClient) AsMockStatsClient() interface{} {
	return m
}
