// Package stats provides functionality for interacting with HAProxy stats API
package stats

// StatHelpers contains helper methods for working with HAProxy stats
type StatHelpers struct{}

// GetItemType returns the type value from the stats item
func (h *StatHelpers) GetItemType(item StatsItem) int {
	return item.GetType()
}

// GetItemWeight returns the weight value from the stats item
func (h *StatHelpers) GetItemWeight(item StatsItem) int {
	return item.GetWeight()
}

// GetItemSessions returns the current sessions value from the stats item
func (h *StatHelpers) GetItemSessions(item StatsItem) int {
	val, _ := item.GetInt("scur")
	return val
}

// GetItemMaxSessions returns the max sessions value from the stats item
func (h *StatHelpers) GetItemMaxSessions(item StatsItem) int {
	val, _ := item.GetInt("smax")
	return val
}

// GetItemBytesIn returns the bytes in value from the stats item
func (h *StatHelpers) GetItemBytesIn(item StatsItem) int64 {
	val, _ := item.GetInt64("bin")
	return val
}

// GetItemBytesOut returns the bytes out value from the stats item
func (h *StatHelpers) GetItemBytesOut(item StatsItem) int64 {
	val, _ := item.GetInt64("bout")
	return val
}

// AddValueToMap adds a stats item field to a map if it has a value
func (h *StatHelpers) AddValueToMap(m map[string]interface{}, key string, item StatsItem, fieldKey string) {
	if val, ok := item.GetInt(fieldKey); ok && val != 0 {
		m[key] = val
	}
}

// AddInt64ValueToMap adds a stats item int64 field to a map if it has a value
func (h *StatHelpers) AddInt64ValueToMap(m map[string]interface{}, key string, item StatsItem, fieldKey string) {
	if val, ok := item.GetInt64(fieldKey); ok && val != 0 {
		m[key] = val
	}
}

// ExtractValues adds multiple common stats to a map in one call
func (h *StatHelpers) ExtractValues(m map[string]interface{}, item StatsItem) {
	// Basic info
	m["type"] = item.GetType()
	m["status"] = item.GetStatus()
	m["weight"] = item.GetWeight()

	// Session info
	h.AddValueToMap(m, "current_sessions", item, "scur")
	h.AddValueToMap(m, "max_sessions", item, "smax")

	// Bandwidth info
	h.AddInt64ValueToMap(m, "bytes_in", item, "bin")
	h.AddInt64ValueToMap(m, "bytes_out", item, "bout")

	// Rate info
	h.AddValueToMap(m, "rate", item, "rate")
	h.AddValueToMap(m, "rate_max", item, "rate_max")
}

// For backward compatibility, providing the global functions
var helpers = &StatHelpers{}

// GetStatsItemType returns the type value from the stats item
func GetStatsItemType(item StatsItem) int {
	return helpers.GetItemType(item)
}

// GetStatsItemWeight returns the weight value from the stats item
func GetStatsItemWeight(item StatsItem) int {
	return helpers.GetItemWeight(item)
}

// GetStatsItemSessions returns the current sessions value from the stats item
func GetStatsItemSessions(item StatsItem) int {
	return helpers.GetItemSessions(item)
}

// GetStatsItemMaxSessions returns the max sessions value from the stats item
func GetStatsItemMaxSessions(item StatsItem) int {
	return helpers.GetItemMaxSessions(item)
}

// GetStatsItemBytesIn returns the bytes in value from the stats item
func GetStatsItemBytesIn(item StatsItem) int64 {
	return helpers.GetItemBytesIn(item)
}

// GetStatsItemBytesOut returns the bytes out value from the stats item
func GetStatsItemBytesOut(item StatsItem) int64 {
	return helpers.GetItemBytesOut(item)
}

// AddStatsValueToMap adds a stats item field to a map if it has a value
func AddStatsValueToMap(m map[string]interface{}, key string, item StatsItem, fieldKey string) {
	helpers.AddValueToMap(m, key, item, fieldKey)
}

// AddStatsInt64ValueToMap adds a stats item int64 field to a map if it has a value
func AddStatsInt64ValueToMap(m map[string]interface{}, key string, item StatsItem, fieldKey string) {
	helpers.AddInt64ValueToMap(m, key, item, fieldKey)
}

// ExtractStatValues adds multiple common stats to a map in one call
func ExtractStatValues(m map[string]interface{}, item StatsItem) {
	helpers.ExtractValues(m, item)
}
