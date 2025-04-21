package stats

// StatsItem represents a single entry in HAProxy stats
type StatsItem struct {
	PxName string `json:"pxname"`
	SvName string `json:"svname"`
	Type   int    `json:"type"`
	Status string `json:"status"`
	Weight int    `json:"weight"`
	// Add other fields as needed
}

// HAProxyStats represents the complete stats response from HAProxy
type HAProxyStats struct {
	Stats []StatsItem `json:"stats"`
}

// StatsSchema represents the schema of HAProxy stats
type StatsSchema struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties"`
}

// Property represents a property in the stats schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
