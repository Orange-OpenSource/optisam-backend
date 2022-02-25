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

// CreateMetricInstanceNumberStandard handles USS metric creation
func (l *MetricRepository) CreateMetricUSS(ctx context.Context, met *v1.MetricUSS, scope string) (retmet *v1.MetricUSS, retErr error) {
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
		//	CommitNow: true,
	}
	txn := l.dg.NewTxn()
	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateMetricUSS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricUSS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricUSS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricUSS - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricUSS - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricUSS - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// GetMetricConfigUSS implements Metric GetMetricConfigUSS function
func (l *MetricRepository) GetMetricConfigUSS(ctx context.Context, metName string, scope string) (*v1.MetricUSS, error) {
	q := `{
		Data(func: eq(metric.name,"` + metName + `"))@filter(eq(scopes,` + scope + `)){
			Name: metric.name
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigUSS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type uss")
	}
	type Resp struct {
		Metric []v1.MetricUSS `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigUSS - Unmarshal failed", zap.Error(err), zap.String("query", q))
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
