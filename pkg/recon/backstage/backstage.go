package backstage

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/rs/zerolog/log"
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
			result = append(result, recon.Option{
				ProviderName: p.Name(),
				ProviderType: p.Type(),
				Id:           entity.Metadata.Name,
				DisplayName:  fmt.Sprintf("%s [%s]", entity.Metadata.Name, getStringValue(entity.Spec, "type")),
				Name:         entity.Metadata.Name,
				Tags:         []string{"backstage", entityType},
				Context: map[string]string{
					"web":        fmt.Sprintf("%s/catalog/%s/%s/%s", p.Config.Host, entity.Metadata.Namespace, strings.ToLower(entity.Kind), entity.Metadata.Name),
					"owner":      getStringValue(entity.Spec, "owner"),
					"lifecycle":  getStringValue(entity.Spec, "lifecycle"),
					"consumedBy": entityRelationToString(entity.Relations, "apiConsumedBy"),
					"dependsOn":  entityRelationToString(entity.Relations, "dependsOn"),
					"ownedBy":    entityRelationToString(entity.Relations, "ownedBy"),
					"partOf":     entityRelationToString(entity.Relations, "partOf"),
				},
			})
		}
	}

	return result, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := recon.LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = recon.SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
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
