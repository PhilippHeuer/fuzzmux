package jira

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/andygrunwald/go-jira"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"net/http"
)

const moduleName = "jira"

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Host is the Backstage hostname or IP address
	Host string `yaml:"host"`

	// BearerToken is the token used to authenticate against the Backstage API (see https://backstage.io/docs/auth/service-to-service-auth/#static-tokens)
	BearerToken string `yaml:"bearer-token,omitempty"`

	// AttributeMapping is a list of field mappings used to map additional attributes to context fields
	AttributeMapping []types.FieldMapping `yaml:"attribute-mapping"`

	// Jql is the Jira Query Language query to filter issues
	Jql string `yaml:"jql"`
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
	var options []recon.Option

	// httpClient
	var httpClient *http.Client
	if p.Config.BearerToken != "" {
		log.Debug().Msg("using bearer token for jira authentication")
		httpClient = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: p.Config.BearerToken,
			TokenType:   "Bearer",
		}))
	}

	// connect
	log.Debug().Str("host", p.Config.Host).Msg("connecting to jira")
	jiraClient, err := jira.NewClient(httpClient, p.Config.Host)
	if err != nil {
		return nil, err
	}

	// query tickets
	log.Debug().Str("jql", p.Config.Jql).Msg("querying jira issues")
	issues, _, err := jiraClient.Issue.Search(p.Config.Jql, &jira.SearchOptions{})
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		entryAttributes := map[string]interface{}{
			"project": issue.Fields.Project.Name,
			"summary": issue.Fields.Summary,
			"type":    issue.Fields.Type.Name,
		}
		if issue.Fields.Status != nil {
			entryAttributes["status"] = issue.Fields.Status.Name
		}
		if issue.Fields.Priority != nil {
			entryAttributes["priority"] = issue.Fields.Priority.Name
		}
		if issue.Fields.Assignee != nil {
			entryAttributes["assignee"] = issue.Fields.Assignee.Name
		}
		if issue.Fields.Reporter != nil {
			entryAttributes["reporter"] = issue.Fields.Reporter.Name
		}
		if issue.Fields.Sprint != nil {
			entryAttributes["sprint.id"] = issue.Fields.Sprint.ID
			entryAttributes["sprint.name"] = issue.Fields.Sprint.Name
			entryAttributes["sprint.startedAt"] = issue.Fields.Sprint.StartDate
			entryAttributes["sprint.endedAt"] = issue.Fields.Sprint.EndDate
		}
		attributes := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)

		options = append(options, recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             issue.Key,
			DisplayName:    fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary),
			Name:           issue.Key,
			Description:    issue.Fields.Summary,
			Web:            fmt.Sprintf("%s/browse/%s", p.Config.Host, issue.Key),
			StartDirectory: "~",
			Tags:           []string{"jira"},
			Context:        attributes,
			ModuleContext: map[string]string{
				"jiraServer":      p.Config.Host,
				"jiraBearerToken": p.Config.BearerToken,
			},
		})
	}

	return options, nil
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
