package keycloak

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

var keycloakMockRequest = testcontainers.ContainerRequest{
	Image:        "quay.io/keycloak/keycloak:latest",
	ExposedPorts: []string{"8080/tcp"},
	WaitingFor:   wait.ForLog("[io.quarkus] (main) Installed features"),
	Env: map[string]string{
		"KC_BOOTSTRAP_ADMIN_USERNAME": "admin",
		"KC_BOOTSTRAP_ADMIN_PASSWORD": "secret",
	},
	Cmd: []string{"start-dev", "-Dkeycloak.profile.feature.upload_scripts=enabled"},
}

func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func TestSearchUsers(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("skipping test")
	}

	ctx := context.Background()
	keycloakServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: keycloakMockRequest,
		Started:          true,
		Reuse:            false,
	})
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, keycloakServer)
	ldapEndpoint, err := keycloakServer.Endpoint(ctx, "")
	require.NoError(t, err)

	// query
	keycloakModule := NewModule(ModuleConfig{
		Host:      "http://" + ldapEndpoint,
		RealmName: "master",
		Username:  "admin",
		Password:  "secret",
		Query:     []KeycloakContent{KeycloakUser},
	})
	options, err := keycloakModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Equal(t, "admin", options[0].Name)
	require.NotEmpty(t, options[0].Id)
}
