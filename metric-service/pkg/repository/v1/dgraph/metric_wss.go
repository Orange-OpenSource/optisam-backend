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

func (l *MetricRepository) CreateMetricWindowServerStandard(ctx context.Context, met *v1.MetricWSS) (retmet *v1.MetricWSS, retErr error) {
	blankID := blankID(met.MetricName)
	nquads := []*api.NQuad{
		{
			Subject:     blankID,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue("metric"),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue(met.MetricType),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.MetricName),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.wss.reference",
			ObjectValue: stringObjectValue(met.Reference),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.wss.cpu",
			ObjectValue: stringObjectValue(met.CPU),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.wss.core",
			ObjectValue: stringObjectValue(met.Core),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricWSS"),
		},
		{
			Subject:     blankID,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(met.Scope),
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
				logger.Log.Error("dgraph/CreateMetricWindowServerStandard - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricWindowServerStandard - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricWindowServerStandard - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricWindowServerStandard - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricWindowServerStandard - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.MetricName]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricWindowServerStandard - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

func (l *MetricRepository) GetMetricConfigWindowServerStandard(ctx context.Context, metName string, scope string) (*v1.MetricWSS, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			MetricName: metric.name
			MetricType: metric.type
			Reference: metric.wss.reference
			CPU: metric.wss.cpu
			Core: metric.wss.core
			Default: metric.default
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigWindowServerStandard - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type win_server_standard")
	}
	type Resp struct {
		Metric []v1.MetricWSS `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigWindowServerStandard - Unmarshal failed", zap.Error(err), zap.String("query", q))
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
