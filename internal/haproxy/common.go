package haproxy

// ServerInfo contains information about a HAProxy server.
type ServerInfo struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	Port              int    `json:"port,omitempty"`
	Status            string `json:"status,omitempty"`
	Weight            int    `json:"weight,omitempty"`
	CheckStatus       string `json:"check_status,omitempty"`
	LastStatusChange  string `json:"last_status_change,omitempty"`
	TotalConnections  int    `json:"total_connections,omitempty"`
	ActiveConnections int    `json:"active_connections,omitempty"`
}

// BackendInfo contains information about a HAProxy backend.
type BackendInfo struct {
	Name     string            `json:"name"`
	Status   string            `json:"status"`
	Servers  []ServerInfo      `json:"servers,omitempty"`
	Sessions int               `json:"sessions"`
	Stats    map[string]string `json:"stats,omitempty"`
}

// Constants for HAProxy object types
const (
	// Server states
	ServerStateReady = "ready"
	ServerStateMaint = "maint"
	ServerStateDrain = "drain"

	// Backend types
	BackendTypeHTTP = "http"
	BackendTypeTCP  = "tcp"

	// Common status values
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

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
