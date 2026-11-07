package keycloak

import (
	"fmt"
	"strings"
)

func attributesToMap(attr map[string]string, additionalAttributes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if attr != nil {
		for key, value := range attr {
			result[key] = value
		}
	}

	for key, value := range additionalAttributes {
		result[key] = value
	}

	return result
}

func attributeSlicesToMap(attr map[string][]string, additionalAttributes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if attr != nil {
		for key, values := range attr {
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

// pageSize is the number of entries requested per page from the keycloak admin api
const pageSize = 100

// paginate calls fetch with increasing offsets until a page with fewer than pageSize entries is returned
func paginate[T any](fetch func(first, max int) ([]T, error)) ([]T, error) {
	var result []T
	for first := 0; ; first += pageSize {
		page, err := fetch(first, pageSize)
		if err != nil {
			return nil, err
		}
		result = append(result, page...)
		if len(page) < pageSize {
			return result, nil
		}
		if len(page) > pageSize {
			// server ignored the pagination parameters and returned everything, requesting more would loop forever
			if first > 0 {
				return page, nil
			}
			return result, nil
		}
	}
}
