package keycloak

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	data := make([]int, 250)
	for i := range data {
		data[i] = i
	}

	result, err := paginate(func(first, max int) ([]int, error) {
		end := min(first+max, len(data))
		if first >= len(data) {
			return nil, nil
		}
		return data[first:end], nil
	})
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestPaginateServerIgnoresParameters(t *testing.T) {
	data := make([]int, 150)
	calls := 0

	result, err := paginate(func(first, max int) ([]int, error) {
		calls++
		return data, nil
	})
	require.NoError(t, err)
	assert.Len(t, result, 150)
	assert.Equal(t, 1, calls)
}
