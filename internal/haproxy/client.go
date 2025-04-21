package haproxy

import (
	"context"
	"fmt"
	"log/slog"

	clientnative "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	configoptions "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/options"
	"github.com/haproxytech/client-native/v6/runtime"
	runtimeoptions "github.com/haproxytech/client-native/v6/runtime/options"

	"github.com/tuannvm/haproxy-mcp-server/internal/config"
)

// HAProxyClient wraps the HAProxy native client.
type HAProxyClient struct {
	Client clientnative.HAProxyClient
	Config *config.Config
}

// NewHAProxyClient creates and initializes a new HAProxy client wrapper.
func NewHAProxyClient(cfg *config.Config) (*HAProxyClient, error) {
	slog.Info("Initializing HAProxy client...")

	// Create context for initialization
	ctx := context.Background()

	// Initialize configuration client
	configOpts := []configoptions.ConfigurationOption{
		configoptions.ConfigurationFile(cfg.HAProxyConfigFile),
	}

	if cfg.HAProxyBinaryPath != "" {
		configOpts = append(configOpts, configoptions.HAProxyBin(cfg.HAProxyBinaryPath))
	}

	if cfg.HAProxyTransactionDir != "" {
		configOpts = append(configOpts, configoptions.TransactionsDir(cfg.HAProxyTransactionDir))
	}

	configClient, err := configuration.New(ctx, configOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy configuration client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy configuration client: %w", err)
	}

	// Initialize runtime client
	runtimeOpts := []runtimeoptions.RuntimeOption{
		runtimeoptions.Socket(cfg.HAProxyRuntimeSocket),
	}

	runtimeClient, err := runtime.New(ctx, runtimeOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy runtime client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy runtime client: %w", err)
	}

	// Create top-level client with both components
	clientOpts := []options.Option{
		options.Configuration(configClient),
		options.Runtime(runtimeClient),
	}

	client, err := clientnative.New(ctx, clientOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy client: %w", err)
	}

	slog.Info("HAProxy client successfully created")
	return &HAProxyClient{
		Client: client,
		Config: cfg,
	}, nil
}

// GetBackends retrieves a list of backend names.
func (c *HAProxyClient) GetBackends() ([]string, error) {
	slog.Debug("HAProxyClient.GetBackends called")

	// This is a placeholder implementation
	// In a real implementation, this would use c.Client to access HAProxy
	return []string{"web-backend", "api-backend", "db-backend"}, nil
}

// GetBackendDetails gets detailed information about a specific backend.
func (c *HAProxyClient) GetBackendDetails(name string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetBackendDetails called", "backend", name)

	// This is a placeholder implementation
	details := map[string]interface{}{
		"name":       name,
		"mode":       "http",
		"balance":    "roundrobin",
		"http_check": true,
		"servers":    []string{"server1", "server2"},
	}

	return details, nil
}

// ListServers retrieves a list of servers for a specific backend.
func (c *HAProxyClient) ListServers(backend string) ([]string, error) {
	slog.Debug("HAProxyClient.ListServers called", "backend", backend)

	// This is a placeholder implementation
	switch backend {
	case "web-backend":
		return []string{"web-server1", "web-server2", "web-server3"}, nil
	case "api-backend":
		return []string{"api-server1", "api-server2"}, nil
	case "db-backend":
		return []string{"db-master", "db-slave"}, nil
	default:
		return []string{"server1", "server2"}, nil
	}
}

// GetServerDetails retrieves detailed information about a specific server.
func (c *HAProxyClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetServerDetails called", "backend", backend, "server", server)

	// This is a placeholder implementation
	details := map[string]interface{}{
		"name":              server,
		"address":           "192.168.1.10",
		"port":              8080,
		"weight":            100,
		"maxconn":           1000,
		"status":            "UP",
		"check_status":      "L7OK",
		"last_state_change": "5d12h35m46s",
	}

	return details, nil
}

// EnableServer enables a server in a specific backend.
func (c *HAProxyClient) EnableServer(backend, server string) error {
	slog.Info("Enabling server", "backend", backend, "server", server)

	// This is a placeholder implementation
	// In a real implementation, this would use c.Client.Runtime() to enable the server
	return nil
}

// DisableServer disables a server in a specific backend.
func (c *HAProxyClient) DisableServer(backend, server string) error {
	slog.Info("Disabling server", "backend", backend, "server", server)

	// This is a placeholder implementation
	// In a real implementation, this would use c.Client.Runtime() to disable the server
	return nil
}

// GetRuntimeInfo retrieves runtime information (like 'show info').
func (c *HAProxyClient) GetRuntimeInfo() (map[string]string, error) {
	slog.Debug("HAProxyClient.GetRuntimeInfo called")

	// This is a placeholder implementation
	info := map[string]string{
		"version":       "2.6.6-84e0957",
		"name":          "HAProxy",
		"uptime":        "1d2h3m12s",
		"process_num":   "1",
		"max_conn":      "10000",
		"cur_conn":      "347",
		"hard_max_conn": "20000",
		"pid":           "12345",
		"threads":       "4",
		"nbproc":        "1",
		"mode":          "daemon",
	}

	return info, nil
}

// ReloadHAProxy triggers a configuration reload.
func (c *HAProxyClient) ReloadHAProxy() error {
	slog.Info("Reloading HAProxy configuration")

	// This is a placeholder implementation
	// In a real implementation, this would use c.Client to reload HAProxy
	return nil
}

// GetStats retrieves runtime statistics.
func (c *HAProxyClient) GetStats() (map[string]interface{}, error) {
	slog.Debug("HAProxyClient.GetStats called")

	// This is a placeholder implementation with realistic sample data
	stats := map[string]interface{}{
		"web-backend": map[string]interface{}{
			"type":             "backend",
			"status":           "UP",
			"current_sessions": 42,
			"max_sessions":     100,
			"sessions_limit":   2000,
			"bytes_in":         12345678,
			"bytes_out":        87654321,
			"servers": map[string]interface{}{
				"web-server1": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 15,
					"check_status":     "L7OK",
				},
				"web-server2": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 27,
					"check_status":     "L7OK",
				},
			},
		},
		"api-backend": map[string]interface{}{
			"type":             "backend",
			"status":           "UP",
			"current_sessions": 18,
			"max_sessions":     50,
			"sessions_limit":   1000,
			"bytes_in":         2345678,
			"bytes_out":        7654321,
			"servers": map[string]interface{}{
				"api-server1": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 8,
					"check_status":     "L7OK",
				},
				"api-server2": map[string]interface{}{
					"status":           "UP",
					"weight":           100,
					"current_sessions": 10,
					"check_status":     "L7OK",
				},
			},
		},
		"web-frontend": map[string]interface{}{
			"type":             "frontend",
			"status":           "OPEN",
			"current_sessions": 60,
			"max_sessions":     150,
			"sessions_limit":   3000,
			"bytes_in":         14345678,
			"bytes_out":        97654321,
		},
	}

	return stats, nil
}
