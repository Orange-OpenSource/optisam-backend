package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func buildQueryIPS(metric *v1.MetricIPSComputed, scopes []string, id ...string) string {
	q := `
{
	var(func:uid($ID)){
		product.equipment @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])) {
		   cn as equipment.$BaseType.$NumCores
		   cf as equipment.$BaseType.$CoreFactor
		   cpu as equipment.$BaseType.$NumCPU
		   comp as  math (cn*cf*cpu)
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
		"$NumCPU":     metric.NumCPUAttr.Name,
		"$CoreFactor": metric.CoreFactorAttr.Name,
		"$Scopes":     strings.Join(scopes, ","),
	})
}
