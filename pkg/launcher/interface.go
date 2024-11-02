package launcher

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
)

type Opts struct {
	SessionName string
	Layout      config.Layout
	AppendMode  AppendMode
}

type AppendMode string

const (
	CreateOrAttachSession AppendMode = "session"
)

type Provider interface {
	Name() string
	Check() bool
	Order() int
	Run(option *recon.Option, opts Opts) error
}
