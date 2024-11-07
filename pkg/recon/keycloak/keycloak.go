package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/cidverse/go-ptr"
	"github.com/rs/zerolog/log"
	"slices"
)

const moduleName = "keycloak"

type Module struct {
	Config config.KeycloakModuleConfig
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleName
}

func (p Module) Type() string {
	return moduleName
}

func (p Module) Options() ([]recon.Option, error) {
	var result []recon.Option
	ctx := context.Background()

	// connect and login
	log.Debug().Str("host", p.Config.Host).Str("realm", p.Config.RealmName).Str("user", p.Config.Username).Msg("connecting to keycloak")
	client := gocloak.NewClient(p.Config.Host)
	token, err := client.LoginAdmin(ctx, p.Config.Username, p.Config.Password, p.Config.RealmName)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate on keycloak: %w", err)
	}

	// query realms
	realms, err := client.GetRealms(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get realms: %w", err)
	}

	// query content
	for _, realm := range realms {
		// clients
		if slices.Contains(p.Config.Query, config.KeycloakClient) {
			allClients, err := client.GetClients(ctx, token.AccessToken, ptr.Value(realm.Realm), gocloak.GetClientsParams{})
			if err != nil {
				return nil, fmt.Errorf("failed to get clients: %w", err)
			}
			for _, cl := range allClients {
				entryAttributes := attributesToMap(cl.Attributes, map[string]interface{}{
					"enabled":      ptr.Value(cl.Enabled),
					"clientId":     ptr.Value(cl.ClientID),
					"rootUrl":      ptr.Value(cl.RootURL),
					"protocol":     ptr.Value(cl.Protocol),
					"publicClient": ptr.Value(cl.PublicClient),
					"description":  ptr.Value(cl.Description),
				})
				attributes := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)
				attributes["web"] = fmt.Sprintf("%s/admin/%s/console/#/%s/clients/%s/settings", p.Config.Host, ptr.Value(realm.Realm), ptr.Value(realm.Realm), ptr.Value(cl.ID))
				attributes["type"] = "client"

				result = append(result, recon.Option{
					ProviderName: p.Name(),
					ProviderType: p.Type(),
					Id:           ptr.Value(cl.ID),
					DisplayName:  fmt.Sprintf("%s [%s] @ %s", ptr.Value(cl.ClientID), "client", ptr.Value(realm.Realm)),
					Name:         ptr.Value(cl.ClientID),
					Tags:         []string{"keycloak", "client"},
					Context:      attributes,
				})
			}
		}
		// users
		if slices.Contains(p.Config.Query, config.KeycloakUser) {
			users, err := client.GetUsers(ctx, token.AccessToken, ptr.Value(realm.Realm), gocloak.GetUsersParams{BriefRepresentation: ptr.False()})
			if err != nil {
				return nil, fmt.Errorf("failed to get users: %w", err)
			}
			for _, user := range users {
				entryAttributes := attributeSlicesToMap(user.Attributes, map[string]interface{}{
					"enabled":     ptr.Value(user.Enabled),
					"email":       ptr.Value(user.Email),
					"firstname":   ptr.Value(user.FirstName),
					"lastname":    ptr.Value(user.LastName),
					"groups":      user.Groups,
					"realmRoles":  user.RealmRoles,
					"clientRoles": clientRolesToString(ptr.Value(user.ClientRoles)),
					"createdAt":   util.ConvertMilliUnixTimestampToRFC3339(user.CreatedTimestamp),
				})
				attributes := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)
				attributes["web"] = fmt.Sprintf("%s/admin/%s/console/#/%s/users/%s", p.Config.Host, ptr.Value(realm.Realm), ptr.Value(realm.Realm), ptr.Value(user.ID))
				attributes["type"] = "user"

				result = append(result, recon.Option{
					ProviderName: p.Name(),
					ProviderType: p.Type(),
					Id:           ptr.Value(user.ID),
					DisplayName:  fmt.Sprintf("%s [%s] @ %s", ptr.Value(user.Username), "user", ptr.Value(realm.Realm)),
					Name:         ptr.Value(user.Username),
					Tags:         []string{"keycloak", "user"},
					Context:      attributes,
				})
			}
		}
		// groups
		if slices.Contains(p.Config.Query, config.KeycloakGroup) {
			groups, err := client.GetGroups(ctx, token.AccessToken, ptr.Value(realm.Realm), gocloak.GetGroupsParams{BriefRepresentation: ptr.False()})
			if err != nil {
				return nil, fmt.Errorf("failed to get groups: %w", err)
			}
			for _, group := range groups {
				entryAttributes := attributeSlicesToMap(group.Attributes, map[string]interface{}{})
				attributes := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)
				attributes["web"] = fmt.Sprintf("%s/admin/%s/console/#/%s/groups/%s/settings", p.Config.Host, ptr.Value(realm.Realm), ptr.Value(realm.Realm), ptr.Value(group.ID))
				attributes["type"] = "group"

				result = append(result, recon.Option{
					ProviderName: p.Name(),
					ProviderType: p.Type(),
					Id:           ptr.Value(group.ID),
					DisplayName:  fmt.Sprintf("%s [%s] @ %s", ptr.Value(group.Name), "group", ptr.Value(realm.Realm)),
					Name:         ptr.Value(group.Name),
					Tags:         []string{"keycloak", "group"},
					Context:      attributes,
				})
			}
		}
	}

	return result, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return append(recon.DefaultColumns(),
		recon.Column{Key: "type", Name: "Type"},
	)
}

func NewModule(config config.KeycloakModuleConfig) Module {
	return Module{
		Config: config,
	}
}
