package ldap

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/go-ldap/ldap/v3"
	"github.com/rs/zerolog/log"
	"strings"
)

const moduleName = "ldap"

type Module struct {
	Config config.LDAPModuleConfig
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

		result = append(result, recon.Option{
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
		})
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

func NewModule(config config.LDAPModuleConfig) Module {
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
