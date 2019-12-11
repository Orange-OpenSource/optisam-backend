// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/protos/api"
	"go.uber.org/zap"
)

type metricIPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.ips.base"`
	AttrNumCores   []*id  `json:"metric.ips.attr_num_cores"`
	AtrrCoreFactor []*id  `json:"metric.ips.attr_core_factor"`
}

// CreateMetricIPS implements Licence CreateMetricIPS function
func (l *LicenseRepository) CreateMetricIPS(ctx context.Context, mat *v1.MetricIPS, scopes []string) (retMat *v1.MetricIPS, retErr error) {
	blankID := blankID(mat.Name)
	nquads := []*api.NQuad{
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "type",
			ObjectValue: stringObjectValue("metric"),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue("ibm.pvu.standard"),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ips.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ips.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ips.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
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
				logger.Log.Error("dgraph/CreateMetricSPS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricSPS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricSPS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricSPS - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricSPS - failed to create matrix", zap.String("reason", err.Error()), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	id, ok := assigned.Uids[mat.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricOPS - failed to create matrix", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	mat.ID = id
	return mat, nil
}

// ListMetricIPS implements Licence ListMetricIPS function
func (l *LicenseRepository) ListMetricIPS(ctx context.Context, scopes []string) ([]*v1.MetricIPS, error) {
	q := `{
		Data(func: eq(metric.type,ibm.pvu.standard)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricIPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type ips.processor.standard")
	}
	type Resp struct {
		Data []*metricIPS
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		//fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricIPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllIPS(data.Data)
}

func converMetricToModelMetricAllIPS(mts []*metricIPS) ([]*v1.MetricIPS, error) {
	mats := make([]*v1.MetricIPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricIPS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricIPS(m *metricIPS) (*v1.MetricIPS, error) {

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - Base equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AttrNumCores not found")
	}

	return &v1.MetricIPS{
		ID:               m.ID,
		Name:             m.Name,
		BaseEqTypeID:     m.Base[0].ID,
		CoreFactorAttrID: m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:    m.AttrNumCores[0].ID,
	}, nil
}
