package stats

// StatsClient is a client for fetching HAProxy stats from the stats page
type StatsClient struct {
	StatsURL string // URL to HAProxy stats page (e.g., http://127.0.0.1:1936/;json)
}

// HAProxyStats represents the root structure of HAProxy stats JSON response
type HAProxyStats struct {
	Stats []StatsItem `json:"stats"`
}

// StatsItem represents a single item in the stats array
type StatsItem struct {
	// Frontend/backend identification
	PxName string `json:"pxname"` // Proxy name
	SvName string `json:"svname"` // Service name
	Type   int    `json:"type"`   // Type (0=frontend, 1=backend, 2=server, 3=socket/listener)

	// Status and health
	Status      string `json:"status"`       // Status (UP, DOWN, MAINT, etc.)
	CheckStatus string `json:"check_status"` // Health check status
	LastChg     int    `json:"lastchg"`      // Seconds since last change
	Weight      int    `json:"weight"`       // Server weight
	Slim        int    `json:"slim"`         // Configured session limit
	Qlimit      int    `json:"qlimit"`       // Queue limit

	// Activity metrics
	Scur    int   `json:"scur"`     // Current sessions
	Smax    int   `json:"smax"`     // Max sessions
	Stot    int64 `json:"stot"`     // Total sessions
	Rate    int   `json:"rate"`     // Current rate (sessions per second)
	RateMax int   `json:"rate_max"` // Max rate
	RateLim int   `json:"rate_lim"` // Rate limit

	// Connection stats
	ConnRate    int   `json:"conn_rate"`     // Connection rate
	ConnRateMax int   `json:"conn_rate_max"` // Max connection rate
	ConnTot     int64 `json:"conn_tot"`      // Total connections

	// Bandwidth
	BinRate  int64 `json:"bin_rate"`  // Bytes in rate (bytes/s)
	BoutRate int64 `json:"bout_rate"` // Bytes out rate (bytes/s)
	Bin      int64 `json:"bin"`       // Bytes in
	Bout     int64 `json:"bout"`      // Bytes out

	// Request/response stats
	ReqRate   int   `json:"req_rate"`   // HTTP request rate
	ReqTot    int64 `json:"req_tot"`    // Total HTTP requests
	Hrsp1xx   int64 `json:"hrsp_1xx"`   // HTTP 1xx responses
	Hrsp2xx   int64 `json:"hrsp_2xx"`   // HTTP 2xx responses
	Hrsp3xx   int64 `json:"hrsp_3xx"`   // HTTP 3xx responses
	Hrsp4xx   int64 `json:"hrsp_4xx"`   // HTTP 4xx responses
	Hrsp5xx   int64 `json:"hrsp_5xx"`   // HTTP 5xx responses
	HrspOther int64 `json:"hrsp_other"` // Other HTTP responses

	// Errors
	EreqRate  int   `json:"ereq_rate"` // Request error rate
	EreqTot   int64 `json:"ereq_tot"`  // Total request errors
	EconnRate int   `json:"econ_rate"` // Connection error rate
	EconnTot  int64 `json:"econ_tot"`  // Total connection errors
	EretrTot  int64 `json:"eretr_tot"` // Total retry attempts

	// Queue stats
	Qcur int `json:"qcur"` // Current queued requests
	Qmax int `json:"qmax"` // Max queued requests

	// Timing
	Ctime int `json:"ctime"` // Connect time in ms
	Rtime int `json:"rtime"` // Response time in ms
	Ttime int `json:"ttime"` // Total time in ms

	// Server details (for server entries)
	CheckCode     int   `json:"check_code"`     // Check code (HTTP status)
	CheckDuration int   `json:"check_duration"` // Check duration in ms
	Hanafail      int   `json:"hanafail"`       // Failed health checks
	CliAbrt       int64 `json:"cli_abrt"`       // Client connection aborts
	SrvAbrt       int64 `json:"srv_abrt"`       // Server connection aborts

	// Additional fields
	Mode string `json:"mode"` // Proxy mode (tcp or http)
	Algo string `json:"algo"` // Load balancing algorithm
}

// StatsSchema represents the JSON schema structure for HAProxy stats
type StatsSchema struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Type        string              `json:"type"`
	Properties  map[string]Property `json:"properties"`
}

// Property represents a schema property
type Property struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
