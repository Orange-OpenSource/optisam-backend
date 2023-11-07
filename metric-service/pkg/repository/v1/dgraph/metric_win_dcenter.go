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

// Metadata for injectors
type metricWSD struct {
	ID         string `json:"uid"`
	MetricType string `json:"metric.type"`
	MetricName string `json:"metric.name"`
	Reference  string `json:"metric.wsd.reference"`
	Core       string `json:"metric.wsd.core"`
	CPU        string `json:"metric.wsd.cpu"`
}

func (l *MetricRepository) CreateMetricDataCenterForScope(ctx context.Context, met *v1.ScopeMetric) (retmet *v1.ScopeMetric, retErr error) {
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
			Predicate:   "metric.wsd.reference",
			ObjectValue: stringObjectValue(met.Reference),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.wsd.cpu",
			ObjectValue: stringObjectValue(met.CPU),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.wsd.core",
			ObjectValue: stringObjectValue(met.Core),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricWSD"),
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
				logger.Log.Error("dgraph/CreateMetricWinDcenter - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricWinDcenter - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricWinDcenter - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricWinDcenter - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricWinDcenter - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.MetricName]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricWinDcenter - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

func (l *MetricRepository) GetMetricConfigDataCenterForScope(ctx context.Context, metName string, scope string) (*v1.ScopeMetric, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			MetricName: metric.name
			MetricType: metric.type
			Reference: metric.wsd.reference
			CPU: metric.wsd.cpu
			Core: metric.wsd.core
			Default: metric.default
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigDataCenterForScope - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type win_dcenter")
	}
	type Resp struct {
		Metric []v1.ScopeMetric `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigDataCenterForScope - Unmarshal failed", zap.Error(err), zap.String("query", q))
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