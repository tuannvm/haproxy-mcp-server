// Package haproxy provides a client for interacting with HAProxy's Runtime API.
package haproxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// NewHAProxyClient creates a new HAProxy client
func NewHAProxyClient(runtimeAPIURL string) (*HAProxyClient, error) {
	// Parse URL to determine connection type
	u, err := url.Parse(runtimeAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime API URL: %w", err)
	}

	// Validate URL scheme
	switch u.Scheme {
	case "unix":
		slog.Debug("Initializing client for Unix socket connection", "path", u.Path)
	case "tcp":
		slog.Debug("Initializing client for TCP connection", "host", u.Host)
	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	client := &HAProxyClient{
		RuntimeAPIURL: runtimeAPIURL,
		ParsedURL:     u,
		Mode:          ClientModeDirect,
	}

	// Test direct connection by executing a simple command
	_, err = client.ExecuteRuntimeCommand("show info")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HAProxy Runtime API: %w", err)
	}

	slog.Info("Successfully connected to HAProxy Runtime API", "url", runtimeAPIURL, "mode", client.Mode)

	return client, nil
}

// executeSocketCommand is a shared helper function that handles command execution via sockets
// with support for context cancellation and timeouts
func (c *HAProxyClient) executeSocketCommand(ctx context.Context, network string, address string, command string) (string, error) {
	slog.Debug("Executing socket command", "network", network, "address", address, "command", command)

	// Check if context is already canceled
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Try socket connection with timeout from context
	var d net.Dialer
	connCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)

	go func() {
		conn, err := d.DialContext(ctx, network, address)
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	// Wait for connection or context cancellation
	var conn net.Conn
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		// Connection failed, try using socat as fallback
		slog.Debug("Direct socket connection failed, trying socat instead",
			"network", network, "error", err)

		if network == "tcp" {
			return c.executeSocatTCPCommand(command)
		} else {
			return c.executeSocatUnixCommand(command)
		}
	case conn = <-connCh:
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				slog.Error("Error closing socket connection",
					"network", network, "error", closeErr)
			}
		}()
	}

	// Use context deadline if available
	deadline, ok := ctx.Deadline()
	if ok {
		err := conn.SetDeadline(deadline)
		if err != nil {
			slog.Error("Failed to set deadline on socket connection",
				"network", network, "error", err)
			return "", fmt.Errorf("failed to set deadline: %w", err)
		}
	} else {
		err := conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			slog.Error("Failed to set deadline on socket connection",
				"network", network, "error", err)
			return "", fmt.Errorf("failed to set deadline: %w", err)
		}
	}

	// Send command
	slog.Debug("Sending command over socket", "command", command)
	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		slog.Error("Failed to send command over socket", "error", err)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response using dynamic buffer
	var buffer bytes.Buffer
	buf := make([]byte, 4096)
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Continue with read
		}

		// Set a short read deadline to allow for context cancellation
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			return "", fmt.Errorf("failed to set read deadline: %w", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// If we timeout on read, check context and try again
				if ctx.Err() != nil {
					return "", ctx.Err()
				}
				continue
			}
			slog.Error("Failed to read response from socket", "error", err)
			return "", fmt.Errorf("failed to read response: %w", err)
		}
		if n == 0 {
			break
		}
		buffer.Write(buf[:n])

		// Check if we've read all available data
		// This is a heuristic - we assume done if we read less than buffer size
		if n < len(buf) {
			break
		}
	}

	response := buffer.String()
	slog.Debug("Received response from socket", "network", network, "response_length", len(response))
	return response, nil
}

// executeDirectUnixCommandWithContext executes a command directly via Unix socket with context
func (c *HAProxyClient) executeDirectUnixCommandWithContext(ctx context.Context, command string) (string, error) {
	slog.Debug("Executing direct Unix command with context", "socket", c.ParsedURL.Path, "command", command)
	return c.executeSocketCommand(ctx, "unix", c.ParsedURL.Path, command)
}

// executeDirectTCPCommandWithContext executes a command directly via TCP with context
func (c *HAProxyClient) executeDirectTCPCommandWithContext(ctx context.Context, command string) (string, error) {
	slog.Debug("Executing direct TCP command with context", "host", c.ParsedURL.Host, "command", command)
	return c.executeSocketCommand(ctx, "tcp", c.ParsedURL.Host, command)
}

// executeSocatCommand executes a command via socat over the given target
func (c *HAProxyClient) executeSocatCommand(target, command string) (string, error) {
	slog.Debug("Executing socat command", "target", target, "command", command)

	// Check if socat is available
	socatPath, err := exec.LookPath("socat")
	if err != nil {
		slog.Error("Socat not found in system PATH", "error", err)
		return "", fmt.Errorf("socat not found: %w", err)
	}
	slog.Debug("Found socat binary", "path", socatPath)

	// Execute command with socat
	cmd := exec.Command(socatPath, target, "stdio")
	cmd.Stdin = bytes.NewBufferString(command + "\n")

	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	slog.Debug("Running socat command", "full_command", fmt.Sprintf("%s %s stdio", socatPath, target))
	err = cmd.Run()
	if err != nil {
		stderr := errOut.String()
		slog.Error("Socat command failed", "error", err, "stderr", stderr)
		return "", fmt.Errorf("socat command failed: %w, stderr: %s", err, stderr)
	}

	response := out.String()
	slog.Debug("Socat command successful", "response_length", len(response))
	return response, nil
}

