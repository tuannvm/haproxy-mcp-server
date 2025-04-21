package testing

import (
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// StatsClientAdapter adapts our mock to the haproxy.StatsClient interface
type StatsClientAdapter struct {
	mock *MockStatsClient
}

// NewStatsClientAdapter creates a new adapter for the mock stats client
func NewStatsClientAdapter(mock *MockStatsClient) haproxy.StatsClient {
	return &StatsClientAdapter{mock: mock}
}

// GetStats implements haproxy.StatsClient
func (a *StatsClientAdapter) GetStats() (*stats.HAProxyStats, error) {
	return a.mock.GetStats()
}

// GetSchema implements haproxy.StatsClient
func (a *StatsClientAdapter) GetSchema() (*stats.StatsSchema, error) {
	return a.mock.GetSchema()
}

// FilterStats implements haproxy.StatsClient
func (a *StatsClientAdapter) FilterStats(stats *stats.HAProxyStats, proxyName, serviceName string) []common.StatItem {
	return a.mock.FilterStats(stats, proxyName, serviceName)
}

// GetFrontends implements haproxy.StatsClient
func (a *StatsClientAdapter) GetFrontends(stats *stats.HAProxyStats) []common.StatItem {
	return a.mock.GetFrontends(stats)
}

// GetBackends implements haproxy.StatsClient
func (a *StatsClientAdapter) GetBackends(stats *stats.HAProxyStats) []common.StatItem {
	return a.mock.GetBackends(stats)
}

// GetServers implements haproxy.StatsClient
func (a *StatsClientAdapter) GetServers(stats *stats.HAProxyStats) []common.StatItem {
	return a.mock.GetServers(stats)
}

// GetServersByBackend implements haproxy.StatsClient
func (a *StatsClientAdapter) GetServersByBackend(stats *stats.HAProxyStats, backendName string) []common.StatItem {
	return a.mock.GetServersByBackend(stats, backendName)
}

// NewMockHAProxyClient creates a new HAProxy client with mock runtime and stats clients
func NewMockHAProxyClient() *haproxy.HAProxyClient {
	return &haproxy.HAProxyClient{
		RuntimeClient: NewMockRuntimeClient(),
		StatsClient:   NewStatsClientAdapter(NewMockStatsClient()),
		StatsURL:      "http://localhost:8404/stats",
	}
}
