package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func buildQueryEquipAttr(metric *v1.MetricEquipAttrStandComputed, scopes []string, id ...string) string {
	q := `{
		var(func:uid($ID)){
			product.equipment @filter(has(equipment.$BaseType.$AttrName) AND eq(scopes,[$Scopes]) ` + filterOnEnvironment(metric.EqTypeTree, strings.Split(metric.Environment, ",")) + `
				values as equipment.$BaseType.$AttrName
			}
  	 		totalSum as sum(val(values))
		}
		Licenses(){
			Licenses: sum(val(totalSum))
		}        
	  }`
	return replacer(q, map[string]string{
		"$ID":          strings.Join(id, ","),
		"$BaseType":    metric.BaseType.Type,
		"$Environment": metric.Environment,
		"$AttrName":    metric.Attribute.Name,
		"$Scopes":      strings.Join(scopes, ","),
	})
}

func filterOnEnvironment(eqTypes []*v1.EquipmentType, env []string) string {
	envs := []string{}
	for _, e := range env {
		envs = append(envs, strings.ToLower(strings.TrimSpace(e)))
	}
	switch len(eqTypes) {
	case 0:
		return ""
	case 1:
		return `AND eq(equipment.` + eqTypes[0].Type + `.environment,[` + strings.Join(envs, ",") + `])) @cascade{`
	default:
		query := ") @cascade{\n"
		for i, eq := range eqTypes[1:] {
			if i != len(eqTypes)-2 {
				query += "equipment.parent{\n"
			} else {
				query += `equipment.parent @filter(eq(equipment.` + eq.Type + `.environment,[` + strings.Join(envs, ",") + `]))`
			}
		}
		return query + "\n}"
	}

}
