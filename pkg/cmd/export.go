package cmd

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func exportCmd() *cobra.Command {
	flags := RootFlags{}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "export option details to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			// params
			moduleNames, _ := cmd.Flags().GetStringArray("module")
			outputColumns, _ := cmd.Flags().GetStringArray("columns")
			outputFormat, _ := cmd.Flags().GetString("format")
			outputFile, _ := cmd.Flags().GetString("output")

			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// collect options
			providers := app.ConfigToReconModules(conf)
			var options []recon.Option
			options, errs := app.CollectOptions(providers, moduleNames, flags.maxCacheAge)
			if len(options) == 0 && len(errs) > 0 {
				log.Fatal().Errs("errors", errs).Msg("failed to collect options")
			} else if len(errs) > 0 {
				log.Warn().Errs("errors", errs).Msg("at least one recon failed to collect options")
			}
			options = app.FilterOptions(options, flags.showTags, flags.hideTags)
			if len(options) == 0 {
				log.Fatal().Msg("no options found")
			}

			// export options
			writer := cmd.OutOrStdout()
			if outputFile != "" {
				file, createErr := os.Create(outputFile) // Use os.Create to open (or create) the file for writing
				if createErr != nil {
					log.Fatal().Err(createErr).Msg("failed to open output file")
				}
				defer func() {
					if closeErr := file.Close(); closeErr != nil {
						log.Fatal().Err(closeErr).Msg("failed to close output file")
					}
				}()
				writer = file
			}
			if outputFormat == "tsv" {
				app.RenderOptionsAsTSV(options, outputColumns, writer)
			} else if outputFormat == "csv" {
				app.RenderOptionsAsCSV(options, outputColumns, writer)
			} else if outputFormat == "json" {
				app.RenderOptionsAsJson(options, outputColumns, writer)
			} else {
				log.Fatal().Str("format", outputFormat).Msg("invalid format")
			}
		},
	}
	cmd.PersistentFlags().StringArrayP("module", "m", []string{}, "modules to collect options from, empty for all")
	cmd.PersistentFlags().StringArrayP("columns", "c", []string{}, "columns to include in output, default: all") // TODO: implement
	cmd.PersistentFlags().StringP("format", "f", "tsv", "output format, valid: csv, tsv, json")
	cmd.PersistentFlags().StringP("output", "o", "", "output file, empty for stdout")
	return cmd
}
