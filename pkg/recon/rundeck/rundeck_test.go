package rundeck

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var rundeckContainerRequest = testcontainers.ContainerRequest{
	Image:        "docker.io/rundeck/rundeck:5.7.0",
	ExposedPorts: []string{"4440/tcp"},
	WaitingFor:   wait.ForLog(`Started Application in`),
	Env: map[string]string{
		"RUNDECK_TOKENS_FILE": "/home/rundeck/server/etc/rundeck/tokens.properties",
	},
	Files: []testcontainers.ContainerFile{
		{
			HostFilePath:      "testdata/grailsdb.mv.db", // example db with one project and job
			ContainerFilePath: "/home/rundeck/server/data/grailsdb.mv.db",
			FileMode:          0o777,
		},
		{
			HostFilePath:      "testdata/tokens.properties", // static api tokens for testing
			ContainerFilePath: "/home/rundeck/server/etc/rundeck/tokens.properties",
			FileMode:          0o777,
		},
	},
}

func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func TestSearchJobs(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("skipping test")
	}

	ctx := context.Background()
	rundeckServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: rundeckContainerRequest,
		Started:          true,
		Reuse:            false,
	})
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, rundeckServer)
	rundeckEndpoint, err := rundeckServer.Endpoint(ctx, "")
	require.NoError(t, err)

	// query
	rundeckModule := NewModule(ModuleConfig{
		Host:        "http://" + rundeckEndpoint,
		AccessToken: "token", // static user token for testing, see token.properties
		Projects:    []string{"example"},
	})
	options, err := rundeckModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Equal(t, "7fdae203-1822-4cdc-bf90-6e8e5a0e074a", options[0].Id)
	require.Equal(t, "Awesome Job", options[0].Name)
}
