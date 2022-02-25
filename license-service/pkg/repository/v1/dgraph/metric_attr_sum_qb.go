package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

func buildQueryAttrSum(metric *v1.MetricAttrSumStandComputed, scopes []string, id ...string) string {
	q := `{
		var(func:uid($ID)){
			product.equipment @filter(has(equipment.$BaseType.$AttrName) AND eq(scopes,[$Scopes])){
				values as equipment.$BaseType.$AttrName
			}
			totalSum as sum(val(values))
		}
		Licenses(func: uid(totalSum)){
			LicensesNoCeil: val(totalSum)  
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":       strings.Join(id, ","),
		"$BaseType": metric.BaseType.Type,
		"$AttrName": metric.Attribute.Name,
		"$Scopes":   strings.Join(scopes, ","),
	})
}
