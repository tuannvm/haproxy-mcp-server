// Package haproxy provides a client for interacting with HAProxy.
package haproxy

import (
	"context"
	"fmt"
	"log/slog"

	clientnative "github.com/haproxytech/client-native/v6"
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
// This client only uses the HAProxy Runtime API, not the Configuration API.
func NewHAProxyClient(cfg *config.Config) (*HAProxyClient, error) {
	slog.Info("Initializing HAProxy client (Runtime API only)...")

	// Create context for initialization
	ctx := context.Background()

	// Initialize runtime client
	runtimeOpts := []runtimeoptions.RuntimeOption{
		runtimeoptions.Socket(cfg.HAProxyRuntimeSocket),
	}

	runtimeClient, err := runtime.New(ctx, runtimeOpts...)
	if err != nil {
		slog.Error("Failed to initialize HAProxy runtime client", "error", err)
		return nil, fmt.Errorf("failed to initialize HAProxy runtime client: %w", err)
	}

	// Create top-level client with only the runtime component
	clientOpts := []options.Option{
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
