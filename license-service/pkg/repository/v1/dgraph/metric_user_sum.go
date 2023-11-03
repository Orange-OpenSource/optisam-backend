package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metricUserSum struct {
	ID   string `json:"uid"`
	Name string `json:"metric.name"`
}

func (l *LicenseRepository) CreateMetricUserSum(ctx context.Context, met *v1.MetricUserSumStand, scope string) (retmet *v1.MetricUserSumStand, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricUserSumStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricUSS"),
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
				logger.Log.Error("dgraph/CreateMetricUserSum - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricUserSum - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricUserSum - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricUserSum - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricUserSum - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricUserSum - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// ListMetricUserSum implements Licence ListMetricUserSum function
func (l *LicenseRepository) ListMetricUserSum(ctx context.Context, scopes ...string) ([]*v1.MetricUserSumStand, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricUserSumStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricUserSum - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricUserSum
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricUserSum - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllUserSum(data.Data)
}

func converMetricToModelMetricAllUserSum(mets []*metricUserSum) ([]*v1.MetricUserSumStand, error) {
	modelMets := make([]*v1.MetricUserSumStand, len(mets))
	for i := range mets {
		modelMets[i] = converMetricToModelMetricUserSum(mets[i])
	}
	return modelMets, nil
}

func converMetricToModelMetricUserSum(m *metricUserSum) *v1.MetricUserSumStand {
	return &v1.MetricUserSumStand{
		ID:   m.ID,
		Name: m.Name,
	}
}
