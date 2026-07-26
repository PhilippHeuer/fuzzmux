package cmd

import (
	"fmt"
	"os"

	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func utilCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "util",
		Short: "Utility commands for querying window manager state",
	}

	cmd.AddCommand(utilWmCmd())
	cmd.AddCommand(utilFocusedPidCmd())
	cmd.AddCommand(utilFocusedCwdCmd())
	cmd.AddCommand(utilFocusedKillCmd())

	return cmd
}

func utilWmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wm",
		Short: "Print the detected window manager slug",
		Run: func(cmd *cobra.Command, args []string) {
			be, err := app.FindLauncher("", config.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("no suitable launcher found")
			}

			fmt.Println(be.Name())
		},
	}
}

func utilFocusedPidCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focused-pid",
		Short: "Print the PID of the currently focused window",
		Run: func(cmd *cobra.Command, args []string) {
			be, err := app.FindLauncher("", config.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("no suitable launcher found")
			}

			pid, err := be.FocusedPID()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get focused PID")
			}

			fmt.Println(pid)
		},
	}
}

func utilFocusedCwdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focused-cwd",
		Short: "Print the working directory of the focused window's process",
		Run: func(cmd *cobra.Command, args []string) {
			be, err := app.FindLauncher("", config.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("no suitable launcher found")
			}

			pid, err := be.FocusedPID()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get focused PID")
			}

			cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
			if err != nil {
				log.Fatal().Err(err).Int("pid", pid).Msg("failed to read process cwd")
			}

			fmt.Println(cwd)
		},
	}
}

func utilFocusedKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focused-kill",
		Short: "Terminate the currently focused window",
		Run: func(cmd *cobra.Command, args []string) {
			be, err := app.FindLauncher("", config.Config{})
			if err != nil {
				log.Fatal().Err(err).Msg("no suitable launcher found")
			}

			pid, err := be.FocusedPID()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get focused PID")
			}

			err = util.KillProcessByPID(pid)
			if err != nil {
				log.Fatal().Err(err).Int("pid", pid).Msg("failed to kill process")
			}
		},
	}
}
