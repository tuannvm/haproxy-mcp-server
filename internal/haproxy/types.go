package haproxy

import (
	"net/url"
)

// Server states
const (
	ServerStateReady       = "ready"
	ServerStateMaint       = "maint"
	ServerStateDrain       = "drain"
	ServerStateSoftStop    = "softstop"
	ServerStateStop        = "stop"
	ServerCheckStateEnable = "enable"
	ServerCheckStateDisabe = "disable"

	// Backend types
	BackendTypeHTTP = "http"
	BackendTypeTCP  = "tcp"

	// Common status values
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

// HAProxyClientMode defines how the client connects to HAProxy
type HAProxyClientMode int

const (
	// ClientModeNative is kept for API compatibility but is no longer supported
	ClientModeNative HAProxyClientMode = iota
	// ClientModeDirect uses direct socket connection (TCP or Unix)
	ClientModeDirect
)

// HAProxyClient provides methods for interacting with HAProxy's Runtime API.
type HAProxyClient struct {
	RuntimeAPIURL    string
	ConfigurationURL string
	ParsedURL        *url.URL
	Mode             HAProxyClientMode
}

// BackendInfo represents detailed information about a backend.
type BackendInfo struct {
	Name     string            `json:"name"`     // Name of the backend
	Status   string            `json:"status"`   // Current status (UP, DOWN, etc.)
	Sessions int               `json:"sessions"` // Current active sessions
	Servers  []ServerInfo      `json:"servers"`  // Servers in this backend
	Stats    map[string]string `json:"stats"`    // Additional statistics
}

// ServerInfo represents detailed information about a server.
type ServerInfo struct {
	Name              string `json:"name"`               // Name of the server
	Address           string `json:"address"`            // IP address of the server
	Port              string `json:"port"`               // Port of the server
	Status            string `json:"status"`             // Current status (UP, DOWN, MAINT, etc.)
	Weight            int    `json:"weight"`             // Current weight
	CheckStatus       string `json:"check_status"`       // Status of the last health check
	LastStatusChange  string `json:"last_status_change"` // Time since last status change
	TotalConnections  int    `json:"total_connections"`  // Total connections
	ActiveConnections int    `json:"active_connections"` // Current active connections
}

// CommandOptions provides options for executing HAProxy commands
type CommandOptions struct {
	Timeout int  // Timeout in seconds
	Verbose bool // Enable verbose output
}

// ExecuteContext provides context for command execution
type ExecuteContext struct {
	CommandName string
	Backend     string
	Server      string
	Options     *CommandOptions
}
