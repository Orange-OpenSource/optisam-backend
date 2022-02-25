package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"strconv"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

// CreateMetricInstanceNumberStandard handles INM metric creation
func (l *MetricRepository) CreateMetricInstanceNumberStandard(ctx context.Context, met *v1.MetricINM, scope string) (retmet *v1.MetricINM, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricInstanceNumberStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:   blankID,
			Predicate: "metric.instancenumber.coefficient",
			ObjectValue: &api.Value{
				Val: &api.Value_IntVal{
					IntVal: int64(met.Coefficient),
				},
			},
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricINM"),
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
				logger.Log.Error("dgraph/CreateMetricInstanceNumberStandard - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricInstanceNumberStandard - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricInstanceNumberStandard - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricInstanceNumberStandard - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricInstanceNumberStandard - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricInstanceNumberStandard - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// GetMetricConfigINM implements Metric GetMetricConfigINM function
func (l *MetricRepository) GetMetricConfigINM(ctx context.Context, metName string, scope string) (*v1.MetricINM, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			Name: metric.name
			Coefficient: metric.instancenumber.coefficient
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigINM - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type sps")
	}
	type Resp struct {
		Metric []v1.MetricINM `json:"Data"`
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
	return &data.Metric[0], nil
}

func (l *MetricRepository) UpdateMetricINM(ctx context.Context, met *v1.MetricINM, scope string) error {
	q := `query {
		var(func: eq(metric.name,` + met.Name + `))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	set := `
		uid(ID) <metric.instancenumber.coefficient> "` + strconv.Itoa(int(met.Coefficient)) + `" .
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
		logger.Log.Error("dgraph/UpdateMetricINM - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}
