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

func buildQueryINM(metric *v1.MetricINMComputed, id ...string) string {
	q := `{
		Licenses(func:uid($ID)){
			Licenses: count(product.equipment)
		}
	  }`
	return replacer(q, map[string]string{
		"$ID": strings.Join(id, ","),
	})
}
