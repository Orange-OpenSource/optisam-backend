package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"strings"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metricACS struct {
	ID       string `json:"uid"`
	Name     string `json:"metric.name"`
	EqType   string `json:"metric.acs.equipment_type"`
	AttrName string `json:"metric.acs.attr_name"`
	Value    string `json:"metric.acs.attr_value"`
}

func (l *MetricRepository) CreateMetricACS(ctx context.Context, met *v1.MetricACS, attribute *v1.Attribute, scope string) (retmet *v1.MetricACS, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricAttrCounterStandard.String()),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.acs.equipment_type",
			ObjectValue: stringObjectValue(met.EqType),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.acs.attr_name",
			ObjectValue: stringObjectValue(met.AttributeName),
		},
		{
			Subject:     blankID,
			Predicate:   "metric.acs.attr_value",
			ObjectValue: stringObjectValue(met.Value),
		},
		{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricACS"),
		},
		{
			Subject:     blankID,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
	}

	if attribute.Type == v1.DataTypeString {
		schemaAttribute := schemaForAttribute(met.EqType, attribute)
		newSchemaAttr := mutateIndexForAttributeSchema(attribute, "equipment."+schemaAttribute)
		if err := l.dg.Alter(context.Background(), &api.Operation{
			Schema: newSchemaAttr,
		}); err != nil {
			logger.Log.Error("dgraph/CreateMetricACS - Alter ")
			return nil, fmt.Errorf("dgraph/CreateMetricACS - cannot mutate index for attribute")
		}
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
				logger.Log.Error("dgraph/CreateMetricACS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricACS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricACS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricACS - cannot commit txn")
		}
	}()
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricACS - failed to create metric", zap.String("reason", err.Error()), zap.Any("metrix", met))
		return nil, errors.New("cannot create metric")
	}
	id, ok := assigned.Uids[met.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricSPS - failed to create metric", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("metric", met))
		return nil, errors.New("cannot create metric")
	}
	met.ID = id
	return met, nil
}

// ListMetricACS implements Licence ListMetricIPS function
func (l *MetricRepository) ListMetricACS(ctx context.Context, scopes string) ([]*v1.MetricACS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricAttrCounterStandard, scopes)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricACS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricACS
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricACS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllACS(data.Data)
}

// GetMetricConfigACS implements Metric GetMetricConfigACS function
func (l *MetricRepository) GetMetricConfigACS(ctx context.Context, metName string, scopes string) (*v1.MetricACS, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scopes + `)){
			Name: metric.name
			EqType: metric.acs.equipment_type
			AttributeName: metric.acs.attr_name
			Value: metric.acs.attr_value
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigIPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type sps")
	}
	type Resp struct {
		Metric []v1.MetricACS `json:"Data"`
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

func (l *MetricRepository) UpdateMetricACS(ctx context.Context, met *v1.MetricACS, scope string) error {
	q := `query {
		var(func: eq(metric.name,` + met.Name + `))@filter(eq(scopes,` + scope + `)){
			ID as uid
		}
	}`
	set := `
	    uid(ID) <metric.acs.equipment_type> "` + met.EqType + `" .
	    uid(ID) <metric.acs.attr_name> "` + met.AttributeName + `" .
		uid(ID) <metric.acs.attr_value> "` + met.Value + `" .	
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
		logger.Log.Error("dgraph/UpdateMetricACS - query failed", zap.Error(err), zap.String("query", req.Query))
		return errors.New("cannot update metric")
	}
	return nil
}

func converMetricToModelMetricAllACS(mets []*metricACS) ([]*v1.MetricACS, error) {
	modelMets := make([]*v1.MetricACS, len(mets))
	for i := range mets {
		m, err := converMetricToModelMetricACS(mets[i])
		if err != nil {
			return nil, err
		}
		modelMets[i] = m
	}

	return modelMets, nil
}

func converMetricToModelMetricACS(m *metricACS) (*v1.MetricACS, error) {

	if len(m.EqType) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricACS - equipment type not found")
	}

	if m.AttrName == "" {
		return nil, errors.New("dgraph converMetricToModelMetricACS - Attribute not found")
	}

	return &v1.MetricACS{
		ID:            m.ID,
		Name:          m.Name,
		EqType:        m.EqType,
		AttributeName: m.AttrName,
		Value:         m.Value,
	}, nil
}

func mutateIndexForAttributeSchema(attr *v1.Attribute, schema string) string {
	if attr.IsSearchable {
		if strings.Contains(schema, "exact") {
			return schema
		}
		idx := strings.Index(schema, "@index(")
		return schema[:idx+7] + "exact," + schema[idx+7:]
	}
	return schema[:len(schema)-1] + "@index(exact) " + schema[len(schema)-1:]
}
