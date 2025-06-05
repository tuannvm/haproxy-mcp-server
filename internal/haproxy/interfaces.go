package haproxy

import (
	"context"

	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/common"
	runtimeclient "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/runtime"
	"github.com/tuannvm/haproxy-mcp-server/internal/haproxy/stats"
)

// CommonClient defines methods supported by both Runtime and Stats APIs
// This interface represents the common functionality available regardless of client mode
type CommonClient interface {
	// Backend operations
	GetBackends() ([]string, error)
	GetBackendDetails(name string) (map[string]interface{}, error)

	// Server operations
	ListServers(backend string) ([]string, error)
	GetServerDetails(backend, server string) (map[string]interface{}, error)

	// Stats operations
	ShowStat(filter string) ([]map[string]string, error)
	ShowServersState(backend string) ([]map[string]string, error)

	// Process info operations (supported by both APIs)
	GetRuntimeInfo() (map[string]string, error)
}

// RuntimeOnlyClient defines methods that are only available when Runtime API is enabled
type RuntimeOnlyClient interface {
	// Runtime API core operations
	ExecuteRuntimeCommand(command string) (string, error)
	ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error)

	// Server manipulation operations
	EnableServer(backend, server string) error
	DisableServer(backend, server string) error
	SetWeight(backend, server string, weight int) (string, error)
	SetServerMaxconn(backend, server string, maxconn int) error
	EnableHealth(backend, server string) error
	DisableHealth(backend, server string) error
	EnableAgent(backend, server string) error
	DisableAgent(backend, server string) error

	// Advanced operations
	DumpStatsFile(filepath string) (string, error)
	DebugCounters() (map[string]interface{}, error)
	ClearCountersAll() error
	AddServer(backend, name, addr string, port, weight int) error
	DelServer(backend, name string) error
	ReloadHAProxy() error
}

// StatsOnlyClient defines methods that are only available when Stats API is enabled
type StatsOnlyClient interface {
	// Stats retrieval operations
	GetStats() (*stats.HAProxyStats, error)
}

// HAProxyClientInterface is the primary client that implements all interfaces
// It can operate in three modes:
// 1. Full mode (RuntimeClient + StatsClient) - implements CommonClient, RuntimeOnlyClient, and StatsOnlyClient
// 2. Runtime-only mode - implements CommonClient and RuntimeOnlyClient
// 3. Stats-only mode - implements CommonClient and StatsOnlyClient
type HAProxyClientInterface interface {
	CommonClient
	RuntimeOnlyClient
	StatsOnlyClient

	// Client mode operations
	GetClientMode() ClientMode
	IsStatsOnlyMode() bool
	IsRuntimeOnlyMode() bool
	IsFullMode() bool

	// Cleanup operations
	Close() error
}

// RuntimeClient defines the interface for the underlying Runtime API client
type RuntimeClient interface {
	// Runtime API operations
	ExecuteRuntimeCommand(command string) (string, error)
	ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error)
	GetProcessInfo() (map[string]string, error)
	GetProcessInfoWithContext(ctx context.Context) (map[string]string, error)
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

// StatsClient defines the interface for the underlying Stats API client
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
