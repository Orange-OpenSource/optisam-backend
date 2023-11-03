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

// CreateMetricUserConcurentStandard handles UCS metric creation
func (l *MetricRepository) CreateMetricUserConcurentStandard(ctx context.Context, met *v1.MetricUCS, scope string) (retmet *v1.MetricUCS, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricUserConcurentStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.user_concurrent.profile",
			ObjectValue: stringObjectValue(met.Profile),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricUCS"),
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
		//	CommitNow: true,
	}
	txn := l.dg.NewTxn()
	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateMetricUserConcurentStandard - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricUserConcurentStandard - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricUserConcurentStandard - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricUserConcurentStandard - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricUserConcurentStandard - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricUserConcurentStandard - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// GetMetricConfigConcurentUser implements Metric GetMetricConfigConcurentUser function
func (l *MetricRepository) GetMetricConfigConcurentUser(ctx context.Context, metName string, scope string) (*v1.MetricUCS, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `))@filter(eq(scopes,` + scope + `)){
			Name: metric.name
			Profile: metric.user_concurrent.profile
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigConcurentUser - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type user_conc")
	}
	type Resp struct {
		Metric []v1.MetricUCS `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigConcurentUser - Unmarshal failed", zap.Error(err), zap.String("query", q))
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

func (l *MetricRepository) UpdateMetricUCS(ctx context.Context, met *v1.MetricUCS, scope string) error {
	q := `query {
		var(func: eq(metric.name,` + met.Name + `))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	set := `
		uid(ID) <metric.user_concurrent.profile> "` + met.Profile + `" .
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
		logger.Log.Error("dgraph/UpdateMetricUCS - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}
