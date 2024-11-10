package recon

import (
	"errors"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/types"
)

type Option struct {
	ProviderName   string            `json:"provider_name"`   // module name
	ProviderType   string            `json:"provider_type"`   // module type
	Id             string            `json:"id"`              // unique id
	DisplayName    string            `json:"display_name"`    // display name for the fuzzy finder
	Name           string            `json:"name"`            // name
	Description    string            `json:"description"`     // description
	Web            string            `json:"web"`             // web url
	StartDirectory string            `json:"start_directory"` // sets the initial working directory
	Tags           []string          `json:"tags"`            // tags
	Context        map[string]string `json:"context"`         // additional context information
	ModuleContext  map[string]string `json:"module_context"`  // internal context information, not exposed to the user
}

func (o Option) ResolveStartDirectory(full bool) string {
	startDirectory := o.StartDirectory
	if startDirectory == "" {
		startDirectory = "~"
	}
	startDirectory = util.ExpandPlaceholders(startDirectory, "name", o.Name)
	startDirectory = util.ExpandPlaceholders(startDirectory, "displayName", o.DisplayName)
	for k, v := range o.Context {
		startDirectory = util.ExpandPlaceholders(startDirectory, k, v)
	}

	if full {
		startDirectory = strings.Replace(startDirectory, "~", os.Getenv("HOME"), -1)
	}

	return startDirectory
}

func (o Option) CreateStartDirectoryIfMissing() error {
	if o.StartDirectory == "" || o.StartDirectory == "~" {
		return nil
	}

	dir := o.ResolveStartDirectory(true)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Join(types.ErrFailedToCreateStartDirectory, err)
		}
	}

	return nil
}

func (o Option) ResolvePlaceholders(input string) string {
	input = os.ExpandEnv(input)

	input = util.ExpandPlaceholders(input, "id", o.Id)
	input = util.ExpandPlaceholders(input, "name", o.Name)
	input = util.ExpandPlaceholders(input, "displayName", o.DisplayName)
	input = util.ExpandPlaceholders(input, "startDirectory", o.ResolveStartDirectory(true))

	// module context
	for k, v := range o.ModuleContext {
		input = util.ExpandPlaceholders(input, k, v)
	}

	// context
	for k, v := range o.Context {
		input = util.ExpandPlaceholders(input, k, v)
	}

	return input
}

func (o Option) RenderPreview() string {
	var builder strings.Builder
	builder.WriteString("# " + o.DisplayName + "\n\n")
	builder.WriteString(fmt.Sprintf("Provider: %s [TYPE: %s]\n", o.ProviderName, o.ProviderType))
	builder.WriteString("Directory: " + o.ResolveStartDirectory(true) + " [" + o.StartDirectory + "]\n")
	if len(o.Tags) > 0 {
		builder.WriteString("\nTags:\n")
		for _, t := range o.Tags {
			builder.WriteString("- " + t + "\n")
		}
	}

	// TODO: custom option render logic should move into the option recon
	switch o.ProviderName {
	case "kubernetes":
		builder.WriteString("\n")
		if o.Context["clusterName"] != "" {
			builder.WriteString(fmt.Sprintf("K8S Cluster Name: %s\n", o.Context["clusterName"]))
		}
		if o.Context["clusterHost"] != "" {
			builder.WriteString(fmt.Sprintf("K8S Cluster API: %s\n", o.Context["clusterHost"]))
		}
		if o.Context["clusterUser"] != "" {
			builder.WriteString(fmt.Sprintf("K8S Cluster User: %s\n", o.Context["clusterUser"]))
		}
		if o.Context["clusterType"] != "" {
			builder.WriteString(fmt.Sprintf("K8S Cluster Type: %s\n", o.Context["clusterType"]))
		}
	case "usql":
		builder.WriteString("\n")
		if o.Context["name"] != "" {
			builder.WriteString(fmt.Sprintf("Name: %s\n", o.Context["name"]))
		}
		if o.Context["hostname"] != "" {
			builder.WriteString(fmt.Sprintf("DB Host: %s\n", o.Context["hostname"]))
		}
		if o.Context["port"] != "" {
			builder.WriteString(fmt.Sprintf("DB Port: %s\n", o.Context["port"]))
		}
		if o.Context["username"] != "" {
			builder.WriteString(fmt.Sprintf("DB Username: %s\n", o.Context["username"]))
		}
		if o.Context["instance"] != "" {
			builder.WriteString(fmt.Sprintf("DB Instance/SID: %s\n", o.Context["instance"]))
		}
		if o.Context["database"] != "" {
			builder.WriteString(fmt.Sprintf("DB Database: %s\n", o.Context["database"]))
		}
	default:
		builder.WriteString("\n")
		if len(o.Context) > 0 {
			builder.WriteString("Context:\n")
			for k, v := range o.Context {
				builder.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}
	}

	// web url
	if o.Web != "" {
		builder.WriteString("\nURL: " + o.Web + "\n")
	}

	// free-text description (with fallback to context)
	if o.Description != "" {
		builder.WriteString("\n\n" + o.Description + "\n")
	} else if o.Context["description"] != "" {
		builder.WriteString("\n\n" + o.Context["description"] + "\n")
	}

	return builder.String()
}

type Column struct {
	Key    string
	Name   string
	Hidden bool
}

type Module interface {
	Name() string                                    // Name returns the name of the module
	Type() string                                    // Type returns the type of the module
	Options() ([]Option, error)                      // Options returns the options
	OptionsOrCache(maxAge float64) ([]Option, error) // OptionsOrCache returns the options from cache or calls Options
	SelectOption(options *Option) error              // Select can be used to run actions / enrich the context before opening the session
	Columns() []Column                               // Columns returns the columns for a tabular view
}

func DefaultColumns() []Column {
	return []Column{
		{Key: "module", Name: "Module"},
		{Key: "id", Name: "ID", Hidden: true},
		{Key: "name", Name: "Name", Hidden: true},
		{Key: "display_name", Name: "Display Name"},
	}
}
