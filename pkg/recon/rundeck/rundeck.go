package rundeck

import (
	"fmt"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/rs/zerolog/log"
)

const moduleType = "rundeck"

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

	// AccessToken is the token used to authenticate against the Rundeck API
	AccessToken string `yaml:"token,omitempty"`

	// AttributeMapping is a list of field mappings used to map additional attributes to context fields
	AttributeMapping []types.FieldMapping `yaml:"attribute-mapping"`

	//Projects is a list of projects to query
	Projects []string
}

func (c *ModuleConfig) DecodeConfig() {
	c.Host = util.ResolveCredentialValue(c.Host)
	c.AccessToken = util.ResolveCredentialValue(c.AccessToken)
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
	p.Config.DecodeConfig()
	var result []recon.Option

	// setup client
	log.Debug().Str("host", p.Config.Host).Msg("connecting to rundeck")
	client := NewClient(p.Config.Host, p.Config.AccessToken)

	// query
	for _, project := range p.Config.Projects {
		log.Debug().Str("host", p.Config.Host).Str("project", project).Msg("querying rundeck jobs")
		jobs, err := client.GetJobs(project, nil)
		if err != nil {
			return nil, err
		}

		for _, job := range jobs {
			jobPath := job.Name
			if job.Group != "" {
				jobPath = job.Group + "/" + job.Name
			}

			entryAttributes := map[string]interface{}{
				"job.id":          job.ID,
				"job.name":        job.Name,
				"job.path":        jobPath,
				"job.group":       job.Group,
				"job.project":     job.Project,
				"job.description": job.Description,
				"job.enabled":     job.Enabled,
				"job.scheduled":   job.Scheduled,
			}
			context := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)

			opt := recon.Option{
				ProviderName:   p.Name(),
				ProviderType:   p.Type(),
				Id:             job.ID,
				DisplayName:    fmt.Sprintf("%s [%s] - %s", jobPath, job.Project, job.Description),
				Name:           jobPath,
				Description:    job.Description,
				Web:            job.Permalink,
				StartDirectory: "~",
				Tags:           []string{"rundeck", "job"},
				Context:        context,
				ModuleContext: map[string]string{
					"rundeckHost":  p.Config.Host,
					"rundeckToken": p.Config.AccessToken,
				},
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
	return recon.DefaultColumns()
}

func NewModule(config ModuleConfig) Module {
	return Module{
		Config: config,
	}
}
