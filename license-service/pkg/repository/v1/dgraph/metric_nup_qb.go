// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"bytes"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
	"text/template"
)

func queryBuilderNUP(mat *v1.MetricNUPComputed, templ *template.Template, id ...string) (string, error) {
	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, mat); err != nil {
		return "", nil
	}

	return formatter(replacer(buf.String(), map[string]string{
		"$ID": strings.Join(id, ","),
	})), nil
}
