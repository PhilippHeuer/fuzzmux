package export

import (
	"slices"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
)

// GenerateOptionTable generates a table of options with the given columns
func GenerateOptionTable(options []recon.Option, columns []recon.Column) ([]string, [][]interface{}) {
	var headers []string
	for _, col := range columns {
		headers = util.AddToSet(headers, col.Key)
	}

	var rows [][]interface{}
	for _, option := range options {
		var row []interface{}
		for _, column := range columns {
			var value interface{}
			switch column.Key {
			case "module":
				value = option.ProviderName
			case "id":
				value = option.Id
			case "name":
				value = option.Name
			case "display_name":
				value = option.DisplayName
			case "directory":
				value = option.ResolveStartDirectory(true)
			default:
				value = option.Context[column.Key]
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}

	return headers, rows
}

// GenerateColumns returns the expected output columns, filtered by name if outputColumns is set
// if outputColumns is not set, all columns are returned (except hidden columns)
func GenerateColumns(modules []recon.Module, outputColumns []string) []recon.Column {
	var columns []recon.Column
	var addedKeys []string

	for _, p := range modules {
		if outputColumns != nil && len(outputColumns) > 0 {
			for _, col := range p.Columns() {
				for _, colName := range outputColumns {
					if col.Name == colName || col.Key == colName {
						if slices.Contains(addedKeys, col.Key) {
							continue
						}

						addedKeys = util.AddToSet(addedKeys, col.Key)
						columns = append(columns, col)
					}
				}
			}
			continue
		} else {
			for _, col := range p.Columns() {
				if slices.Contains(addedKeys, col.Key) || col.Hidden {
					continue
				}

				addedKeys = util.AddToSet(addedKeys, col.Key)
				columns = append(columns, col)
			}
		}
	}

	return columns
}
