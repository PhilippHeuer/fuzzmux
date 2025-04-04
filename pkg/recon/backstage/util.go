package backstage

import (
	"fmt"
	"github.com/datolabs-io/go-backstage/v3"
)

func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%s", val)
	}
	return ""
}

func entityRelationToString(data []backstage.EntityRelation, relationType string) string {
	for _, relation := range data {
		if relation.Type == relationType {
			return relation.Target.Namespace + "/" + relation.Target.Kind + "/" + relation.Target.Name
		}
	}
	return ""
}
