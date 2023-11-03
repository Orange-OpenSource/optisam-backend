package dgraph

import (
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

const (
	qApplicationProduct = `
	var(func: eq(application.id,$appID))@filter(eq(scopes,[$Scopes])) {
		app_inst as  application.instance
	   }
	var(func: uid($id)) {
		prod_inst as ~instance.product@filter(uid(app_inst))
	   }
	var(func: uid(prod_inst)) {
		ins_equip as instance.equipment
	   }
	`
	qDirectEquipmentsForAppProduct = `
	var(func: uid($id)){
		product.equipment @filter(eq(equipment.type,$CurrentType) AND uid(ins_equip) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs as uid
	   }
	}`
	qDirectEquipments = `var(func: uid($id)){
		product.equipment @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes]) $CurTypeCondition){
			$CurrentTypeIDs as uid
	   }
	}`

	qEquipmentFromChild = `var (func: uid($id)){
		equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs_c as uid
		}  
	}`
	qParentEquipment = `var (func: uid($id)){
		equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){
			$CurrentTypeIDs as uid
		}  
	}`

	qBaseCalulation = `~equipment.parent @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])){
			$BaseType_p_$CurrentType as  uid 
			cpu_$CurrentType as equipment.$BaseType.$AttrNumCPU
			cores_$CurrentType as equipment.$BaseType.$AttrNumCores
			coreFactor_$CurrentType as equipment.$BaseType.$AttrCoreFactor
			$BaseType_t_$CurrentType as math(cpu_$CurrentType*cores_$CurrentType*coreFactor_$CurrentType)  
		}`

	qBaseCalulationCeil = `~equipment.parent @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])){
			$BaseType_p_$CurrentType as  uid 
			cpu_$CurrentType as equipment.$BaseType.$AttrNumCPU
			cores_$CurrentType as equipment.$BaseType.$AttrNumCores
			coreFactor_$CurrentType as equipment.$BaseType.$AttrCoreFactor
			$BaseType_t_$CurrentType as math(ceil (cpu_$CurrentType*cores_$CurrentType*coreFactor_$CurrentType))  
	    }`

	qCalculateLicenses = `var(func: uid($CurrentTypeIDs))$Filter{
        $Query
		$CurrentType_t as sum(val($ChildType_t_$CurrentType))
		$RoundOff
	}`

	qCalculateLicensesCeil = `var(func: uid($CurrentTypeIDs))$Filter{
        $Query
		$CurrentType_t as sum(val($ChildType_t_$CurrentType_ceil))
		$RoundOff
	}`

	qRoundOff = `$VAR_ceil as math(ceil $VAR)  `

	qCalculateLicensesTraverseChildren = `~equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){ 
		$CurrentType_p_$ParentType as uid
		$Query
		$CurrentType_t_$ParentType as sum(val($ChildType_t_$ParentType))
		$RoundOff      
	}`

	qCalculateLicensesTraverseChildrenCeil = `~equipment.parent @filter(eq(equipment.type,$CurrentType) AND eq(scopes,[$Scopes])){ 
		$CurrentType_p_$ParentType as uid
		$Query
		$CurrentType_t_$ParentType as sum(val($ChildType_t_$ParentType_ceil))
		$RoundOff      
	}`

	qLicenses = `
	Licenses()@normalize{
		$Query
		Licenses:math($Aggregations)
	}
	`
	qLevelAggregation = `l_$CurrentType as sum(val($CurrentType_t))`

	qLevelAggregationCeil = `l_$CurrentType as sum(val($CurrentType_t_ceil))`

	qJustBaseCalculation = `
	var(func: uid($id$))$Filter{
		cpu_$BaseType as equipment.$BaseType.$AttrNumCPU
		cores_$BaseType as equipment.$BaseType.$AttrNumCores
		coreFactor_$BaseType as equipment.$BaseType.$AttrCoreFactor
		$BaseType_t as math(ceil cpu_$BaseType*cores_$BaseType*coreFactor_$BaseType)
	}`
)

// type opsMatrix struct {
// 	EqTypeTree     []*v1.EquipmentType
// 	BaseType       *v1.EquipmentType
// 	AggregateLevel *v1.EquipmentType
// 	CPU            *v1.Attribute
// 	Cores          *v1.Attribute
// 	CoreFactor     *v1.Attribute
// }

func replacer(q string, params map[string]string) string {
	for key, val := range params {
		q = strings.Replace(q, key, val, -1) // nolint: gocritic
	}
	return q
}

func queryBuilder(ops *v1.MetricOPSComputed, scopes []string, allotedMetricsEq map[string]interface{}, id ...string) string {
	// q := ""
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

func queryBuilderForAppProduct(ops *v1.MetricOPSComputed, appID string, scopes []string, prodID string) string {
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
		getAppProduct(appID),
		getToBaseForAppProduct(ops.EqTypeTree[:index+1]),
		getToTop(ops.EqTypeTree[index:], index > 0),
		caluclateFromTop(ops.EqTypeTree, ops.CoreFactorAttr, ops.NumCPUAttr, ops.NumCoresAttr, aggregateIndex-index, index),
		licenses(ops.EqTypeTree[index:], aggregateIndex-index),
	}, "\n\t"), map[string]string{"$id": prodID, "$Scopes": strings.Join(scopes, ",")}) + "\n}"
}

