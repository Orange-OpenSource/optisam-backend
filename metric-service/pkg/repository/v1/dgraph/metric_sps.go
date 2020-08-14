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
	v1 "optisam-backend/metric-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metricSPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.sps.base"`
	AttrNumCores   []*id  `json:"metric.sps.attr_num_cores"`
	AtrrCoreFactor []*id  `json:"metric.sps.attr_core_factor"`
}

// CreateMetricSPS implements Licence CreateMetricSPS function
func (l *MetricRepository) CreateMetricSPS(ctx context.Context, mat *v1.MetricSPS, scopes []string) (retMat *v1.MetricSPS, retErr error) {
	blankID := blankID(mat.Name)
	nquads := []*api.NQuad{
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "type_name",
			ObjectValue: stringObjectValue("metric"),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.type",
			ObjectValue: stringObjectValue(v1.MetricSPSSagProcessorStandard.String()),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.sps.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.sps.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.sps.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricSPS"),
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
		logger.Log.Error("dgraph/CreateMetricSPS - failed to create matrix", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	mat.ID = id
	return mat, nil
}

// ListMetricSPS implements Licence ListMetricSPS function
func (l *MetricRepository) ListMetricSPS(ctx context.Context, scopes []string) ([]*v1.MetricSPS, error) {
	q := `{
		Data(func: eq(metric.type,sag.processor.standard)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricSPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type IBM.pvu.standard")
	}
	type Resp struct {
		Data []*metricSPS
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricSPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllSPS(data.Data)
}

// GetMetricConfigSPS implements Metric GetMetricConfigSPS function
func (l *MetricRepository) GetMetricConfigSPS(ctx context.Context, metName string, scopes []string) (*v1.MetricSPSConfig, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)){
			Name: metric.name
			BaseEqType: metric.sps.base{
				 metadata.equipment.type
			}
			CoreFactorAttr: metric.sps.attr_core_factor{
				attribute.name
			}
			NumCoreAttr: metric.sps.attr_num_cores{
				attribute.name
			}
		}
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigSPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type sps")
	}
	type Resp struct {
		Metric []metricInfo `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigSPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &v1.MetricSPSConfig{
		ID:             data.Metric[0].ID,
		Name:           data.Metric[0].Name,
		NumCoreAttr:    data.Metric[0].NumCoreAttr[0].AttributeName,
		CoreFactorAttr: data.Metric[0].CoreFactorAttr[0].AttributeName,
		BaseEqType:     data.Metric[0].BaseEqType[0].MetadtaEquipmentType,
	}, nil
}

func converMetricToModelMetricAllSPS(mts []*metricSPS) ([]*v1.MetricSPS, error) {
	mats := make([]*v1.MetricSPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricSPS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricSPS(m *metricSPS) (*v1.MetricSPS, error) {

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - Base equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - AttrNumCores not found")
	}

	return &v1.MetricSPS{
		ID:               m.ID,
		Name:             m.Name,
		BaseEqTypeID:     m.Base[0].ID,
		CoreFactorAttrID: m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:    m.AttrNumCores[0].ID,
	}, nil
}
