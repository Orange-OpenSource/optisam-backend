// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
)

func queryBuilderSPS(metric *v1.MetricSPSComputed, scopes []string, id ...string) string {
	q := `
	{
		var(func:uid($ID)){
		   ~instance.product {
			   prodIds as uid
			   instance.id
			}
		}
	  
	   var(func:uid(prodIds)) @filter(eq(instance.environment,Production) AND eq(scopes,[$Scopes])) {
		  instance.equipment @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])) {
			equipIDs as uid 
		  }
		}

		var(func:uid(prodIds)) @filter( NOT eq(instance.environment,Production) AND eq(scopes,[$Scopes])) {
			instance.equipment @filter(eq(equipment.type,$BaseType) AND eq(scopes,[$Scopes])) {
			  equipIDs_non_prod as uid 
			}
		}
		
	    var(func:uid(equipIDs)){
			cn as equipment.$BaseType.$NumCores
			cf as equipment.$BaseType.$CoreFactor
			comp as  math (ceil (cn*cf))
		}
		
		var(func:uid(equipIDs_non_prod)){
			cn_non_prod as equipment.$BaseType.$NumCores
			cf_non_prod as equipment.$BaseType.$CoreFactor
			comp_non_prod as  math (ceil (cn_non_prod*cf_non_prod))
	    }
	  
	    Licenses(){
		   Licenses: sum(val(comp))
		}
		LicensesNonProd(){
			Licenses: sum(val(comp_non_prod))
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
