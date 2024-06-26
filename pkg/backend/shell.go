package backend

import (
	"os"
	"strings"
	"syscall"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

type Shell struct {
}

func (p Shell) Name() string {
	return "simple"
}

func (p Shell) Check() bool {
	return true
}

func (p Shell) Run(option *provider.Option, opts Opts) error {
	// gather information
	startDirectory := os.ExpandEnv(option.StartDirectory)
	var commands []string
	for _, w := range opts.Layout.Apps {
		if w.Default && len(w.Commands) > 0 {
			for _, c := range config.CommandsAsStringSlice(w.Commands) {
				commands = append(commands, expandCommand(option, c))
			}
		}
	}

	// chdir
	err := os.Chdir(startDirectory)
	if err != nil {
		return err
	}

	// shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// exec
	if len(commands) > 0 {
		err = syscall.Exec(shell, append([]string{shell, "-c"}, strings.Join(commands, " && ")), os.Environ())
		if err != nil {
			return err
		}
	} else {
		err = syscall.Exec(shell, []string{shell}, os.Environ())
		if err != nil {
			return err
		}
	}

	return nil
}
