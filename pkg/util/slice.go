package util

import "slices"

func AddToSet(slice []string, value string) []string {
	if !slices.Contains(slice, value) {
		return append(slice, value)
	}
	return slice
}
