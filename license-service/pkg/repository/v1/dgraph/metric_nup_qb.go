package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

// func queryBuilderNUP(mat *v1.MetricNUPComputed, templ *template.Template, id ...string) (string, error) {
// 	buf := &bytes.Buffer{}
// 	if err := templ.Execute(buf, mat); err != nil {
// 		return "", err
// 	}

// 	return formatter(replacer(buf.String(), map[string]string{
// 		"$ID": strings.Join(id, ","),
// 	})), nil
// }

func queryBuilderOPSForNUP(ops *v1.MetricNUPComputed, scopes []string, allotedMetricsEq map[string]interface{}, id ...string) string {
	index := -1
	aggregateIndex := -1
	for i := range ops.EqTypeTree {
		if ops.EqTypeTree[i].Type == ops.BaseType.Type {
			index = i
		}
		if ops.EqTypeTree[i].Type == ops.AggregateLevel.Type {
			aggregateIndex = i
		}
	}

	return "{\n\t" + replacer(strings.Join([]string{
		getToBase(ops.EqTypeTree[:index+1], allotedMetricsEq),
		getToTop(ops.EqTypeTree[index:], index > 0),
		caluclateFromTop(ops.EqTypeTree, ops.CoreFactorAttr, ops.NumCPUAttr, ops.NumCoresAttr, aggregateIndex-index, index),
		licenses(ops.EqTypeTree[index:], aggregateIndex-index),
	}, "\n\t"), map[string]string{
		"$id":     strings.Join(id, ","),
		"$Scopes": strings.Join(scopes, ",")}) + "\n}"
}

func buildQueryUsersForNUP(scopes []string, EquipmentIDs string, id ...string) string {
	equipmentIdsCondition := `$EquipmentIDs`
	if EquipmentIDs != "" {
		equipmentIdsCondition = ` AND NOT eq(users.id,[$EquipmentIDs]) `
	}
	q := `{
		var(func:uid($ID)){
			product.users @filter(eq(scopes,[$Scopes]) ` + equipmentIdsCondition + `){
				uc as users.count
	 		}
		}
		Users(){
			TotalUserCount: sum(val(uc))
	  	}
	  }`
	return replacer(q, map[string]string{
		"$ID":           strings.Join(id, ","),
		"$Scopes":       strings.Join(scopes, ","),
		"$EquipmentIDs": strings.Replace(EquipmentIDs, "\\", "", -1),
	})
}
