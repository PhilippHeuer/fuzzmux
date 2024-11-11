package recon

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"reflect"
	"testing"
)

func TestContextMapping(t *testing.T) {
	tests := []struct {
		name          string
		attributes    map[string]interface{}
		mappingConfig []types.FieldMapping
		expected      map[string]string
	}{
		{
			name: "LDAP Time",
			attributes: map[string]interface{}{
				"created": "20050917134246.0Z",
			},
			mappingConfig: []types.FieldMapping{
				{Source: "created", Target: "createdAt", Format: "ldaptime"},
			},
			expected: map[string]string{
				"createdAt": "2005-09-17T13:42:46Z",
			},
		},
		{
			name: "Unix Timestamp in Milliseconds",
			attributes: map[string]interface{}{
				"modified": int64(1638316800000),
			},
			mappingConfig: []types.FieldMapping{
				{Source: "modified", Target: "updatedAt", Format: "unixmillis"},
			},
			expected: map[string]string{
				"updatedAt": "2021-12-01T00:00:00Z",
			},
		},
		{
			name: "String Slice",
			attributes: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			mappingConfig: []types.FieldMapping{
				{Source: "tags", Target: "tagList"},
			},
			expected: map[string]string{
				"tagList": "tag1, tag2",
			},
		},
		{
			name: "Empty Source or Target",
			attributes: map[string]interface{}{
				"created": "20050917134246.0Z",
			},
			mappingConfig: []types.FieldMapping{
				{Source: "", Target: "emptySource"},     // Missing source
				{Source: "created", Target: ""},         // Missing target
				{Source: "created", Target: "emptyMap"}, // Correct mapping to test alongside empty cases
			},
			expected: map[string]string{
				"emptyMap": "20050917134246.0Z",
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run AttributeMapping function with the provided attributes and mapping config
			result := AttributeMapping(tt.attributes, tt.mappingConfig)

			// Compare result with expected output
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("AttributeMapping(%v, %v) = %v; want %v", tt.attributes, tt.mappingConfig, result, tt.expected)
			}
		})
	}
}
