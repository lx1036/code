package libdocker

import "time"

// ParseDockerTimestamp parses the timestamp returned by Interface from string to time.Time
func ParseDockerTimestamp(s string) (time.Time, error) {
	// Timestamp returned by Docker is in time.RFC3339Nano format.
	return time.Parse(time.RFC3339Nano, s)
}
