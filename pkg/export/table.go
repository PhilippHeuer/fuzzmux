package export

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"slices"
)

// GenerateOptionTable generates a table of options with the given columns
func GenerateOptionTable(options []recon.Option, columns []recon.Column) ([]string, []map[string]string) {
	var headers []string
	for _, col := range columns {
		headers = util.AddToSet(headers, col.Key)
	}

	var rows []map[string]string
	for _, option := range options {
		row := make(map[string]string)
		for _, column := range columns {
			switch column.Key {
			case "module":
				row[column.Key] = option.ProviderName
			case "id":
				row[column.Key] = option.Id
			case "name":
				row[column.Key] = option.Name
			case "display_name":
				row[column.Key] = option.DisplayName
			case "directory":
				row[column.Key] = option.ResolveStartDirectory(true)
			default:
				row[column.Key] = option.Context[column.Key]
			}
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
