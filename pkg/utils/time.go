package utils

import (
	"time"
)

// GetCurrentUnixTimestamp returns the current Unix timestamp in seconds
func GetCurrentUnixTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentUnixTimestampMillis returns the current Unix timestamp in milliseconds
func GetCurrentUnixTimestampMillis() int64 {
	return time.Now().UnixMilli()
}

// UnixToTime converts a Unix timestamp (seconds) to time.Time
func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FormatTimestamp formats a Unix timestamp as RFC3339 string
func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(time.RFC3339)
}

// ParseRFC3339 parses an RFC3339 formatted string to time.Time
func ParseRFC3339(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// GetDuration calculates the duration between two Unix timestamps
func GetDuration(startTimestamp, endTimestamp int64) time.Duration {
	return time.Duration(endTimestamp-startTimestamp) * time.Second
}

// IsWithinWindow checks if a timestamp is within a time window from now
func IsWithinWindow(timestamp int64, window time.Duration) bool {
	eventTime := time.Unix(timestamp, 0)
	now := time.Now()
	return now.Sub(eventTime) <= window
}

// GetWindowStart returns the start timestamp for a time window
func GetWindowStart(window time.Duration) int64 {
	return time.Now().Add(-window).Unix()
}
