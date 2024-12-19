package cmd

import (
	"fmt"
	"os"

	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/export"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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

			// generate data
			columns := export.GenerateColumns(modules, outputColumns)
			header, rows := export.GenerateOptionTable(options, columns)

			// data
			data := clioutputwriter.TabularData{
				Headers: header,
				Rows:    [][]interface{}{},
			}
			for _, row := range rows {
				data.Rows = append(data.Rows, row)
			}

			// filter columns
			if len(outputColumns) > 0 {
				data = clioutputwriter.FilterColumns(data, outputColumns)
			}

			// print
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
			err = clioutputwriter.PrintData(writer, data, clioutputwriter.Format(outputFormat))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to print data")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringArrayP("module", "m", []string{}, "modules to collect options from, empty for all")
	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))
	cmd.Flags().StringSliceP("columns", "c", []string{}, "columns to display")
	cmd.Flags().StringP("output", "o", "", "output file, empty for stdout")

	return cmd
}
