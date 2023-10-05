package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {

		// query sessions
		//_, _, _, err := shell.ExecuteCommand("tmux", "list-sessions", "-F", "#{window_id},#{window_name")
		//if err != nil {
		//	return
		//}

		/*
			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tURI")
			for _, feed := range cfg.Feeds {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", feed.Name, feed.URL)
			}
			_ = w.Flush()

		*/
	},
}
