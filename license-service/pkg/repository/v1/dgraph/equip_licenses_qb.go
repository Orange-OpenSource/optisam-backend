// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"bytes"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"text/template"
)

func queryBuilderEquipOPS(mat *v1.MetricOPSComputed, templ *template.Template, equipID, eqType string) (string, error) {
	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, &EquipProcCal{
		EquipID: equipID,
		EqType:  eqType,
		Met:     mat,
	}); err != nil {
		return "", err
	}

	return formatter(buf.String()), nil
}
