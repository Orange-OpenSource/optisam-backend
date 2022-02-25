package dgraph

import (
	"text/template"

	"github.com/dgraph-io/dgo/v2"
)

type templateType string

// LicenseRepository for Dgraph
type LicenseRepository struct {
	dg        *dgo.Dgraph
	templates map[templateType]*template.Template
}

// NewLicenseRepository creates new Repository
func NewLicenseRepository(dg *dgo.Dgraph) *LicenseRepository {
	return &LicenseRepository{
		dg: dg,
	}
}

// NewLicenseRepositoryWithTemplates creates new Repository with templates
func NewLicenseRepositoryWithTemplates(dg *dgo.Dgraph) (*LicenseRepository, error) {
	nupTempl, err := templateNup()
	if err != nil {
		return nil, err
	}
	opsEquipTmpl, err := templEquipOPS()
	if err != nil {
		return nil, err
	}
	return &LicenseRepository{
		dg: dg,
		templates: map[templateType]*template.Template{
			nupTemplate:      nupTempl,
			opsEquipTemplate: opsEquipTmpl,
		},
	}, nil
}
