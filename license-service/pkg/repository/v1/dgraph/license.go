package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
	"text/template"

	"github.com/dgraph-io/dgo/v2"
	"go.uber.org/zap"
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

// CreateEquipmentType implements Licence CreateEquipmentType function
func (l *LicenseRepository) GetAggregations(ctx context.Context, editor, scope string) ([]*v1.Aggregation, error) {

	q := `{
		aggregations(func:eq(dgraph.type,"Aggregation"))@filter(eq(scopes,"` + scope + `") `

	if editor != "" {
		q += ` and eq(aggregation.editor,"` + editor + `")`
	}
	q += `){
			aggregation.name
		   	aggregation.swidtags
	    }
	}`
	logger.Log.Debug("Query called for GetAggregations", zap.String("query", q))
	respJSON, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetAggregations - query failed", zap.Error(err), zap.String("query", q))
		return nil, fmt.Errorf(" Failed to get  aggregation info %s", err.Error())
	}

	var data *RightsInfo
	if err := json.Unmarshal(respJSON.Json, &data); err != nil {
		logger.Log.Error("dgraph/GetAggregations - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if data == nil {
		return nil, v1.ErrNoData
	}
	return data.Aggregations, nil
}

// CreateEquipmentType implements Licence CreateEquipmentType function
func (l *LicenseRepository) GetAcqRights(ctx context.Context, swidtags []string, editor, scope string) ([]*v1.Acqrights, error) {

	q := `{
		acqrights(func:eq(dgraph.type,"AcquiredRights"))@filter(eq(scopes,"` + scope + `")`

	if editor != "" {
		q += ` and eq(acqRights.editor,"` + editor + `")`
	}
	if len(swidtags) != 0 {
		q += `and not eq (acqRights.swidtag,[` + strings.Join(swidtags, ",") + `])`
	}
	q += `) @groupby(acqRights.swidtag){
	    }
	}`
	logger.Log.Debug("Query called for GetAcqRights", zap.String("query", q))
	respJSON, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetAcqRights - query failed", zap.Error(err), zap.String("query", q))
		return nil, fmt.Errorf(" Failed to get  acqrights info %s", err.Error())
	}

	var data *RightsInfo
	if err := json.Unmarshal(respJSON.Json, &data); err != nil {
		logger.Log.Error("dgraph/GetAcqRights - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if data == nil {
		return nil, v1.ErrNoData
	}
	if len(data.AcquiredRights) == 0 {
		return nil, v1.ErrNoData
	}
	return data.AcquiredRights[0].GroupBySwidtag, nil
}

type RightsInfo struct {
	Aggregations   []*v1.Aggregation          `json:"aggregations"`
	AcquiredRights []*AcqRightsGroupBySwidtag `json:"acqrights"`
}

type AcqRightsGroupBySwidtag struct {
	GroupBySwidtag []*v1.Acqrights `json:"@groupby"`
}
