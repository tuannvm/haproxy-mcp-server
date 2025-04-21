package haproxy

import (
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
	runtimeclient "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/runtime"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// RuntimeClient defines the interface for interacting with HAProxy's Runtime API
type RuntimeClient interface {
	// Runtime API operations
	ExecuteRuntimeCommand(command string) (string, error)
	GetProcessInfo() (map[string]string, error)
	Close() error

	// Backend operations
	ListBackends() ([]string, error)
	GetBackendInfo(name string) (*runtimeclient.BackendInfo, error)
	EnableBackend(name string) error
	DisableBackend(name string) error

	// Server operations
	ListServers(backend string) ([]string, error)
	GetServerDetails(backend, server string) (map[string]interface{}, error)
	EnableServer(backend, server string) error
	DisableServer(backend, server string) error
	SetServerWeight(backend, server string, weight int) error
	SetServerMaxconn(backend, server string, maxconn int) error
	GetServerState(backend, server string) (string, error)
}

// StatsClient defines the interface for interacting with HAProxy's Stats API
type StatsClient interface {
	// Stats API operations
	GetStats() (*stats.HAProxyStats, error)
	GetSchema() (*stats.StatsSchema, error)

	// Data filtering operations
	FilterStats(stats *stats.HAProxyStats, proxyName, serviceName string) []common.StatItem
	GetFrontends(stats *stats.HAProxyStats) []common.StatItem
	GetBackends(stats *stats.HAProxyStats) []common.StatItem
	GetServers(stats *stats.HAProxyStats) []common.StatItem
	GetServersByBackend(stats *stats.HAProxyStats, backendName string) []common.StatItem
}
