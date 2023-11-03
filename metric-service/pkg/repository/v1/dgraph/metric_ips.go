package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metricIPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.ips.base"`
	AttrNumCores   []*id  `json:"metric.ips.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.ips.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.ips.attr_core_factor"`
}

// CreateMetricIPS implements Licence CreateMetricIPS function
func (l *MetricRepository) CreateMetricIPS(ctx context.Context, mat *v1.MetricIPS, scope string) (retMat *v1.MetricIPS, retErr error) {
	blankID := blankID(mat.Name)
	nquads := []*api.NQuad{
		{
			Subject:     blankID,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue("metric"),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue(v1.MetricIPSIbmPvuStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		{
			Subject:   blankID,
			Predicate: "metric.ips.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ips.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ips.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ips.attr_num_cpu",
			ObjectId:  mat.NumCPUAttrID,
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricIPS"),
		},
		{
			Subject:     blankID,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.default",
			ObjectValue: boolObjectValue(mat.Default),
		},
	}

	mu := &api.Mutation{
		Set: nquads,
		//	CommitNow: true,
	}
	txn := l.dg.NewTxn()

	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateMetricSPS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricSPS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricSPS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricSPS - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricIPS - failed to create metrics", zap.String("reason", err.Error()), zap.Any("metric", mat))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[mat.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricIPS - failed to create metrics", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", mat))
		return nil, errors.New("cannot create metric")
	}
	mat.ID = id
	return mat, nil
}

// ListMetricIPS implements Licence ListMetricIPS function
func (l *MetricRepository) ListMetricIPS(ctx context.Context, scope string) ([]*v1.MetricIPS, error) {
	q := `{
		Data(func: eq(metric.type,ibm.pvu.standard)) @filter(eq(scopes,` + scope + `)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricIPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type ibm.pvu.standard")
	}
	type Resp struct {
		Data []*metricIPS
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricIPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllIPS(data.Data)
}

// GetMetricConfigIPS implements Metric GetMetricConfigIPS function
func (l *MetricRepository) GetMetricConfigIPS(ctx context.Context, metName string, scope string) (*v1.MetricIPSConfig, error) {
	q := `{
		Data(func: eq(metric.name,"` + metName + `"))@filter(eq(scopes,"` + scope + `")){
			Name: metric.name
			BaseEqType: metric.ips.base{
				 metadata.equipment.type
			}
			CoreFactorAttr: metric.ips.attr_core_factor{
				attribute.name
			}
			NumCoreAttr: metric.ips.attr_num_cores{
				attribute.name
			}
			NumCPUAttr: metric.ips.attr_num_cpu{
				attribute.name
			}
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigIPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type ips")
	}
	type Resp struct {
		Metric []metricInfo `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		//	fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigIPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &v1.MetricIPSConfig{
		ID:             data.Metric[0].ID,
		Name:           data.Metric[0].Name,
		NumCoreAttr:    data.Metric[0].NumCoreAttr[0].AttributeName,
		NumCPUAttr:     data.Metric[0].NumCPUAttr[0].AttributeName,
		CoreFactorAttr: data.Metric[0].CoreFactorAttr[0].AttributeName,
		BaseEqType:     data.Metric[0].BaseEqType[0].MetadtaEquipmentType,
	}, nil
}

func (l *MetricRepository) GetMetricConfigIPSID(ctx context.Context, metName string, scope string) (*v1.MetricIPS, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			 uid
			 metric.name
			 metric.ips.base{uid}
			 metric.ips.attr_core_factor{uid}
			 metric.ips.attr_num_cores{uid}
			 metric.ips.attr_num_cpu{uid}
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigIPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type ips")
	}
	type Resp struct {
		Metric []*metricIPS `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigIPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricIPS(data.Metric[0])
}

func (l *MetricRepository) UpdateMetricIPS(ctx context.Context, met *v1.MetricIPS, scope string) error {
	q := `query {
		var(func: eq(metric.name,"` + met.Name + `"))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	del := `
	uid(ID) <metric.ips.base> * .
	uid(ID) <metric.ips.attr_core_factor> * .
	uid(ID) <metric.ips.attr_num_cores> * .
	uid(ID) <metric.ips.attr_num_cpu> * .	
    `
	set := `
	    uid(ID) <metric.ips.base> <` + met.BaseEqTypeID + `> .
	    uid(ID) <metric.ips.attr_core_factor> <` + met.CoreFactorAttrID + `> .
		uid(ID) <metric.ips.attr_num_cores> <` + met.NumCoreAttrID + `> .
		uid(ID) <metric.ips.attr_num_cpu> <` + met.NumCPUAttrID + `> .	
	`
	req := &api.Request{
		Query: q,
		Mutations: []*api.Mutation{
			{
				DelNquads: []byte(del),
			},
			{
				SetNquads: []byte(set),
			},
		},
		CommitNow: true,
	}
	if _, err := l.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("dgraph/UpdateMetricIPS - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}

func converMetricToModelMetricAllIPS(mts []*metricIPS) ([]*v1.MetricIPS, error) {
	mats := make([]*v1.MetricIPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricIPS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricIPS(m *metricIPS) (*v1.MetricIPS, error) {

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - Base equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AttrNumCores not found")
	}

	if len(m.AttrNumCPU) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AttrNumCPU not found")
	}

	return &v1.MetricIPS{
		ID:               m.ID,
		Name:             m.Name,
		BaseEqTypeID:     m.Base[0].ID,
		CoreFactorAttrID: m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:    m.AttrNumCores[0].ID,
		NumCPUAttrID:     m.AttrNumCPU[0].ID,
	}, nil
}

func blankID(id string) string {
	return "_:" + id
}

func stringObjectValue(val string) *api.Value {
	return &api.Value{
		Val: &api.Value_StrVal{
			StrVal: val,
		},
	}
}
