package ldap

import (
	"context"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/testcontainers/testcontainers-go"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

var ldapMockRequest = testcontainers.ContainerRequest{
	Image:        "docker.io/thoteam/slapd-server-mock:latest",
	ExposedPorts: []string{"389/tcp"},
	WaitingFor:   wait.ForLog("slapd starting"),
	Env: map[string]string{
		"LDAP_ORGANIZATION": "example org",
		"LDAP_DOMAIN":       "example.com",
		"LDAP_SECRET":       "secret",
	},
}

func TestSearchUsers(t *testing.T) {
	ctx := context.Background()
	ldapServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: ldapMockRequest,
		Started:          true,
	})
	require.NoError(t, err)
	ldapEndpoint, err := ldapServer.Endpoint(ctx, "")
	require.NoError(t, err)

	// query
	ldapModule := NewModule(config.LDAPModuleConfig{
		Host:                  "ldap://" + ldapEndpoint,
		BaseDistinguishedName: "dc=example,dc=com",
		BindDistinguishedName: "cn=admin,dc=example,dc=com",
		BindPassword:          "secret",
		Filter:                "(objectClass=person)",
	})
	options, err := ldapModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Equal(t, "Admin User1", options[0].Name)
	require.Equal(t, "uid=adminuser1,ou=people,dc=example,dc=com", options[0].Id)

	// cleanup
	testcontainers.CleanupContainer(t, ldapServer)
	require.NoError(t, err)
}

func TestSearchGroups(t *testing.T) {
	ctx := context.Background()
	ldapServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: ldapMockRequest,
		Started:          true,
	})
	require.NoError(t, err)
	ldapEndpoint, err := ldapServer.Endpoint(ctx, "")
	require.NoError(t, err)

	// query
	ldapModule := NewModule(config.LDAPModuleConfig{
		Host:                  "ldap://" + ldapEndpoint,
		BaseDistinguishedName: "dc=example,dc=com",
		BindDistinguishedName: "cn=admin,dc=example,dc=com",
		BindPassword:          "secret",
		Filter:                "(|(objectClass=group)(objectClass=posixGroup)(objectClass=groupOfNames))",
	})
	options, err := ldapModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Equal(t, "admins", options[0].Name)
	require.Equal(t, "cn=admins,ou=groups,dc=example,dc=com", options[0].Id)
	require.Equal(t, "developers", options[1].Name)
	require.Equal(t, "cn=developers,ou=groups,dc=example,dc=com", options[1].Id)

	// cleanup
	testcontainers.CleanupContainer(t, ldapServer)
	require.NoError(t, err)
}
