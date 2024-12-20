package static

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStaticOptions(t *testing.T) {
	// query
	staticModule := NewModule(ModuleConfig{
		StaticOptions: []StaticOption{
			{
				Id:          "1",
				Name:        "custom-option",
				DisplayName: "Custom Option",
				Description: "This is a custom option",
				Web:         "https://example.com",
			},
		},
	})
	options, err := staticModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Equal(t, "1", options[0].Id)
	require.Equal(t, "custom-option", options[0].Name)
	require.Equal(t, "Custom Option", options[0].DisplayName)
	require.Equal(t, "This is a custom option", options[0].Description)
	require.Equal(t, "https://example.com", options[0].Web)
}
