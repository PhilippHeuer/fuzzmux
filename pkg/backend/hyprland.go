package backend

import (
	"fmt"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/core/util"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	hyprclient "github.com/labi-le/hyprland-ipc-client/v3"
	"github.com/rs/zerolog/log"
)

type Hyprland struct {
}

func (p Hyprland) Name() string {
	return "hyprland"
}

func (p Hyprland) Check() bool {
	_, ok := os.LookupEnv("HYPRLAND_INSTANCE_SIGNATURE")
	return ok
}

func (p Hyprland) Order() int {
	return 201
}

func (p Hyprland) Run(option *provider.Option, opts Opts) error {
	// resolve vars
	startDirectory := option.ResolveStartDirectory(true)

	// ipc client
	client := hyprclient.MustClient(os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"))
	if client == nil {
		return fmt.Errorf("failed to connect to hyprland ipc socket")
	}
	log.Debug().Msg("connected to hyprland ipc")

	// active workspace
	ws, err := client.ActiveWorkspace()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get focused workspace")
	}
	log.Debug().Int("id", ws.Id).Msg("active workspace")

	// kill all active windows in workspace
	if opts.Layout.ClearWorkspace {
		log.Debug().Msg("clearing current workspace")

		clients, err := client.Clients()
		if err != nil {
			return err
		}

		for _, c := range clients {
			if c.Workspace.Id == ws.Id {
				log.Trace().Int("pid", c.Pid).Int("workspace", c.Workspace.Id).Msg("killing process")
				err := util.KillProcessByPID(c.Pid)
				if err != nil {
					return err
				}
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
			cmd, err = getTerminalCommand(os.Getenv("TERM"), startDirectory, option.ResolvePlaceholders(script.String()))
			if err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to prepare command to start app")
			}
		}
		log.Trace().Str("name", app.Name).Str("cmd", cmd).Msg("started app")

		// execute command
		cmdErr := hyprlandIPCCommand(client, fmt.Sprintf("exec cd %q && %s", startDirectory, cmd))
		if cmdErr != nil {
			log.Fatal().Err(cmdErr).Str("name", app.Name).Msg("failed to start app")
		}
	}

	return nil
}

func hyprlandIPCCommand(client *hyprclient.IPCClient, command string) error {
	q := hyprclient.NewByteQueue()
	q.Add([]byte(command))

	_, err := client.Dispatch(q)
	return err
}
