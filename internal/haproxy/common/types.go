package common

// StatItem represents a single HAProxy statistics item
// This is a common type used across packages
type StatItem struct {
	ProxyName   string `json:"pxname"`
	ServiceName string `json:"svname"`
	Type        int    `json:"type"`
	Status      string `json:"status"`
	Weight      int    `json:"weight"`
	// Add other fields as needed
}

// Stats represents a subset of the HAProxy stats data relevant to our needs.
// This is a local helper type to make working with the stats data easier.
type Stats struct {
	// Proxy name
	Pxname string
	// Service name (FRONTEND for frontend, BACKEND for backend, or server name)
	Svname string
	// Type (1=frontend, 2=backend, 3=server)
	Type string
	// Status (UP, DOWN, etc.)
	Status string
	// Mode (http, tcp)
	Mode string
	// Current sessions
	Scur string
	// Max sessions
	Smax string
	// Session limit
	Slim string
	// Bytes in
	Bin string
	// Bytes out
	Bout string
}
