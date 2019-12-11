// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
)

// Example Query
/*{
	var(func:uid(0x11190)){
	   ~instance.product  {
		   prodIds as uid
		   instance.id
		}
	}

   var(func:uid(prodIds)) @filter(eq(instance.environment,Production)) {
	  instance.equipment @filter(eq(equipment.type,Server)) {
		equipIDs as uid
	  }
	}

	var(func:uid(prodIds)) @filter( NOT eq(instance.environment,Production)) {
		instance.equipment @filter(eq(equipment.type,Server)) {
		  equipIDs_non_prod as uid
		}
	}

	var(func:uid(equipIDs)){
		cn as equipment.Server.ServerCoresNumber
		cf as equipment.Server.OracleCoreFactor
		comp as  math (cn*cf)
	}

	var(func:uid(equipIDs_non_prod)){
		cn_non_prod as equipment.Server.ServerCoresNumber
		cf_non_prod as equipment.Server.OracleCoreFactor
		comp_non_prod as  math (cn_non_prod*cf_non_prod)
	}

	Licenses(){
	   Licenses: sum(val(comp))
	   LicensesNonProd: sum(val(comp_non_prod))
	}
}*/

func queryBuilderSPS(id string, metric *v1.MetricSPSComputed) string {
	q := `
	{
		var(func:uid($ID)){
		   ~instance.product {
			   prodIds as uid
			   instance.id
			}
		}
	  
	   var(func:uid(prodIds)) @filter(eq(instance.environment,Production)) {
		  instance.equipment @filter(eq(equipment.type,$BaseType)) {
			equipIDs as uid 
		  }
		}

		var(func:uid(prodIds)) @filter( NOT eq(instance.environment,Production)) {
			instance.equipment @filter(eq(equipment.type,$BaseType)) {
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
		"$ID":         id,
		"$BaseType":   metric.BaseType.Type,
		"$NumCores":   metric.NumCoresAttr.Name,
		"$CoreFactor": metric.CoreFactorAttr.Name,
	})
}
