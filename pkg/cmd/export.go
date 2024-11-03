package cmd

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/export"
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
			modules, options := app.GatherReconOptions(conf, moduleNames, flags.showTags, flags.hideTags, flags.maxCacheAge)
			if len(options) == 0 {
				log.Fatal().Msg("no options found")
			}

			// generate table
			var columns = export.GenerateColumns(modules, outputColumns)
			header, rows := export.GenerateOptionTable(options, columns)

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
				export.RenderAsTSV(header, rows, writer)
			} else if outputFormat == "csv" {
				err = export.RenderAsCSV(header, rows, writer)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to generate CSV")
				}
			} else if outputFormat == "json" {
				export.RenderAsJSON(rows, writer)
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
