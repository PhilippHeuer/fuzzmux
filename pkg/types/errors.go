package types

import (
	"errors"
)

var (
	ErrNoLauncherAvailable            = errors.New("no available launcher found")
	ErrNoProvidersAvailable           = errors.New("no providers available")
	ErrNoOptionsAvailable             = errors.New("no options available")
	ErrNoOptionSelected               = errors.New("no option selected")
	ErrReconModuleNotFound            = errors.New("recon module not found")
	ErrFailedToGetOptionsFromProvider = errors.New("failed to get options from recon")
	ErrFailedToRenderOptions          = errors.New("failed to render options")
	ErrSomeProvidersFailed            = errors.New("failed to get options from recon")
	ErrAllProvidersFailed             = errors.New("all providers failed to generate options")
	ErrFailedToCreateStartDirectory   = errors.New("failed to create start directory")
)
