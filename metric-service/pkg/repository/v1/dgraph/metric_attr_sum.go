package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metricAttrSum struct {
	ID       string  `json:"uid"`
	Name     string  `json:"metric.name"`
	EqType   string  `json:"metric.attr_sum.equipment_type"`
	AttrName string  `json:"metric.attr_sum.attr_name"`
	RefValue float64 `json:"metric.attr_sum.reference_value"`
}

func (l *MetricRepository) CreateMetricAttrSum(ctx context.Context, met *v1.MetricAttrSumStand, attribute *v1.Attribute, scope string) (retmet *v1.MetricAttrSumStand, retErr error) {
	blankID := blankID(met.Name)
	nquads := []*api.NQuad{
		{
			Subject:     blankID,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue("metric"),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue(v1.MetricAttrSumStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.attr_sum.equipment_type",
			ObjectValue: stringObjectValue(met.EqType),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.attr_sum.attr_name",
			ObjectValue: stringObjectValue(met.AttributeName),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.attr_sum.reference_value",
			ObjectValue: floatObjectValue(met.ReferenceValue),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricAttrSum"),
		},
		{
			Subject:     blankID,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.default",
			ObjectValue: boolObjectValue(met.Default),
		},
	}
	mu := &api.Mutation{
		Set: nquads,
		//	SetNquads: []byte,
		//	CommitNow: true,
	}
	txn := l.dg.NewTxn()
	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateMetricAttrSum - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricAttrSum - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricAttrSum - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricAttrSum - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricAttrSum - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricAttrSum - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// ListMetricAttrSum implements Licence ListMetricAttrSum function
func (l *MetricRepository) ListMetricAttrSum(ctx context.Context, scopes string) ([]*v1.MetricAttrSumStand, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricAttrSumStandard, scopes)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricAttrSum - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricAttrSum
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricAttrSum - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllAttrSum(data.Data)
}

// GetMetricConfigAttrSum implements Metric GetMetricConfigAttrSum function
func (l *MetricRepository) GetMetricConfigAttrSum(ctx context.Context, metName string, scopes string) (*v1.MetricAttrSumStand, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scopes + `)){
			Name: metric.name
			EqType: metric.attr_sum.equipment_type
			AttributeName: metric.attr_sum.attr_name
			ReferenceValue: metric.attr_sum.reference_value
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigAttrSum - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type sps")
	}
	type Resp struct {
		Metric []v1.MetricAttrSumStand `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigAttrSum - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &data.Metric[0], nil
}

func (l *MetricRepository) UpdateMetricAttrSum(ctx context.Context, met *v1.MetricAttrSumStand, scope string) error {
	q := `query {
		var(func: eq(metric.name,` + met.Name + `))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	set := `
	    uid(ID) <metric.attr_sum.equipment_type> "` + met.EqType + `" .
	    uid(ID) <metric.attr_sum.attr_name> "` + met.AttributeName + `" .
		uid(ID) <metric.attr_sum.reference_value> "` + strconv.FormatFloat(met.ReferenceValue, 'E', -1, 64) + `" .	
	`
	req := &api.Request{
		Query: q,
		Mutations: []*api.Mutation{
			{
				SetNquads: []byte(set),
			},
		},
		CommitNow: true,
	}
	if _, err := l.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("dgraph/UpdateMetricAttrSum - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}

func converMetricToModelMetricAllAttrSum(mets []*metricAttrSum) ([]*v1.MetricAttrSumStand, error) {
	modelMets := make([]*v1.MetricAttrSumStand, len(mets))
	for i := range mets {
		m, err := converMetricToModelMetricAttrSum(mets[i])
		if err != nil {
			return nil, err
		}
		modelMets[i] = m
	}
	return modelMets, nil
}

func converMetricToModelMetricAttrSum(m *metricAttrSum) (*v1.MetricAttrSumStand, error) {
	if len(m.EqType) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricAttrSum - equipment type not found")
	}
	if m.AttrName == "" {
		return nil, errors.New("dgraph converMetricToModelMetricAttrSum - Attribute not found")
	}
	return &v1.MetricAttrSumStand{
		ID:             m.ID,
		Name:           m.Name,
		EqType:         m.EqType,
		AttributeName:  m.AttrName,
		ReferenceValue: m.RefValue,
	}, nil
}

func floatObjectValue(val float64) *api.Value {
	return &api.Value{
		Val: &api.Value_DoubleVal{
			DoubleVal: val,
		},
	}
}
