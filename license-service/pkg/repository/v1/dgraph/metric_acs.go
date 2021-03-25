// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
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

func (l *LicenseRepository) CreateMetricACS(ctx context.Context, met *v1.MetricACS, attribute *v1.Attribute, scopes ...string) (retmet *v1.MetricACS, retErr error) {
	blankID := blankID(met.Name)
	nquads := []*api.NQuad{
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue("metric"),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue(v1.MetricAttrCounterStandard.String()),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(met.Name),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.acs.equipment_type",
			ObjectValue: stringObjectValue(met.EqType),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.acs.attr_name",
			ObjectValue: stringObjectValue(met.AttributeName),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.acs.attr_value",
			ObjectValue: stringObjectValue(met.Value),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricACS"),
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
func (l *LicenseRepository) ListMetricACS(ctx context.Context, scopes ...string) ([]*v1.MetricACS, error) {
	respJson, err := l.listMetricWithMetricType(ctx, v1.MetricAttrCounterStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricACS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricACS
	}
	var data Resp
	if err := json.Unmarshal(respJson, &data); err != nil {
		//fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricACS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllACS(data.Data)
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
