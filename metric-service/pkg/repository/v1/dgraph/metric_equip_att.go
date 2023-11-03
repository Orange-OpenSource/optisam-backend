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

// type metricEquipAttr struct {
// 	ID          string `json:"uid"`
// 	Name        string `json:"metric.name"`
// 	EqType      string `json:"metric.equip_attr.equipment_type"`
// 	AttrName    string `json:"metric.equip_attr.attr_name"`
// 	Environment string `json:"metric.equip_attr.environment"`
// 	Value       int32  `json:"metric.equip_attr.value"`
// }

func (l *MetricRepository) CreateMetricEquipAttrStandard(ctx context.Context, met *v1.MetricEquipAttrStand, attribute *v1.Attribute, scope string) (retmet *v1.MetricEquipAttrStand, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricEquipAttrStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.equip_attr.equipment_type",
			ObjectValue: stringObjectValue(met.EqType),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.equip_attr.attr_name",
			ObjectValue: stringObjectValue(met.AttributeName),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.equip_attr.environment",
			ObjectValue: stringObjectValue(met.Environment),
		},
		{
			Subject:   blankID,
			Predicate: "metric.equip_attr.value",
			ObjectValue: &api.Value{
				Val: &api.Value_IntVal{
					IntVal: int64(met.Value),
				},
			},
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricEquipAttr"),
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
				logger.Log.Error("dgraph/CreateMetricEquipAttrStandard - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricEquipAttrStandard - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricEquipAttrStandard - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricEquipAttrStandard - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricEquipAttrStandard - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricEquipAttrStandard - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// GetMetricConfigAttrSum implements Metric GetMetricConfigAttrSum function
func (l *MetricRepository) GetMetricConfigEquipAttr(ctx context.Context, metName string, scopes string) (*v1.MetricEquipAttrStand, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scopes + `)){
			Name: metric.name
			EqType: metric.equip_attr.equipment_type
			AttributeName: metric.equip_attr.attr_name
			Environment: metric.equip_attr.environment
			Value: metric.equip_attr.value
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigEquipAttr - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type sps")
	}
	type Resp struct {
		Metric []v1.MetricEquipAttrStand `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigEquipAttr - Unmarshal failed", zap.Error(err), zap.String("query", q))
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

func (l *MetricRepository) UpdateMetricEquipAttr(ctx context.Context, met *v1.MetricEquipAttrStand, scope string) error {
	q := `query {
		var(func: eq(metric.name,` + met.Name + `))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	set := `
	    uid(ID) <metric.equip_attr.equipment_type> "` + met.EqType + `" .
	    uid(ID) <metric.equip_attr.attr_name> "` + met.AttributeName + `" .
		uid(ID) <metric.equip_attr.environment> "` + met.Environment + `" .
		uid(ID) <metric.equip_attr.value> "` + strconv.Itoa(int(met.Value)) + `" .	
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
		logger.Log.Error("dgraph/UpdateMetricEquipAttr - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}
