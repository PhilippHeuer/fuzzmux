package recon

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func AttributeMapping(attributes map[string]interface{}, contextMapping []types.FieldMapping) map[string]string {
	context := make(map[string]string)

	for _, mapping := range contextMapping {
		// require source and target
		if mapping.Source == "" || mapping.Target == "" {
			continue
		}

		// get value
		value, ok := attributes[mapping.Source]
		if !ok {
			continue
		}

		// map attributes
		valueStr, err := formatAttributeValue(value, mapping.Format, mapping.Source)
		if err != nil {
			log.Warn().Str("source", mapping.Source).Str("format", mapping.Format).Msg(err.Error())
			continue
		}
		context[mapping.Target] = valueStr
	}

	return context
}

func formatAttributeValue(value interface{}, format, source string) (string, error) {
	switch v := value.(type) {
	case string:
		return formatString(v, format)
	case []string:
		return formatStringSlice(v, format)
	case int64:
		return formatInt64(v, format)
	case *int64:
		if v != nil {
			return formatInt64(*v, format)
		}
	case time.Time:
		return v.Format(time.RFC3339), nil
	case *time.Time:
		if v != nil {
			return v.Format(time.RFC3339), nil
		}
	}
	return "", fmt.Errorf("unsupported type for source %s with format %s", source, format)
}

func formatString(value, format string) (string, error) {
	switch format {
	case "ldaptime":
		return util.ConvertLDAPTimeToRFC3339(value), nil
	default:
		return value, nil
	}
}

func formatStringSlice(values []string, format string) (string, error) {
	switch format {
	case "ldaptime":
		return util.ConvertLDAPTimeToRFC3339(values[0]), nil
	default:
		return strings.Join(values, ", "), nil
	}
}

func formatInt64(value int64, format string) (string, error) {
	switch format {
	case "unixmillis":
		return util.ConvertMilliUnixTimestampToRFC3339(&value), nil
	default:
		return fmt.Sprintf("%d", value), nil
	}
}
