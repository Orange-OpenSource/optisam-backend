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

func buildQueryACS(metric *v1.MetricACSComputed, scopes []string, id ...string) string {
	q := `{
		var(func:uid($ID)){
			  attrCount as product.equipment @filter(eq(equipment.$BaseType.$AttrName,"$Value") AND eq(scopes,[$Scopes]))
		  }
		Licenses(func:uid(attrCount)){
		  Licenses:count(uid)
		}
	  }`
	return replacer(q, map[string]string{
		"$ID":       strings.Join(id, ","),
		"$BaseType": metric.BaseType.Type,
		"$AttrName": metric.Attribute.Name,
		"$Value":    metric.Value,
		"$Scopes":   strings.Join(scopes, ","),
	})
}
