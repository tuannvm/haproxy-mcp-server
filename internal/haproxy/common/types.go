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
