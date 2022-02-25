package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

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

func (l *LicenseRepository) CreateMetricAttrSum(ctx context.Context, met *v1.MetricAttrSumStand, attribute *v1.Attribute, scope string) (retmet *v1.MetricAttrSumStand, retErr error) {
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
func (l *LicenseRepository) ListMetricAttrSum(ctx context.Context, scopes ...string) ([]*v1.MetricAttrSumStand, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricAttrSumStandard, scopes...)
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
