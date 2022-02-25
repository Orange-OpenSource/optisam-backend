package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

func buildQueryIPS(metric *v1.MetricIPSComputed, scopes []string, id ...string) string {
	q := `
{
	var(func:uid($ID)){
		product.equipment @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])) {
		   cn as equipment.$BaseType.$NumCores
		   cf as equipment.$BaseType.$CoreFactor
		   comp as  math (cn*cf)
		}
	}
	Licenses(){
		Licenses: sum(val(comp))
	}
}
   `
	return replacer(q, map[string]string{
		"$ID":         strings.Join(id, ","),
		"$BaseType":   metric.BaseType.Type,
		"$NumCores":   metric.NumCoresAttr.Name,
		"$CoreFactor": metric.CoreFactorAttr.Name,
		"$Scopes":     strings.Join(scopes, ","),
	})
}
