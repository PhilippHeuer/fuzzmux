package keycloak

import (
	"fmt"
	"strings"
	"time"
)

// FormatTimestampToISO takes a pointer to an int64 timestamp (in milliseconds) and returns a formatted ISO 8601 date string.
// If the timestamp is nil, it returns an empty string.
func timestampToISO(timestamp *int64) string {
	if timestamp == nil {
		return ""
	}

	seconds := *timestamp / 1000
	return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
}

func clientRolesToString(clientRoles map[string][]string) string {
	var roles []string

	for client, roleList := range clientRoles {
		if len(roleList) > 0 {
			roles = append(roles, fmt.Sprintf("%s=%s", client, strings.Join(roleList, ",")))
		}
	}

	return strings.Join(roles, " ")
}