func getAppProduct(appID string) string {
	return replacer(qApplicationProduct, map[string]string{
		"$appID": appID,
	})
}
func getToBaseForAppProduct(eqTypes []*v1.EquipmentType) string {
	queries := []string{}
	for i := range eqTypes {
		if i == 0 {
			vars := map[string]string{
				"$CurrentType": eqTypes[i].Type,
			}
			queries = append(queries, replacer(qDirectEquipmentsForAppProduct, vars))
			continue
		}
		ids := []string{eqTypes[i-1].Type + "IDs"}
		if i > 1 {
			ids = append(ids, eqTypes[i-1].Type+"IDs_c")
		}
		vars := map[string]string{
			"$id":          strings.Join(ids, ","),
			"$CurrentType": eqTypes[i].Type,
		}
		queries = append(queries, replacer(qEquipmentFromChild, vars))
		vars = map[string]string{
			"$CurrentType": eqTypes[i].Type,
		}
		queries = append(queries, replacer(qDirectEquipmentsForAppProduct, vars))
	}
	return strings.Join(queries, "\n\t")
}

func getToBase(eqTypes []*v1.EquipmentType, allotedMetricsEq map[string]interface{}) string {
	queries := []string{}
	for i := range eqTypes {
		currentTypeCondition := ""
		if i == 0 {
			if eqTypes[i].Type != "server" {
				if _, ok := allotedMetricsEq["notAllotedVirtualMachine"]; ok {
					if allotedMetricsEq["notAllotedVirtualMachine"].(string) != "" {
						currentTypeCondition = " AND not( uid(" + allotedMetricsEq["notAllotedVirtualMachine"].(string) + "))"
					}
				}
			} else {
				if _, ok := allotedMetricsEq["notAlloted"]; ok {
					if allotedMetricsEq["notAlloted"].(string) != "" {
						currentTypeCondition = " AND not( uid(" + allotedMetricsEq["notAlloted"].(string) + "))"
					}
				}
			}
			vars := map[string]string{
				"$CurrentType":      eqTypes[i].Type,
				"$CurTypeCondition": currentTypeCondition,
			}
			queries = append(queries, replacer(qDirectEquipments, vars))
			currentTypeCondition = ""
			continue
		}
		ids := []string{eqTypes[i-1].Type + "IDs"}
		if i > 1 {
			ids = append(ids, eqTypes[i-1].Type+"IDs_c")
		}
		vars := map[string]string{
			"$id":          strings.Join(ids, ","),
			"$CurrentType": eqTypes[i].Type,
		}
		queries = append(queries, replacer(qEquipmentFromChild, vars))

		// if _, ok := allotedMetricsEq["alloted"]; ok {
		// 	if allotedMetricsEq["alloted"].(string) != "" {
		// 		currentTypeCondition = " AND (" + allotedMetricsEq["alloted"].(string) + ")"
		// 	}
		// }
		//&& allotedMetricsEq["alloted"].(string) == ""
		if _, ok := allotedMetricsEq["notAlloted"]; ok {
			if allotedMetricsEq["notAlloted"].(string) != "" {
				currentTypeCondition = " AND not( uid(" + allotedMetricsEq["notAlloted"].(string) + "))"
			}
		}
		vars = map[string]string{
			"$CurrentType":      eqTypes[i].Type,
			"$CurTypeCondition": currentTypeCondition,
		}
		queries = append(queries, replacer(qDirectEquipments, vars))
	}
	return strings.Join(queries, "\n\t")
}

