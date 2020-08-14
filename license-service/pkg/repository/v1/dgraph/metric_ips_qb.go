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

func buildQueryIPS(metric *v1.MetricIPSComputed, id ...string) string {
	q := `
{
	var(func:uid($ID)){
		product.equipment @filter(eq(equipment.type,$BaseType)) {
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
	})
}
