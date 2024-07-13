package backend

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	"github.com/joshuarubin/go-sway"
	"github.com/rs/zerolog/log"
)

type SWAY struct {
}

func (p SWAY) Name() string {
	return "sway"
}

func (p SWAY) Check() bool {
	_, ok := os.LookupEnv("SWAYSOCK")
	return ok
}

func (p SWAY) Order() int {
	return 201
}

func (p SWAY) Run(option *provider.Option, opts Opts) error {
	// resolve vars
	startDirectory := os.ExpandEnv(option.StartDirectory)

	// context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// sway client
	client, err := sway.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to sway")
	}

	// get sway workspace
	ws, err := currentSwayWorkspace(ctx, client)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get focused workspace")
	}

	// kill all active windows in workspace
	if opts.Layout.ClearWorkspace {
		for _, node := range ws.Nodes {
			if node.Type == "con" {
				_, killErr := client.RunCommand(ctx, fmt.Sprintf("[con_id=%d] kill", node.ID))
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
			cmd = expandCommand(option, app.Commands[0].Command)
		} else {
			cmd, err = getTerminalCommand(os.Getenv("TERM"), startDirectory, expandCommand(option, script.String()))
			if err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to prepare command to start app")
			}
		}

		_, cmdErr := client.RunCommand(ctx, fmt.Sprintf("exec cd %q && %s", startDirectory, cmd))
		if cmdErr != nil {
			log.Fatal().Err(cmdErr).Str("name", app.Name).Msg("failed to start app")
		}
	}

	return nil
}

func currentSwayWorkspace(ctx context.Context, client sway.Client) (*sway.Node, error) {
	// get current tree and workspace
	n, err := client.GetTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sway tree: %w", err)
	}
	ws, err := focusedWorkspace(n, n.Focus[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get focused workspace: %w", err)
	}

	return ws, nil
}

// focusedWorkspace returns the focused workspace
func focusedWorkspace(n *sway.Node, targetID int64) (*sway.Node, error) {
	for _, node := range n.Nodes {
		if node.ID != targetID {
			continue
		}

		if node.Type == sway.NodeWorkspace {
			return node, nil
		} else {
			return focusedWorkspace(node, node.Focus[0])
		}
	}

	return nil, fmt.Errorf("workspace not found")
}
