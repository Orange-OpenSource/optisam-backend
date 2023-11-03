package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func buildQueryACS(metric *v1.MetricACSComputed, scopes []string, id ...string) string {
	q := `{
		var(func:uid($ID)){
			  attrCount as product.equipment @filter(eq(equipment.$BaseType.$AttrName,"$Value") AND eq(scopes,[$Scopes]))
		  }
		Licenses(func:uid(attrCount)){
		  Licenses:count(uid)
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":       strings.Join(id, ","),
		"$BaseType": metric.BaseType.Type,
		"$AttrName": metric.Attribute.Name,
		"$Value":    metric.Value,
		"$Scopes":   strings.Join(scopes, ","),
	})
}
