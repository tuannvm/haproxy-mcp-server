// Package haproxy provides a client for interacting with HAProxy's Runtime API.
package haproxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/haproxytech/client-native/v6/runtime"
	runtimeoptions "github.com/haproxytech/client-native/v6/runtime/options"
)

// NewHAProxyClient creates a new HAProxyClient using the provided Runtime API endpoint.
func NewHAProxyClient(runtimeAPIURL, configurationURL string) (*HAProxyClient, error) {
	slog.Debug("Creating new HAProxy client", "runtimeAPIURL", runtimeAPIURL, "configurationURL", configurationURL)

	// Parse the Runtime API URL
	parsedRuntimeURL, err := url.Parse(runtimeAPIURL)
	if err != nil {
		slog.Error("Failed to parse runtime API URL", "url", runtimeAPIURL, "error", err)
		return nil, fmt.Errorf("failed to parse runtime API URL: %w", err)
	}

	// Validate the Runtime API URL
	if parsedRuntimeURL.Scheme != "unix" && parsedRuntimeURL.Scheme != "tcp" {
		slog.Error("Unsupported runtime API URL scheme", "scheme", parsedRuntimeURL.Scheme)
		return nil, fmt.Errorf("unsupported runtime API URL scheme: %s (must be unix or tcp)", parsedRuntimeURL.Scheme)
	}

	// Create context and set up socket option
	ctx := context.Background()
	socketOpt := runtimeoptions.Socket(parsedRuntimeURL.Path)

	// Create runtime client
	runtimeClient, err := runtime.New(ctx, socketOpt)
	if err != nil {
		slog.Error("Failed to create HAProxy runtime client", "error", err)
		return nil, fmt.Errorf("failed to create HAProxy runtime client: %w", err)
	}

	// Test the connection with a simple command
	_, err = runtimeClient.ExecuteRaw("help")
	if err != nil {
		slog.Error("Failed to connect to runtime API", "error", err)
		return nil, fmt.Errorf("failed to connect to runtime API: %w", err)
	}

	return &HAProxyClient{
		RuntimeAPIURL:    runtimeAPIURL,
		ConfigurationURL: configurationURL,
		client:           runtimeClient,
	}, nil
}

// Runtime returns the underlying native runtime client.
func (c *HAProxyClient) Runtime() runtime.Runtime {
	return c.client
}

// ExecuteRuntimeCommand executes a command on HAProxy's Runtime API and processes the response.
func (c *HAProxyClient) ExecuteRuntimeCommand(command string) (string, error) {
	slog.Debug("Executing runtime command", "command", command)

	// Execute command with raw response, then process it
	result, err := c.client.ExecuteRaw(command)
	if err != nil {
		slog.Error("Failed to execute runtime command", "command", command, "error", err)
		return "", fmt.Errorf("failed to execute runtime command: %w", err)
	}

	// Process response to handle error codes returned by HAProxy
	if len(result) > 4 {
		switch result[0:4] {
		case "[3]:", "[2]:", "[1]:", "[0]:":
			return "", fmt.Errorf("[%c] %s [%s]", result[1], result[4:], command)
		}
	}

	slog.Debug("Successfully executed runtime command", "command", command)
	return result, nil
}

// GetProcessInfo retrieves information about the HAProxy process.
func (c *HAProxyClient) GetProcessInfo() (map[string]string, error) {
	slog.Debug("Getting HAProxy process info")

	// Get info directly from the runtime client
	info, err := c.client.GetInfo()
	if err != nil {
		slog.Error("Failed to get process info", "error", err)
		return nil, fmt.Errorf("failed to get process info: %w", err)
	}

	// Convert the structured info to a simple map
	infoMap := make(map[string]string)

	// Access fields through the Info field
	if info.Info != nil {
		if info.Info.Version != "" {
			infoMap["version"] = info.Info.Version
		}

		if info.Info.Uptime != nil {
			infoMap["uptime"] = fmt.Sprintf("%d", *info.Info.Uptime)
		}

		if info.Info.ProcessNum != nil {
			infoMap["process_num"] = fmt.Sprintf("%d", *info.Info.ProcessNum)
		}

		if info.Info.Pid != nil {
			infoMap["pid"] = fmt.Sprintf("%d", *info.Info.Pid)
		}
	}

	slog.Debug("Successfully retrieved HAProxy process info")
	return infoMap, nil
}

// Close closes the HAProxy client connection.
func (c *HAProxyClient) Close() error {
	slog.Debug("Closing HAProxy client")
	// No explicit close method in the upstream client, but adding this for future-proofing
	return nil
}

// GetHaproxyAPIEndpoint returns the URL for the HAProxy API from socket path.
// This is a utility function for clients that need the API URL.
func GetHaproxyAPIEndpoint(socketPath string) (string, error) {
	slog.Debug("Getting HAProxy API endpoint", "socketPath", socketPath)

	// Validate socket path
	if socketPath == "" {
		return "", fmt.Errorf("HAProxy socket path is empty")
	}

	// Create a URL with unix socket protocol using the socket path
	u := &url.URL{
		Scheme: "unix",
		Path:   socketPath,
	}

	// Create the API URL
	apiURL := fmt.Sprintf("%s/v2", u)
	slog.Debug("HAProxy API endpoint", "url", apiURL)

	return apiURL, nil
}
