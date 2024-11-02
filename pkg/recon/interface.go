package recon

import (
	"errors"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/types"
)

type Option struct {
	ProviderName   string            `json:"provider_name"`   // recon name
	Id             string            `json:"id"`              // unique id
	DisplayName    string            `json:"display_name"`    // display name for the fuzzy finder
	Name           string            `json:"name"`            // name
	StartDirectory string            `json:"start_directory"` // sets the initial working directory
	Tags           []string          `json:"tags"`            // tags
	Context        map[string]string `json:"context"`         // additional context information
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

	input = util.ExpandPlaceholders(input, "name", o.Name)
	input = util.ExpandPlaceholders(input, "displayName", o.DisplayName)
	input = util.ExpandPlaceholders(input, "startDirectory", o.ResolveStartDirectory(true))

	for k, v := range o.Context {
		input = util.ExpandPlaceholders(input, k, v)
	}

	return input
}

type Module interface {
	Name() string                                    // Name returns the name of the recon
	Options() ([]Option, error)                      // Options returns the options
	OptionsOrCache(maxAge float64) ([]Option, error) // OptionsOrCache returns the options from cache or calls Options
	SelectOption(options *Option) error              // Select can be used to run actions / enrich the context before opening the session
}
