package app

import (
	"encoding/json"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"io"
	"text/tabwriter"
)

// RenderOptionsAsTSV prints the options as tab separated values to the writer using the tabwriter
func RenderOptionsAsTSV(options []recon.Option, columns []string, writer io.Writer) {
	w := tabwriter.NewWriter(writer, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, "MODULE\tID\tNAME")
	for _, option := range options {
		_, _ = fmt.Fprintln(w, ""+option.ProviderName+"\t"+option.Id+"\t"+option.DisplayName)
	}
	_ = w.Flush()
}

// RenderOptionsAsCSV prints the options in the CSV format
func RenderOptionsAsCSV(options []recon.Option, columns []string, writer io.Writer) {
	w := tabwriter.NewWriter(writer, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, "MODULE,ID,NAME")
	for _, option := range options {
		_, _ = fmt.Fprintln(w, ""+option.ProviderName+","+option.Id+","+option.DisplayName)
	}
	_ = w.Flush()
}

// RenderOptionsAsJson prints the options as JSON to the writer using the json encoder
func RenderOptionsAsJson(options []recon.Option, columns []string, writer io.Writer) {
	output, _ := json.MarshalIndent(options, "", " ")
	_, _ = writer.Write(output)
}