// executeSocatTCPCommand executes a command via socat over TCP
func (c *HAProxyClient) executeSocatTCPCommand(command string) (string, error) {
	target := fmt.Sprintf("tcp-connect:%s", c.ParsedURL.Host)
	return c.executeSocatCommand(target, command)
}

// executeSocatUnixCommand executes a command via socat over Unix socket
func (c *HAProxyClient) executeSocatUnixCommand(command string) (string, error) {
	target := fmt.Sprintf("unix-connect:%s", c.ParsedURL.Path)
	return c.executeSocatCommand(target, command)
}

// executeDirectCommandWithContext executes a command directly via TCP or Unix socket with context
func (c *HAProxyClient) executeDirectCommandWithContext(ctx context.Context, command string) (string, error) {
	if c.ParsedURL.Scheme == "tcp" {
		return c.executeDirectTCPCommandWithContext(ctx, command)
	} else {
		return c.executeDirectUnixCommandWithContext(ctx, command)
	}
}

// executeWithErrorHandling is a helper method that executes a command with context
// and handles common error patterns and logging
func (c *HAProxyClient) executeWithErrorHandling(ctx context.Context, command, opDescription string) (string, error) {
	slog.Debug(fmt.Sprintf("Executing %s command", opDescription), "command", command)

	result, err := c.ExecuteRuntimeCommandWithContext(ctx, command)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to %s", opDescription), "command", command, "error", err)
		return "", fmt.Errorf("failed to %s: %w", opDescription, err)
	}

	slog.Debug(fmt.Sprintf("Successfully executed %s command", opDescription), "command", command)
	return result, nil
}

// ExecuteRuntimeCommand executes a command on HAProxy's Runtime API and processes the response.
func (c *HAProxyClient) ExecuteRuntimeCommand(command string) (string, error) {
	// Use default background context
	return c.ExecuteRuntimeCommandWithContext(context.Background(), command)
}

// ExecuteRuntimeCommandWithContext executes a command on HAProxy's Runtime API with context and processes the response.
func (c *HAProxyClient) ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error) {
	slog.Debug("Executing runtime command with context", "command", command)

	// Only use direct connection
	result, err := c.executeDirectCommandWithContext(ctx, command)
	if err != nil {
		slog.Error("Failed to execute runtime command", "command", command, "error", err)
		return "", fmt.Errorf("failed to execute runtime command: %w", err)
	}

	// Process response to handle error codes returned by HAProxy
	if len(result) > 4 {
		if result[0] == '[' && result[2] == ']' && result[3] == ':' {
			code := int(result[1] - '0')
			message := strings.TrimSpace(result[4:])
			slog.Debug("HAProxy returned error code", "code", code, "message", message)
			return "", NewHAProxyError(code, message, command)
		}
	}

	slog.Debug("Successfully executed runtime command", "command", command)
	return result, nil
}

// GetProcessInfo retrieves information about the HAProxy process.
func (c *HAProxyClient) GetProcessInfo() (map[string]string, error) {
	return c.GetProcessInfoWithContext(context.Background())
}

// GetProcessInfoWithContext retrieves information about the HAProxy process with context support.
func (c *HAProxyClient) GetProcessInfoWithContext(ctx context.Context) (map[string]string, error) {
	slog.Debug("Getting HAProxy process info")

	// Execute the 'show info' command with context
	result, err := c.executeWithErrorHandling(ctx, "show info", "get process info")
	if err != nil {
		return nil, err // Error already logged by executeWithErrorHandling
	}

	// Parse the result into a map
	infoMap := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Split each line by colon to get key-value pairs
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			infoMap[key] = value
		}
	}

	// Add HAProxy name (not directly available from API)
	infoMap["name"] = "HAProxy"

	slog.Debug("Successfully retrieved HAProxy process info")
	return infoMap, nil
}

// Close closes the HAProxy client connection.
func (c *HAProxyClient) Close() error {
	slog.Debug("Closing HAProxy client")
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

// executeDirectCommand executes a command directly via TCP or Unix socket
// Deprecated: Use executeDirectCommandWithContext instead
// Kept for backward compatibility
// nolint:unused
func (c *HAProxyClient) executeDirectCommand(command string) (string, error) {
	return c.executeDirectCommandWithContext(context.Background(), command)
}

// executeDirectTCPCommand executes a command directly via TCP
// Deprecated: Use executeDirectTCPCommandWithContext instead
// Kept for backward compatibility
// nolint:unused
func (c *HAProxyClient) executeDirectTCPCommand(command string) (string, error) {
	// Use default background context
	ctx := context.Background()
	return c.executeDirectTCPCommandWithContext(ctx, command)
}

// executeDirectUnixCommand executes a command directly via Unix socket
// Deprecated: Use executeDirectUnixCommandWithContext instead
// Kept for backward compatibility
// nolint:unused
func (c *HAProxyClient) executeDirectUnixCommand(command string) (string, error) {
	slog.Debug("Executing direct Unix command", "socket", c.ParsedURL.Path, "command", command)

	// Use default background context
	ctx := context.Background()
	return c.executeDirectUnixCommandWithContext(ctx, command)
}
