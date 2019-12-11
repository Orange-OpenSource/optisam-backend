// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package querybuilders

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

const (
	qDirectEquipments = `var(func: uid($id)){
		product.equipment @filter(eq(equipment.type,$CurrentType)){
			$CurrentTypeIDs as uid
	   }
	}`

	qEquipmentFromChild = `var (func: uid($id)){
		equipment.parent @filter(eq(equipment.type,$CurrentType)){
			$CurrentTypeIDs_c as uid
		}  
	}`

	qParentEquipment = `var (func: uid($id)){
		equipment.parent @filter(eq(equipment.type,$CurrentType)){
			$CurrentTypeIDs as uid
		}  
	}`

	qBaseCalulation = `~equipment.parent @filter(eq(equipment.type,$BaseType)){
			$BaseType_p_$CurrentType as  uid 
			cpu_$CurrentType as equipment.$BaseType.$AttrNumCPU
			cores_$CurrentType as equipment.$BaseType.$AttrNumCores
			coreFactor_$CurrentType as equipment.$BaseType.$AttrCoreFactor
			$BaseType_t_$CurrentType as math(cpu_$CurrentType*cores_$CurrentType*coreFactor_$CurrentType)  
		}`

	qBaseCalulationCeil = `~equipment.parent @filter(eq(equipment.type,$BaseType)){
			$BaseType_p_$CurrentType as  uid 
			cpu_$CurrentType as equipment.$BaseType.$AttrNumCPU
			cores_$CurrentType as equipment.$BaseType.$AttrNumCores
			coreFactor_$CurrentType as equipment.$BaseType.$AttrCoreFactor
			$BaseType_t_$CurrentType as math(ceil cpu_$CurrentType*cores_$CurrentType*coreFactor_$CurrentType)  
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

	qCalculateLicensesTraverseChildren = `~equipment.parent @filter(eq(equipment.type,$CurrentType)){ 
		$CurrentType_p_$ParentType as uid
		$Query
		$CurrentType_t_$ParentType as sum(val($ChildType_t_$ParentType))
		$RoundOff      
	}`

	qCalculateLicensesTraverseChildrenCeil = `~equipment.parent @filter(eq(equipment.type,$CurrentType)){ 
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
	var(func: uid($BaseTypeIDs))$Filter{
		cpu_$BaseType as equipment.$BaseType.$AttrNumCPU
		cores_$BaseType as equipment.$BaseType.$AttrNumCores
		coreFactor_$BaseType as equipment.$BaseType.$AttrCoreFactor
		$BaseType_t as math(ceil cpu_$BaseType*cores_$BaseType*coreFactor_$BaseType)
	}`
)

type opsMatrix struct {
	EqTypeTree     []*v1.EquipmentType
	BaseType       *v1.EquipmentType
	AggregateLevel *v1.EquipmentType
	CPU            *v1.Attribute
	Cores          *v1.Attribute
	CoreFactor     *v1.Attribute
}

func replacer(q string, params map[string]string) string {
	for key, val := range params {
		q = strings.Replace(q, key, val, -1)
	}
	return q
}

func queryBuilder(ops *v1.MetricOPSComputed, id string) string {
	//q := ""
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
		getToBase(ops.EqTypeTree[:index+1]),
		getToTop(ops.EqTypeTree[index:], index > 0),
		caluclateFromTop(ops.EqTypeTree[index:], ops.CoreFactorAttr, ops.NumCPUAttr, ops.NumCoresAttr, aggregateIndex-index),
		licenses(ops.EqTypeTree[index:], aggregateIndex-index),
	}, "\n\t"), map[string]string{"$id": id}) + "\n}"
}

func getToBase(eqTypes []*v1.EquipmentType) string {
	queries := []string{}
	for i := range eqTypes {
		if i == 0 {
			vars := map[string]string{
				"$CurrentType": eqTypes[i].Type,
			}
			queries = append(queries, replacer(qDirectEquipments, vars))
			continue
		}
		vars := map[string]string{
			"$id":          eqTypes[i-1].Type + "IDs",
			"$CurrentType": eqTypes[i].Type,
		}
		queries = append(queries, replacer(qEquipmentFromChild, vars))
		vars = map[string]string{
			"$CurrentType": eqTypes[i].Type,
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

func caluclateFromTop(eqTypes []*v1.EquipmentType, cf, cpu, cores *v1.Attribute, agIdx int) string {
	queries := []string{}
	filterIDs := []string{}
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
		if i == agIdx+1 {
			cl = qCalculateLicensesCeil
		}

		if i == 0 {
			q := replacer(qJustBaseCalculation, map[string]string{
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
