package errs

import (
	"errors"
)

var (
	ErrNoProvidersAvailable           = errors.New("no providers available")
	ErrNoOptionsAvailable             = errors.New("no options available")
	ErrNoOptionSelected               = errors.New("no option selected")
	ErrFailedToGetOptionsFromProvider = errors.New("failed to get options from provider")
	ErrFailedToRenderOptions          = errors.New("failed to render options")
)
