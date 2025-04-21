package haproxy

import (
	"fmt"
	"strings"
)

// parseCSVStats parses HAProxy stats output in CSV format
func parseCSVStats(statsOutput string) ([]string, []map[string]string, error) {
	lines := splitAndTrim(statsOutput)
	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("invalid stats output format: insufficient lines")
	}

	// Get headers from first line
	headers := strings.Split(lines[0], ",")

	// Process data lines
	results := make([]map[string]string, 0, len(lines)-1)

	for i := 1; i < len(lines); i++ {
		data := strings.Split(lines[i], ",")
		if len(data) < len(headers) {
			continue // Skip incomplete lines
		}

		// Create a map of field name to value
		fieldMap := make(map[string]string)
		for j := 0; j < len(headers) && j < len(data); j++ {
			fieldMap[headers[j]] = data[j]
		}

		results = append(results, fieldMap)
	}

	return headers, results, nil
}

// splitAndTrim splits a string by newline and trims each line
func splitAndTrim(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			result = append(result, trimmed)
		}
	}

	return result
}
