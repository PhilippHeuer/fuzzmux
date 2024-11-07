package keycloak

import (
	"fmt"
	"strings"
)

func attributesToMap(attr *map[string]string, additionalAttributes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if attr != nil {
		for key, value := range *attr {
			result[key] = value
		}
	}

	for key, value := range additionalAttributes {
		result[key] = value
	}

	return result
}

func attributeSlicesToMap(attr *map[string][]string, additionalAttributes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if attr != nil {
		for key, values := range *attr {
			if len(values) == 1 {
				result[key] = values[0]
			} else {
				result[key] = values
			}
		}
	}

	for key, value := range additionalAttributes {
		result[key] = value
	}

	return result
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
