package common

import (
	"fmt"
	"strings"
)

// GetIntValue extracts an integer value from either string or int field
// This consolidates the common pattern of checking multiple fields for the same value
func GetIntValue(stringValue, intValue int) int {
	if stringValue != 0 {
		return stringValue
	}
	return intValue
}

// GetInt64Value extracts an int64 value from either string or int field
func GetInt64Value(stringValue, intValue int64) int64 {
	if stringValue != 0 {
		return stringValue
	}
	return intValue
}

// FormatAPIError formats an error message based on the API mode
// This helps create consistent error messages across the codebase
func FormatAPIError(err error, action string, isStatsOnlyMode bool) error {
	if isStatsOnlyMode {
		return fmt.Errorf("failed to %s: %w (running in stats-only mode)", action, err)
	}
	return fmt.Errorf("failed to %s: %w", action, err)
}

// FormatModeSpecificError creates a more descriptive error message based on the mode and error content
func FormatModeSpecificError(err error, action string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check if the error message already contains mode information
	if strings.Contains(errMsg, "running in stats-only mode") {
		return fmt.Errorf("failed to %s: %v", action, err)
	} else if strings.Contains(errMsg, "running in runtime-only mode") {
		return fmt.Errorf("failed to %s: %v", action, err)
	} else if strings.Contains(errMsg, "HAPROXY_RUNTIME_ENABLED=false") ||
		strings.Contains(errMsg, "stats-only mode") {
		return fmt.Errorf("failed to %s: %v (running in stats-only mode)", action, err)
	} else if strings.Contains(errMsg, "stats client failed") ||
		strings.Contains(errMsg, "stats client is not initialized") {
		return fmt.Errorf("failed to %s: %v (running in runtime-only mode)", action, err)
	}

	return fmt.Errorf("failed to %s: %v", action, err)
}

// MapValueToInt safely converts an interface value to int
func MapValueToInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		var result int
		if _, err := fmt.Sscanf(v, "%d", &result); err == nil {
			return result, true
		}
	}
	return 0, false
}

// MapValueToString safely converts an interface value to string
func MapValueToString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case int:
		return fmt.Sprintf("%d", v), true
	case float64:
		return fmt.Sprintf("%v", v), true
	case bool:
		return fmt.Sprintf("%v", v), true
	}
	return "", false
}

// ExtractIntValue extracts an integer value from a map with fallbacks
func ExtractIntValue(data map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if val, ok := MapValueToInt(data[key]); ok {
			return val
		}
	}
	return 0
}

// ExtractStringValue extracts a string value from a map with fallbacks
func ExtractStringValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := MapValueToString(data[key]); ok {
			return val
		}
	}
	return ""
}
