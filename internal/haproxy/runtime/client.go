// Package haproxy provides a client for interacting with HAProxy's Runtime API.
package haproxy

import (
	"bytes"
	"fmt"
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

// executeDirectCommand executes a command directly via TCP or Unix socket
func (c *HAProxyClient) executeDirectCommand(command string) (string, error) {
	if c.ParsedURL.Scheme == "tcp" {
		return c.executeDirectTCPCommand(command)
	} else {
		return c.executeDirectUnixCommand(command)
	}
}

// executeDirectTCPCommand executes a command directly via TCP
func (c *HAProxyClient) executeDirectTCPCommand(command string) (string, error) {
	slog.Debug("Executing direct TCP command", "host", c.ParsedURL.Host, "command", command)

	// Try TCP connection
	slog.Debug("Attempting direct TCP connection to HAProxy", "address", c.ParsedURL.Host)
	conn, err := net.DialTimeout("tcp", c.ParsedURL.Host, 5*time.Second)
	if err != nil {
		slog.Debug("Direct TCP connection failed, trying socat instead", "error", err)
		// Try using socat if direct connection fails
		return c.executeSocatTCPCommand(command)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("Error closing TCP connection", "error", closeErr)
		}
	}()
	slog.Debug("Successfully established TCP connection to HAProxy")

	// Set deadlines
	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		slog.Error("Failed to set deadline on TCP connection", "error", err)
		return "", fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send command
	slog.Debug("Sending command to HAProxy", "command", command)
	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		slog.Error("Failed to send command to HAProxy", "error", err)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	buf := make([]byte, 4096)
	slog.Debug("Reading response from HAProxy")
	n, err := conn.Read(buf)
	if err != nil {
		slog.Error("Failed to read response from HAProxy", "error", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buf[:n])
	slog.Debug("Received response from HAProxy", "response_length", len(response))
	return response, nil
}

// executeDirectUnixCommand executes a command directly via Unix socket
func (c *HAProxyClient) executeDirectUnixCommand(command string) (string, error) {
	slog.Debug("Executing direct Unix command", "socket", c.ParsedURL.Path, "command", command)

	// Try Unix socket connection
	conn, err := net.DialTimeout("unix", c.ParsedURL.Path, 5*time.Second)
	if err != nil {
		// Try using socat if direct connection fails
		return c.executeSocatUnixCommand(command)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("Error closing Unix socket connection", "error", closeErr)
		}
	}()

	// Set deadlines
	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send command
	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(buf[:n]), nil
}

// executeSocatTCPCommand executes a command via socat over TCP
func (c *HAProxyClient) executeSocatTCPCommand(command string) (string, error) {
	slog.Debug("Executing socat TCP command", "host", c.ParsedURL.Host, "command", command)

	// Check if socat is available
	socatPath, err := exec.LookPath("socat")
	if err != nil {
		slog.Error("Socat not found in system PATH", "error", err)
		return "", fmt.Errorf("socat not found: %w", err)
	}
	slog.Debug("Found socat binary", "path", socatPath)

	// Prepare socat command args
	socatTarget := fmt.Sprintf("tcp-connect:%s", c.ParsedURL.Host)
	slog.Debug("Preparing socat command", "target", socatTarget)

	// Execute command with socat
	cmd := exec.Command(socatPath, socatTarget, "stdio")
	cmd.Stdin = bytes.NewBufferString(command + "\n")

	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	slog.Debug("Running socat command", "full_command", fmt.Sprintf("%s %s stdio", socatPath, socatTarget))
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

// executeSocatUnixCommand executes a command via socat over Unix socket
func (c *HAProxyClient) executeSocatUnixCommand(command string) (string, error) {
	slog.Debug("Executing socat Unix command", "socket", c.ParsedURL.Path, "command", command)

	// Check if socat is available
	socatPath, err := exec.LookPath("socat")
	if err != nil {
		return "", fmt.Errorf("socat not found: %w", err)
	}

	// Execute command with socat
	cmd := exec.Command(socatPath, fmt.Sprintf("unix-connect:%s", c.ParsedURL.Path), "stdio")
	cmd.Stdin = bytes.NewBufferString(command + "\n")

	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("socat command failed: %w, stderr: %s", err, errOut.String())
	}

	return out.String(), nil
}

// ExecuteRuntimeCommand executes a command on HAProxy's Runtime API and processes the response.
func (c *HAProxyClient) ExecuteRuntimeCommand(command string) (string, error) {
	slog.Debug("Executing runtime command", "command", command)

	// Only use direct connection
	result, err := c.executeDirectCommand(command)
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
// This is refactored to use direct command execution rather than client-native.
func (c *HAProxyClient) GetProcessInfo() (map[string]string, error) {
	slog.Debug("Getting HAProxy process info")

	// Execute the 'show info' command directly
	result, err := c.ExecuteRuntimeCommand("show info")
	if err != nil {
		slog.Error("Failed to get process info", "error", err)
		return nil, fmt.Errorf("failed to get process info: %w", err)
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
