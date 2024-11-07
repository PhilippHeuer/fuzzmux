package backstage

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/tdabasinskas/go-backstage/v2/backstage"
	"slices"
	"strings"
)

const moduleName = "backstage"

type Module struct {
	Config config.BackstageModuleConfig
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

	// connect
	client, err := backstage.NewClient(p.Config.Host, "default", nil)
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

		if slices.Contains(p.Config.Query, entityType) || len(p.Config.Query) == 0 {
			data := map[string]interface{}{
				"metadata.name":        entity.Metadata.Name,
				"metadata.namespace":   entity.Metadata.Namespace,
				"metadata.description": entity.Metadata.Description,
				"consumedBy":           entityRelationToString(entity.Relations, "apiConsumedBy"),
				"dependsOn":            entityRelationToString(entity.Relations, "dependsOn"),
				"ownedBy":              entityRelationToString(entity.Relations, "ownedBy"),
				"partOf":               entityRelationToString(entity.Relations, "partOf"),
			}
			for key, value := range entity.Spec {
				data["spec."+key] = value
			}
			for key, value := range entity.Metadata.Labels {
				data["metadata.labels."+key] = value
			}
			for key, value := range entity.Metadata.Annotations {
				data["metadata.annotations."+key] = value
			}
			attributes := recon.AttributeMapping(data, p.Config.AttributeMapping)
			attributes["web"] = fmt.Sprintf("%s/catalog/%s/%s/%s", p.Config.Host, entity.Metadata.Namespace, strings.ToLower(entity.Kind), entity.Metadata.Name)

			result = append(result, recon.Option{
				ProviderName: p.Name(),
				ProviderType: p.Type(),
				Id:           entity.Metadata.Name,
				DisplayName:  fmt.Sprintf("%s [%s]", entity.Metadata.Name, getStringValue(entity.Spec, "type")),
				Name:         entity.Metadata.Name,
				Tags:         []string{"backstage", entityType},
				Context:      attributes,
			})
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

func NewModule(config config.BackstageModuleConfig) Module {
	return Module{
		Config: config,
	}
}
