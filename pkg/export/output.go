package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// RenderAsTSV prints data as TSV to the writer
func RenderAsTSV(headers []string, rows []map[string]string, writer io.Writer) {
	w := tabwriter.NewWriter(writer, 1, 1, 1, ' ', 0)
	// Write headers
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	// Write rows
	for _, row := range rows {
		fmt.Fprintln(w, joinRow(headers, row, "\t"))
	}
	w.Flush()
}

// RenderAsCSV prints data as CSV to the writer
func RenderAsCSV(headers []string, rows []map[string]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	err := csvWriter.Write(headers)
	if err != nil {
		return err
	}
	for _, row := range rows {
		err = csvWriter.Write(mapRow(headers, row))
		if err != nil {
			return err
		}
	}
	csvWriter.Flush()

	return csvWriter.Error()
}

// RenderAsJSON prints data as JSON to the writer
func RenderAsJSON(rows []map[string]string, writer io.Writer) {
	json.NewEncoder(writer).Encode(rows)
}

func joinRow(headers []string, row map[string]string, sep string) string {
	values := make([]string, len(headers))
	for i, header := range headers {
		values[i] = row[header]
	}
	return fmt.Sprint(strings.Join(values, sep))
}

func mapRow(headers []string, row map[string]string) []string {
	values := make([]string, len(headers))
	for i, header := range headers {
		values[i] = row[header]
	}
	return values
}