func getToTop(eqTypes []*v1.EquipmentType, baseHasChilds bool) string {
	queries := []string{}
	var childIDs []string
	for i := range eqTypes {
		if i == 0 {
			childIDs = append(childIDs, eqTypes[i].Type+"IDs")
			if i == 0 && baseHasChilds {
				childIDs = append(childIDs, eqTypes[i].Type+"IDs_c")
			}
			continue
		}
		vars := map[string]string{
			"$id":          strings.Join(childIDs, ","),
			"$CurrentType": eqTypes[i].Type,
		}
		queries = append(queries, replacer(qParentEquipment, vars))
		childIDs = []string{
			eqTypes[i].Type + "IDs",
		}
	}
	return strings.Join(queries, "\n\t")
}

func getUIDFilter(currentType string, uids ...string) string {
	if len(uids) == 0 {
		return ""
	}
	newUids := make([]string, len(uids))
	for i := range uids {
		newUids[i] = currentType + "_p_" + uids[i]
	}
	return `@filter( NOT uid(` + strings.Join(newUids, ",") + `))`
}

func caluclateFromTop(eqTypesAll []*v1.EquipmentType, cf, cpu, cores *v1.Attribute, agIdx, baseIdx int) string {
	queries := []string{}
	filterIDs := []string{}
	eqTypes := eqTypesAll[baseIdx:]
	for i := len(eqTypes) - 1; i >= 0; i-- {
		vars := map[string]string{
			"$CurrentType": eqTypes[i].Type,
			"$ChildType":   "",
			"$Filter":      getUIDFilter(eqTypes[i].Type, filterIDs...),
			"$RoundOff":    "",
		}

		if i <= agIdx {
			vars["$RoundOff"] = replacer(qRoundOff, map[string]string{
				"$VAR": eqTypes[i].Type + "_t",
			})
		}

		cl := qCalculateLicenses
		if i == agIdx+1 && agIdx != 0 {
			cl = qCalculateLicensesCeil
		}

		if i == 0 {
			q := replacer(qJustBaseCalculation, map[string]string{
				"$id$": func() string {
					if baseIdx > 0 {
						return strings.Join([]string{eqTypes[i].Type + "IDs", eqTypes[i].Type + "IDs_c"}, ",")
					}
					return eqTypes[i].Type + "IDs"
				}(),
				"$BaseType":       eqTypes[i].Type,
				"$Filter":         getUIDFilter(eqTypes[i].Type, filterIDs...),
				"$AttrNumCPU":     cpu.Name,
				"$AttrNumCores":   cores.Name,
				"$AttrCoreFactor": cf.Name,
				"$RoundOff":       "",
			})
			queries = append(queries, q)
			continue
		}
		filterIDs = append(filterIDs, eqTypes[i].Type)
		vars["$ChildType"] = eqTypes[i-1].Type

		q := replacer(cl, vars)
		for j := i - 1; j >= 0; j-- {
			if j == 0 {
				base := qBaseCalulation
				if agIdx == j {
					base = qBaseCalulationCeil
				}
				q = replacer(q, map[string]string{
					"$Query": replacer(base, map[string]string{
						"$CurrentType":    eqTypes[i].Type,
						"$BaseType":       eqTypes[j].Type,
						"$AttrNumCPU":     cpu.Name,
						"$AttrNumCores":   cores.Name,
						"$AttrCoreFactor": cf.Name,
					}),
				})
				continue
			}

			varsCC := map[string]string{
				"$CurrentType": eqTypes[j].Type,
				"$ParentType":  eqTypes[i].Type,
				"$ChildType":   eqTypes[j-1].Type,
				"$RoundOff":    "",
			}

			if j == agIdx {
				varsCC["$RoundOff"] = replacer(qRoundOff, map[string]string{
					"$VAR": eqTypes[j].Type + "_t_" + eqTypes[i].Type,
				})
			}
			cl := qCalculateLicensesTraverseChildren
			if j == agIdx+1 {
				cl = qCalculateLicensesTraverseChildrenCeil
			}

			q = replacer(q, map[string]string{
				"$Query": replacer(cl, varsCC),
			})
		}
		queries = append(queries, q)
	}
	return strings.Join(queries, "\n\t")
}

func licenses(eqTypes []*v1.EquipmentType, agIdx int) string {
	queries := []string{}
	types := []string{}
	for i := range eqTypes {
		types = append(types, "l_"+eqTypes[i].Type)
		la := qLevelAggregation
		if i <= agIdx && i > 0 {
			la = qLevelAggregationCeil
		}
		queries = append(queries, replacer(la, map[string]string{
			"$CurrentType": eqTypes[i].Type,
		}))
	}
	return replacer(qLicenses, map[string]string{
		"$Query":        strings.Join(queries, "\n\t\t"),
		"$Aggregations": strings.Join(types, "+"),
	})
}
