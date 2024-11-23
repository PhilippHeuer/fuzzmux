package backstage

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/datolabs-io/go-backstage/v3"
	"golang.org/x/oauth2"
	"net/http"
	"slices"
	"strings"
)

const moduleType = "backstage"

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// Host is the Backstage hostname or IP address
	Host string `yaml:"host"`

	// BearerToken is the token used to authenticate against the Backstage API (see https://backstage.io/docs/auth/service-to-service-auth/#static-tokens)
	BearerToken string `yaml:"bearer-token,omitempty"`

	// AttributeMapping is a list of field mappings used to map additional attributes to context fields
	AttributeMapping []types.FieldMapping `yaml:"attribute-mapping"`

	// Query is a list of content types that should be queried
	Query []string `yaml:"query"`
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleType
}

func (p Module) Type() string {
	return moduleType
}

func (p Module) Options() ([]recon.Option, error) {
	var result []recon.Option

	// httpClient
	var httpClient *http.Client
	if p.Config.BearerToken != "" {
		httpClient = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: p.Config.BearerToken,
			TokenType:   "Bearer",
		}))
	}

	// connect
	client, err := backstage.NewClient(p.Config.Host, "default", httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backstage: %w", err)
	}

	// query
	entities, _, err := client.Catalog.Entities.List(context.Background(), &backstage.ListEntityOptions{
		Filters: []string{},
		Fields:  []string{},
		Order:   []backstage.ListEntityOrder{{Direction: backstage.OrderDescending, Field: "metadata.name"}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query backstage: %w", err)
	}

	for _, entity := range entities {
		entityType := getStringValue(entity.Spec, "type")
		queryId := strings.ToLower(entity.Kind)
		if entityType != "" {
			queryId = fmt.Sprintf("%s/%s", queryId, entityType)
		}

		if slices.Contains(p.Config.Query, queryId) || len(p.Config.Query) == 0 {
			data := map[string]interface{}{
				"kind":                 entity.Kind,
				"spec.type":            getStringValue(entity.Spec, "type"),
				"metadata.name":        entity.Metadata.Name,
				"metadata.namespace":   entity.Metadata.Namespace,
				"metadata.description": entity.Metadata.Description,
				"consumedBy":           entityRelationToString(entity.Relations, "apiConsumedBy"),
				"dependsOn":            entityRelationToString(entity.Relations, "dependsOn"),
				"ownedBy":              entityRelationToString(entity.Relations, "ownedBy"),
				"partOf":               entityRelationToString(entity.Relations, "partOf"),
			}
			for key, value := range entity.Spec {
				if str, ok := value.(string); ok { // TODO: handle nested objects
					data["spec."+key] = str
				}
			}
			for key, value := range entity.Metadata.Labels {
				data["metadata.labels."+key] = value
			}
			for key, value := range entity.Metadata.Annotations {
				data["metadata.annotations."+key] = value
			}
			attributes := recon.AttributeMapping(data, p.Config.AttributeMapping)

			opt := recon.Option{
				ProviderName: p.Name(),
				ProviderType: p.Type(),
				Id:           entity.Metadata.Name,
				DisplayName:  fmt.Sprintf("%s [%s]", entity.Metadata.Name, queryId),
				Name:         entity.Metadata.Name,
				Description:  entity.Metadata.Description,
				Web:          fmt.Sprintf("%s/catalog/%s/%s/%s", p.Config.Host, entity.Metadata.Namespace, strings.ToLower(entity.Kind), entity.Metadata.Name),
				Tags:         []string{"backstage", entityType},
				Context:      attributes,
			}
			opt.ProcessUserTemplateStrings(p.Config.DisplayName, p.Config.StartDirectory)
			result = append(result, opt)
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
	return append(recon.DefaultColumns())
}

func NewModule(config ModuleConfig) Module {
	return Module{
		Config: config,
	}
}
