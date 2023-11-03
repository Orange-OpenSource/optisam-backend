package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

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
		q += `and not eq (acqRights.swidtag,["` + strings.Join(swidtags, ",") + `"])`
	}
	q += `) @groupby(acqRights.swidtag,acqRights.productName){
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

// licensesQuery will execute dGraph statement
func (l *LicenseRepository) licensesQuery(ctx context.Context, q string) (uint64, error) {
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSD - query failed",
			"error", err.Error(),
			"query", q,
		)
		return 0, fmt.Errorf("query failed, err: %v", err)
	}

	type licenses struct {
		Licenses float64
	}

	type totalLicenses struct {
		Licenses []*licenses
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSD - Unmarshal failed",
			"error", err.Error(),
			"response", resp.Json,
		)
		return 0, fmt.Errorf("unmarshal failed, err: %v", err)
	}

	if len(data.Licenses) == 0 {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSD -"+v1.ErrNoData.Error(),
			"error", v1.ErrNoData.Error(),
			"response", resp.Json,
		)
		return 0, v1.ErrNoData
	}

	return uint64(data.Licenses[0].Licenses), nil
}

type RightsInfo struct {
	Aggregations   []*v1.Aggregation          `json:"aggregations"`
	AcquiredRights []*AcqRightsGroupBySwidtag `json:"acqrights"`
}

type AcqRightsGroupBySwidtag struct {
	GroupBySwidtag []*v1.Acqrights `json:"@groupby"`
}
