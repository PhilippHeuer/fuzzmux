package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Version will be set at build time
var Version string

// RepositoryStatus will be set at build time
var RepositoryStatus string

// CommitHash will be set at build time
var CommitHash string

// BuildAt will be set at build time
var BuildAt string

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "print version information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(os.Stdout, "GitVersion:    %s\n", Version)
			_, _ = fmt.Fprintf(os.Stdout, "GitCommit:     %s\n", CommitHash)
			_, _ = fmt.Fprintf(os.Stdout, "GitTreeState:  %s\n", RepositoryStatus)
			_, _ = fmt.Fprintf(os.Stdout, "BuildDate:     %s\n", BuildAt)
			_, _ = fmt.Fprintf(os.Stdout, "GoVersion:     %s\n", runtime.Version())
			_, _ = fmt.Fprintf(os.Stdout, "Compiler:      %s\n", runtime.Compiler)
			_, _ = fmt.Fprintf(os.Stdout, "Platform:      %s\n", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}

	return cmd
}
