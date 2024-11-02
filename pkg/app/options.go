package app

import (
	"errors"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"slices"
)

// CollectOptions collects the options from the providers, optionally filtered by name
func CollectOptions(modules []recon.Module, byName []string, maxCacheAge int) ([]recon.Option, []error) {
	var options []recon.Option
	var errs []error

	for _, m := range modules {
		if len(byName) > 0 && !slices.Contains(byName, m.Name()) {
			continue
		}

		opts, err := m.OptionsOrCache(float64(maxCacheAge))
		if err != nil {
			errs = append(errs, errors.Join(types.ErrFailedToGetOptionsFromProvider, err))
		}

		options = append(options, opts...)
	}

	return options, errs
}

// FilterOptions filters the options, showTags are required, hideTags
func FilterOptions(options []recon.Option, showTags []string, hideTags []string) []recon.Option {
	var filtered []recon.Option
	hideTags = append(hideTags, "hidden") // always hide hidden options, used for e.g. git ssh hosts

	for _, o := range options {
		showTagFound := false
		for _, showTag := range showTags {
			if slices.Contains(o.Tags, showTag) {
				showTagFound = true
				break
			}
		}

		hideTagFound := false
		for _, hideTag := range hideTags {
			if slices.Contains(o.Tags, hideTag) {
				hideTagFound = true
				break
			}
		}

		if (showTagFound && !hideTagFound) || (len(showTags) == 0 && !hideTagFound) {
			filtered = append(filtered, o)
		}
	}

	return filtered
}
