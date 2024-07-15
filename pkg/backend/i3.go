package backend

import (
	"fmt"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	"github.com/rs/zerolog/log"
	"go.i3wm.org/i3/v4"
)

type I3 struct {
}

func (p I3) Name() string {
	return "i3"
}

func (p I3) Check() bool {
	_, ok := os.LookupEnv("I3SOCK")
	return ok
}

func (p I3) Order() int {
	return 200
}

func (p I3) Run(option *provider.Option, opts Opts) error {
	// start directory
	startDirectory := option.ResolveStartDirectory(true)

	// get sway workspace
	ws, err := currentI3Workspace()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get focused workspace")
	}

	// kill all active windows in workspace
	if opts.Layout.ClearWorkspace {
		for _, node := range ws.Nodes {
			if node.Type == "con" {
				_, killErr := i3.RunCommand(fmt.Sprintf("[con_id=%d] kill", node.ID))
				if killErr != nil {
					log.Warn().Err(killErr).Msg("failed to kill window")
				}
			}
		}
	}

	// start apps
	for _, app := range opts.Layout.Apps {
		// launch script
		script := strings.Builder{}
		for _, cmd := range app.Commands {
			script.WriteString(cmd.Command)
			script.WriteString("; ")
		}
		if script.Len() == 0 {
			script.WriteString("${SHELL}")
		}

		// start app
		var cmd string
		if app.GUI && len(app.Commands) == 1 {
			cmd = option.ResolvePlaceholders(app.Commands[0].Command)
		} else {
			cmd, err = getTerminalCommand(os.Getenv("TERM"), startDirectory, option.ResolvePlaceholders(script.String()))
			if err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to prepare command to start app")
			}
		}

		_, cmdErr := i3.RunCommand(fmt.Sprintf("exec cd %q && %s", startDirectory, cmd))
		if cmdErr != nil {
			log.Fatal().Err(cmdErr).Str("name", app.Name).Msg("failed to start app")
		}
	}

	return nil
}

func currentI3Workspace() (*i3.Node, error) {
	// get current tree and workspace
	n, err := i3.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get sway tree: %w", err)
	}
	ws, err := focusedI3Workspace(n.Root, n.Root.Focus[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get focused workspace: %w", err)
	}

	return ws, nil
}

// focusedWorkspace returns the focused workspace
func focusedI3Workspace(n *i3.Node, targetID i3.NodeID) (*i3.Node, error) {
	for _, node := range n.Nodes {
		if node.ID != targetID {
			continue
		}

		if node.Type == i3.WorkspaceNode {
			return node, nil
		} else {
			return focusedI3Workspace(node, node.Focus[0])
		}
	}

	return nil, fmt.Errorf("workspace not found")
}
