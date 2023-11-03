package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"
	"go.uber.org/zap"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc/codes"
)

// SQL for metrics

func (l *MetricRepository) CreateMetricSQLStandard(ctx context.Context, met *v1.MetricSQLStand) (retmet *v1.MetricSQLStand, retErr error) {
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
			Predicate:   "metric.sql.standard.reference",
			ObjectValue: stringObjectValue(met.Reference),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.sql.standard.cpu",
			ObjectValue: stringObjectValue(met.CPU),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.sql.standard.core",
			ObjectValue: stringObjectValue(met.Core),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricSQLStandard"),
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
				logger.Log.Sugar().Errorw("service/v1 - CreateMetricSAGProcessorStandard - failed to discard txn",
					"scope", met.Scope,
					"status", codes.Internal,
					"reason", err.Error(),
				)
				retErr = fmt.Errorf("dgraph/CreateMetricAttrSum - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Sugar().Errorw("service/v1 - CreateMetricSAGProcessorStandard - failed to commit txn",
				"scope", met.Scope,
				"status", codes.Internal,
				"reason", err.Error(),
			)
			retErr = fmt.Errorf("dgraph/CreateMetricAttrSum - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - CreateMetricSAGProcessorStandard - failed to create metric",
			"metric", met,
			"scope", met.Scope,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.MetricName]
	if !ok {
		logger.Log.Sugar().Errorw("service/v1 - CreateMetricSAGProcessorStandard - cannot fide Uids in assigned maps",
			"metric", met,
			"scope", met.Scope,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// func boolObjectValue(val bool) *api.Value {
// 	return &api.Value{
// 		Val: &api.Value_BoolVal{
// 			BoolVal: val,
// 		},
// 	}
// }

// GetMetricConfigSQLStandard implements Metric GetMetricConfigSQLStandard function
func (l *MetricRepository) GetMetricConfigSQLStandard(ctx context.Context, metName string, scope string) (*v1.MetricSQLStand, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			MetricName: metric.name
			MetricType: metric.type
			Reference: metric.sql.standard.reference
			CPU: metric.sql.standard.cpu
			Core: metric.sql.standard.core
			Default: metric.default
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigSQLStandard - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type sql_standard")
	}
	type Resp struct {
		Metric []v1.MetricSQLStand `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigSQLStandard - Unmarshal failed", zap.Error(err), zap.String("query", q))
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
