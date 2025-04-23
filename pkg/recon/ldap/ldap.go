package ldap

import (
	"fmt"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/go-ldap/ldap/v3"
	"github.com/rs/zerolog/log"
)

const moduleType = "ldap"

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

	// Host is the LDAP server hostname or IP address
	Host string `yaml:"host"`

	// BaseDistinguishedName (DN) for LDAP base search (e.g., "dc=example,dc=com")
	BaseDistinguishedName string `yaml:"base-dn"`

	// BindDistinguishedName (DN) used for LDAP binding (e.g., "cn=admin,dc=example,dc=com")
	BindDistinguishedName string `yaml:"bind-dn"`

	// Password for LDAP bind user
	BindPassword string `yaml:"bind-password"`

	// AttributeMapping is a list of field mappings used to map LDAP fields to context fields
	AttributeMapping []types.FieldMapping `yaml:"attribute-mapping"`

	// Filter is the LDAP search filter (e.g., "(&(objectClass=organizationalPerson))")
	Filter string `yaml:"filter"`
}

func (c *ModuleConfig) DecodeConfig() {
	c.Host = util.ResolveCredentialValue(c.Host)
	c.BindDistinguishedName = util.ResolveCredentialValue(c.BindDistinguishedName)
	c.BindPassword = util.ResolveCredentialValue(c.BindPassword)
	c.BaseDistinguishedName = util.ResolveCredentialValue(c.BaseDistinguishedName)
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

	// connect
	log.Debug().Str("host", p.Config.Host).Msg("connecting to ldap")
	l, err := ldap.DialURL(p.Config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ldap: %w", err)
	}
	defer l.Close()

	// bind
	if p.Config.BindDistinguishedName != "" {
		log.Debug().Str("bindDN", p.Config.BindDistinguishedName).Msg("binding to ldap")
		err = l.Bind(p.Config.BindDistinguishedName, p.Config.BindPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to bind to ldap: %w", err)
		}
	}

	// search
	log.Debug().Str("filter", p.Config.Filter).Str("base", p.Config.BaseDistinguishedName).Msg("searching ldap")
	searchRequest := ldap.NewSearchRequest(
		p.Config.BaseDistinguishedName,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		p.Config.Filter,
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search ldap: %w", err)
	}

	// add to result
	for _, entry := range sr.Entries {
		log.Trace().Str("dn", entry.DN).Interface("attributes", entry.Attributes).Msg("found entry")
		entryAttributes := attributesToMap(entry.Attributes)
		context := recon.AttributeMapping(entryAttributes, p.Config.AttributeMapping)

		opt := recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             entry.DN,
			DisplayName:    fmt.Sprintf("%s [%s]", entry.GetAttributeValue("cn"), entry.DN),
			Name:           entry.GetAttributeValue("cn"),
			StartDirectory: "~",
			Tags:           []string{"ldap"},
			Context:        context,
			ModuleContext: map[string]string{
				"ldapHost":         p.Config.Host,
				"ldapBindDN":       p.Config.BindDistinguishedName,
				"ldapBindPassword": p.Config.BindPassword,
			},
		}
		opt.ProcessUserTemplateStrings(p.Config.DisplayName, p.Config.StartDirectory)
		result = append(result, opt)
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
		recon.Column{Key: "user", Name: "User"},
	)
}

func NewModule(config ModuleConfig) Module {
	if config.Filter == "" {
		config.Filter = "(|(objectClass=*))"
	}

	return Module{
		Config: config,
	}
}

func attributesToMap(attr []*ldap.EntryAttribute) map[string]interface{} {
	result := make(map[string]interface{}, len(attr))
	for _, a := range attr {
		result[a.Name] = strings.Join(a.Values, ",")
	}
	return result
}
