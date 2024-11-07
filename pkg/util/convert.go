package util

import (
	"time"
)

// ConvertLDAPTimeToRFC3339 converts an Active Directory Generalized Time format (YYYYMMDDHHMMSS.0Z) string to RFC3339 format.
func ConvertLDAPTimeToRFC3339(adDate string) string {
	if adDate == "" {
		return ""
	}

	parsedTime, err := time.Parse("20060102150405.0Z", adDate)
	if err != nil {
		return ""
	}

	return parsedTime.Format(time.RFC3339)
}

// ConvertMilliUnixTimestampToRFC3339 takes a pointer to an int64 timestamp (in milliseconds) and returns a formatted RFC3339 date string.
func ConvertMilliUnixTimestampToRFC3339(timestamp *int64) string {
	if timestamp == nil {
		return ""
	}

	seconds := *timestamp / 1000
	return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
}
