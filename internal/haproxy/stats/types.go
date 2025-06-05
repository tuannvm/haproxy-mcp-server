package stats

import (
	"fmt"
	"strconv"
)

// StatProperties defines a flexible map structure for storing HAProxy stat properties
type StatProperties map[string]interface{}

// StatsItem represents a single entry in HAProxy stats with a flexible property structure
type StatsItem struct {
	Properties StatProperties
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

// NewStatsItem creates a new StatsItem from a generic map
func NewStatsItem(data map[string]interface{}) StatsItem {
	return StatsItem{
		Properties: StatProperties(data),
	}
}

// Get returns a value from the stats item properties
func (s StatsItem) Get(key string) (interface{}, bool) {
	val, ok := s.Properties[key]
	return val, ok
}

// GetString returns a string value from the stats item properties
func (s StatsItem) GetString(key string) (string, bool) {
	// Try primary key
	if val, ok := s.Properties[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal, true
		}
	}
	return "", false
}

// GetInt returns an int value from the stats item properties
func (s StatsItem) GetInt(key string) (int, bool) {
	// Try direct integer value
	if val, ok := s.Properties[key]; ok {
		switch v := val.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal, true
			}
		}
	}

	// Try string variant with suffix
	strKey := fmt.Sprintf("%s,string", key)
	if val, ok := s.Properties[strKey]; ok {
		if strVal, ok := val.(string); ok {
			if intVal, err := strconv.Atoi(strVal); err == nil {
				return intVal, true
			}
		}
	}

	return 0, false
}

// GetInt64 returns an int64 value from the stats item properties
func (s StatsItem) GetInt64(key string) (int64, bool) {
	// Try direct integer value
	if val, ok := s.Properties[key]; ok {
		switch v := val.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		case string:
			if int64Val, err := strconv.ParseInt(v, 10, 64); err == nil {
				return int64Val, true
			}
		}
	}

	// Try string variant with suffix
	strKey := fmt.Sprintf("%s,string", key)
	if val, ok := s.Properties[strKey]; ok {
		if strVal, ok := val.(string); ok {
			if int64Val, err := strconv.ParseInt(strVal, 10, 64); err == nil {
				return int64Val, true
			}
		}
	}

	return 0, false
}

// GetProxyName returns the proxy name from the stats item
func (s StatsItem) GetProxyName() string {
	if val, ok := s.GetString("pxname"); ok {
		return val
	}
	if val, ok := s.GetString("name"); ok {
		return val
	}
	return ""
}

// GetServiceName returns the service name from the stats item
func (s StatsItem) GetServiceName() string {
	if val, ok := s.GetString("svname"); ok {
		return val
	}
	if val, ok := s.GetString("server"); ok {
		return val
	}
	return ""
}

// GetType returns the type value from the stats item
func (s StatsItem) GetType() int {
	if val, ok := s.GetInt("type"); ok {
		return val
	}
	return 0
}

// GetStatus returns the status value from the stats item
func (s StatsItem) GetStatus() string {
	if val, ok := s.GetString("status"); ok {
		return val
	}
	return ""
}

// GetWeight returns the weight value from the stats item
func (s StatsItem) GetWeight() int {
	if val, ok := s.GetInt("weight"); ok {
		return val
	}
	return 0
}
