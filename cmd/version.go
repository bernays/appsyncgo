package cmd

import "time"

// VersionInfo holds version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    time.Time
}

// ParseDate parse string using time.RFC3339 format or default to time.Now()
func ParseDate(value string) time.Time {
	if value == "unknown" {
		return time.Now()
	}

	parsedDate, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsedDate = time.Now()
	}
	return parsedDate
}
