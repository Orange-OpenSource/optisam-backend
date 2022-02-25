package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type id struct {
	ID string `json:"uid"`
}

type metric struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Bottom         []*id  `json:"metric.ops.bottom"`
	Base           []*id  `json:"metric.ops.base"`
	Aggregate      []*id  `json:"metric.ops.aggregate"`
	Top            []*id  `json:"metric.ops.top"`
	AttrNumCores   []*id  `json:"metric.ops.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.ops.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.ops.attr_core_factor"`
}

type EqField struct {
	MetadtaEquipmentType string `json:"metadata.equipment.type"`
}
type AttrField struct {
	AttributeName string `json:"attribute.name"`
}

type metricInfo struct {
	ID                   string
	Name                 string
	BottomEqType         []EqField
	BaseEqType           []EqField
	AggregateLevelEqType []EqField
	TopEqType            []EqField
	NumCoreAttr          []AttrField
	NumCPUAttr           []AttrField
	CoreFactorAttr       []AttrField
	NumOfUsers           uint32
}

// CreateMetricOPS implements Licence CreateMetricOPS function
func (l *MetricRepository) CreateMetricOPS(ctx context.Context, mat *v1.MetricOPS, scope string) (retMat *v1.MetricOPS, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricOPSOracleProcessorStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.bottom",
			ObjectId:  mat.StartEqTypeID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.aggregate",
			ObjectId:  mat.AggerateLevelEqTypeID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.top",
			ObjectId:  mat.EndEqTypeID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
		},
		{
			Subject:   blankID,
			Predicate: "metric.ops.attr_num_cpu",
			ObjectId:  mat.NumCPUAttrID,
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricOracleOPS"),
		},
		{
			Subject:     blankID,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
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
				logger.Log.Error("dgraph/CreateMetricOPS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricOPS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricOPS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricOPS - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricOPS - failed to create matrix", zap.String("reason", err.Error()), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	id, ok := assigned.Uids[mat.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricOPS - failed to create matrix", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	mat.ID = id
	return mat, nil
}

// ListMetricOPS implements Licence ListMetricOPS function
func (l *MetricRepository) ListMetricOPS(ctx context.Context, scope string) ([]*v1.MetricOPS, error) {
	q := `{
		Data(func: eq(metric.type,oracle.processor.standard)) @filter(eq(scopes,` + scope + `)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricOPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type oracle.processor.standard")
	}
	type Resp struct {
		Data []*metric
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricOPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAll(data.Data)
}

// GetMetricConfigOPS implements Metric GetMetricOPS function
func (l *MetricRepository) GetMetricConfigOPS(ctx context.Context, metName string, scope string) (*v1.MetricOPSConfig, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scope + `)){
			Name: metric.name
			BaseEqType: metric.ops.base{
				 metadata.equipment.type
			}
			TopEqType: metric.ops.top{
				metadata.equipment.type
			} 
			BottomEqType: metric.ops.bottom{
				metadata.equipment.type
			} 
			AggregateLevelEqType: metric.ops.aggregate{
				metadata.equipment.type
			}
			CoreFactorAttr: metric.ops.attr_core_factor{
				attribute.name
			}
			NumCoreAttr: metric.ops.attr_num_cores{
				attribute.name
			}
			NumCPUAttr: metric.ops.attr_num_cpu{
				attribute.name
			}
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricOPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type oracle.processor.standard")
	}
	type Resp struct {
		Metric []metricInfo `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricOPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &v1.MetricOPSConfig{
		ID:                  data.Metric[0].ID,
		Name:                data.Metric[0].Name,
		NumCPUAttr:          data.Metric[0].NumCPUAttr[0].AttributeName,
		NumCoreAttr:         data.Metric[0].NumCoreAttr[0].AttributeName,
		CoreFactorAttr:      data.Metric[0].CoreFactorAttr[0].AttributeName,
		StartEqType:         data.Metric[0].BottomEqType[0].MetadtaEquipmentType,
		BaseEqType:          data.Metric[0].BaseEqType[0].MetadtaEquipmentType,
		EndEqType:           data.Metric[0].TopEqType[0].MetadtaEquipmentType,
		AggerateLevelEqType: data.Metric[0].AggregateLevelEqType[0].MetadtaEquipmentType,
	}, nil
}

// GetMetricConfigOPS implements Metric GetMetricOPS function
func (l *MetricRepository) GetMetricConfigOPSID(ctx context.Context, metName string, scope string) (*v1.MetricOPS, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scope + `)){
			uid
			 metric.name
			metric.ops.base{uid}
			metric.ops.top{uid}
			 metric.ops.bottom{uid}
			 metric.ops.aggregate{uid}
			 metric.ops.attr_core_factor{uid}
			 metric.ops.attr_num_cores{uid}
			 metric.ops.attr_num_cpu{uid}
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricOPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type oracle.processor.standard")
	}
	type Resp struct {
		Metric []*metric `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricOPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetric(data.Metric[0])
}

func (l *MetricRepository) UpdateMetricOPS(ctx context.Context, met *v1.MetricOPS, scope string) error {
	q := `query {
		var(func: eq(metric.name,"` + met.Name + `"))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	del := `
	uid(ID) <metric.ops.bottom> * .
	uid(ID) <metric.ops.base> * .
	uid(ID) <metric.ops.aggregate> * .
	uid(ID) <metric.ops.top> * .
	uid(ID) <metric.ops.attr_core_factor> * .
	uid(ID) <metric.ops.attr_num_cores> * .
	uid(ID) <metric.ops.attr_num_cpu> * .	
`
	set := `
	    uid(ID) <metric.ops.bottom> <` + met.StartEqTypeID + `> .
		uid(ID) <metric.ops.base> <` + met.BaseEqTypeID + `> .
		uid(ID) <metric.ops.aggregate> <` + met.AggerateLevelEqTypeID + `> .
		uid(ID) <metric.ops.top> <` + met.EndEqTypeID + `> .
		uid(ID) <metric.ops.attr_core_factor> <` + met.CoreFactorAttrID + `> .
		uid(ID) <metric.ops.attr_num_cores> <` + met.NumCoreAttrID + `> .
	    uid(ID) <metric.ops.attr_num_cpu> <` + met.NumCPUAttrID + `> .	
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
		logger.Log.Error("dgraph/UpdateMetricOPS - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}

func converMetricToModelMetricAll(mts []*metric) ([]*v1.MetricOPS, error) {
	mats := make([]*v1.MetricOPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetric(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetric(m *metric) (*v1.MetricOPS, error) {
	if len(m.Bottom) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - bottom equipment level not found")
	}

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Base equipment level not found")
	}

	if len(m.Aggregate) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Aggregate equipment level not found")
	}

	if len(m.Top) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Top equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCPU) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCPU not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCores not found")
	}

	return &v1.MetricOPS{
		ID:                    m.ID,
		Name:                  m.Name,
		StartEqTypeID:         m.Bottom[0].ID,
		BaseEqTypeID:          m.Base[0].ID,
		AggerateLevelEqTypeID: m.Aggregate[0].ID,
		EndEqTypeID:           m.Top[0].ID,
		CoreFactorAttrID:      m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:         m.AttrNumCores[0].ID,
		NumCPUAttrID:          m.AttrNumCPU[0].ID,
	}, nil
}
