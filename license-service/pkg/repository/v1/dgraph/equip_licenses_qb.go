package dgraph

import (
	"bytes"
	"text/template"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func queryBuilderEquipOPS(mat *v1.MetricOPSComputed, templ *template.Template, equipID, eqType string, scopes string) (string, error) {
	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, &EquipProcCal{
		EquipID: equipID,
		EqType:  eqType,
		Scopes:  scopes,
		Met:     mat,
	}); err != nil {
		return "", err
	}

	return formatter(buf.String()), nil
}
