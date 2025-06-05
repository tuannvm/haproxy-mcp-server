package testing

import (
	"fmt"

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
	statspkg "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// MockStatsClient implements the haproxy.StatsClient interface for testing
type MockStatsClient struct {
	haproxy.StatsClient // Embedded interface for proper implementation

	// Configuration for mocking behavior
	FailGetStats  bool
	FailGetSchema bool

	// Mocked return values
	Stats  *statspkg.HAProxyStats
	Schema *statspkg.StatsSchema
}

// NewMockStatsClient creates a new mock stats client with default settings
func NewMockStatsClient() *MockStatsClient {
	return &MockStatsClient{
		Stats: &statspkg.HAProxyStats{
			Stats: []statspkg.StatsItem{
				statspkg.NewStatsItem(map[string]interface{}{
					"pxname": "backend1",
					"svname": "BACKEND",
					"type":   1, // Backend type
					"status": "UP",
					"weight": 100,
				}),
				statspkg.NewStatsItem(map[string]interface{}{
					"pxname": "backend1",
					"svname": "server1",
					"type":   2, // Server type
					"status": "UP",
					"weight": 100,
				}),
			},
		},
		Schema: &statspkg.StatsSchema{
			Title:       "HAProxy Stats Schema",
			Description: "Schema for HAProxy statistics",
			Type:        "object",
			Properties:  map[string]statspkg.Property{},
		},
	}
}

// GetStats implements StatsClient.GetStats
func (m *MockStatsClient) GetStats() (*statspkg.HAProxyStats, error) {
	if m.FailGetStats {
		return nil, fmt.Errorf("mock error getting stats")
	}
	return m.Stats, nil
}

// GetSchema implements StatsClient.GetSchema
func (m *MockStatsClient) GetSchema() (*statspkg.StatsSchema, error) {
	if m.FailGetSchema {
		return nil, fmt.Errorf("mock error getting schema")
	}
	return m.Schema, nil
}

// Helper function to filter stats items and convert them to common.StatItem
func filterStatsItems(items []statspkg.StatsItem, filter func(item statspkg.StatsItem) bool) []common.StatItem {
	var result []common.StatItem

	for _, item := range items {
		if filter(item) {
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

// FilterStats implements StatsClient.FilterStats
func (m *MockStatsClient) FilterStats(stats *statspkg.HAProxyStats, proxyName, serviceName string) []common.StatItem {
	return filterStatsItems(stats.Stats, func(item statspkg.StatsItem) bool {
		return (proxyName == "" || item.GetProxyName() == proxyName) &&
			(serviceName == "" || item.GetServiceName() == serviceName)
	})
}

// GetFrontends implements StatsClient.GetFrontends
func (m *MockStatsClient) GetFrontends(stats *statspkg.HAProxyStats) []common.StatItem {
	return filterStatsItems(stats.Stats, func(item statspkg.StatsItem) bool {
		return item.GetType() == 0 // Type 0 is frontend
	})
}

// GetBackends implements StatsClient.GetBackends
func (m *MockStatsClient) GetBackends(stats *statspkg.HAProxyStats) []common.StatItem {
	return filterStatsItems(stats.Stats, func(item statspkg.StatsItem) bool {
		return item.GetType() == 1 // Type 1 is backend
	})
}

// GetServers implements StatsClient.GetServers
func (m *MockStatsClient) GetServers(stats *statspkg.HAProxyStats) []common.StatItem {
	return filterStatsItems(stats.Stats, func(item statspkg.StatsItem) bool {
		return item.GetType() == 2 // Type 2 is server
	})
}

// GetServersByBackend implements StatsClient.GetServersByBackend
func (m *MockStatsClient) GetServersByBackend(stats *statspkg.HAProxyStats, backendName string) []common.StatItem {
	return filterStatsItems(stats.Stats, func(item statspkg.StatsItem) bool {
		return item.GetType() == 2 && item.GetProxyName() == backendName // Type 2 is server
	})
}
