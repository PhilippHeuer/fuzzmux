package sway

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"os"
	"strings"
	"time"

	"github.com/joshuarubin/go-sway"
	"github.com/rs/zerolog/log"
)

type SWAY struct {
}

func (p SWAY) Name() string {
	return "sway"
}

func (p SWAY) Check() bool {
	sock := os.Getenv("SWAYSOCK")
	if sock == "" {
		return false
	}
	_, err := os.Stat(sock)
	return err == nil
}

func (p SWAY) Order() int {
	return 201
}

func (p SWAY) Run(option *recon.Option, opts launcher.Opts) error {
	// resolve vars
	startDirectory := option.ResolveStartDirectory(true)

	// context, the sway client closes its connection once the context is done, so it must cover the whole launch
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// sway client
	client, err := sway.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to sway")
	}
	log.Debug().Msg("connected to sway ipc")

	// get sway workspace
	ws, err := currentSwayWorkspace(ctx, client)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get focused workspace")
	}
	log.Debug().Int64("id", ws.ID).Msg("active workspace")

	// kill all active windows in workspace
	if opts.Layout.ClearWorkspace {
		log.Debug().Msg("clearing current workspace")
		for _, node := range swayWindows(ws) {
			if node.PID != nil && util.IsAncestorProcess(int(*node.PID)) {
				log.Debug().Int64("id", node.ID).Msg("skipping window running tmx")
				continue
			}
			log.Trace().Int64("id", node.ID).Msg("killing window by nodeId")
			_, killErr := client.RunCommand(ctx, fmt.Sprintf("[con_id=%d] kill", node.ID))
			if killErr != nil {
				log.Warn().Err(killErr).Msg("failed to kill window")
			}
		}
	}

	// start apps
	for _, app := range opts.Layout.Apps {
		log.Debug().Str("name", app.Name).Msg("starting app")

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
			cmd, err = util.GetTerminalCommand(os.Getenv("TERM"), startDirectory, option.ResolvePlaceholders(script.String()))
			if err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to prepare command to start app")
			}
		}
		log.Trace().Str("name", app.Name).Str("cmd", cmd).Msg("started app")

		// execute command
		// sway keeps quotes for exec and only splits commands on ";" outside of quotes
		_, cmdErr := client.RunCommand(ctx, fmt.Sprintf("exec cd %s && %s", util.ShellQuote(startDirectory), cmd))
		if cmdErr != nil {
			log.Fatal().Err(cmdErr).Str("name", app.Name).Msg("failed to start app")
		}
	}

	return nil
}

func (p SWAY) FocusedPID() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client, err := sway.New(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to sway: %w", err)
	}

	tree, err := client.GetTree(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get sway tree: %w", err)
	}

	focused := tree.FocusedNode()
	if focused == nil || focused.PID == nil {
		return 0, fmt.Errorf("no focused window found")
	}

	return int(*focused.PID), nil
}

func currentSwayWorkspace(ctx context.Context, client sway.Client) (*sway.Node, error) {
	// get current tree and workspace
	n, err := client.GetTree(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sway tree: %w", err)
	}
	if len(n.Focus) == 0 {
		return nil, fmt.Errorf("failed to get focused workspace: no focused output")
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
		}
		if len(node.Focus) == 0 {
			break
		}
		return focusedWorkspace(node, node.Focus[0])
	}

	return nil, fmt.Errorf("workspace not found")
}

// swayWindows returns all windows (leaf containers, tiled and floating) within the node
func swayWindows(n *sway.Node) []*sway.Node {
	var result []*sway.Node
	for _, child := range append(append([]*sway.Node{}, n.Nodes...), n.FloatingNodes...) {
		if len(child.Nodes) == 0 && len(child.FloatingNodes) == 0 {
			if child.Type == sway.NodeCon || child.Type == sway.NodeFloatingCon {
				result = append(result, child)
			}
			continue
		}
		result = append(result, swayWindows(child)...)
	}
	return result
}
