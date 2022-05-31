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
			cpu as equipment.$BaseType.$NumCPU
			comp as  math (ceil (cn*cf*cpu))
		}
		
		var(func:uid(equipIDs_non_prod)){
			cn_non_prod as equipment.$BaseType.$NumCores
			cf_non_prod as equipment.$BaseType.$CoreFactor
			cpu_non_prod as equipment.$BaseType.$NumCPU
			comp_non_prod as  math (ceil (cn_non_prod*cf_non_prod*cpu_non_prod))
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
		"$NumCPU":     metric.NumCPUAttr.Name,
		"$CoreFactor": metric.CoreFactorAttr.Name,
		"$Scopes":     strings.Join(scopes, ","),
	})
}
