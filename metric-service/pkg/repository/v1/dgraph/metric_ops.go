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

type id struct {
	ID string `json:"uid"`
}

type metric struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Bottom         []*id  `json:"metric.ops.bottom"`
	Base           []*id  `json:"metric.ops.base"`
	Aggregate      []*id  `json:"metric.ops.aggregate"`
	Top            []*id  `json:"metric.ops.top"`
	AttrNumCores   []*id  `json:"metric.ops.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.ops.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.ops.attr_core_factor"`
}

type EqField struct {
	MetadtaEquipmentType string `json:"metadata.equipment.type"`
}
type AttrField struct {
	AttributeName string `json:"attribute.name"`
}

type metricInfo struct {
	ID                   string
	Name                 string
	BottomEqType         []EqField
	BaseEqType           []EqField
	AggregateLevelEqType []EqField
	TopEqType            []EqField
	NumCoreAttr          []AttrField
	NumCPUAttr           []AttrField
	CoreFactorAttr       []AttrField
	NumOfUsers           uint32
}

// CreateMetricOPS implements Licence CreateMetricOPS function
func (l *MetricRepository) CreateMetricOPS(ctx context.Context, mat *v1.MetricOPS, scopes []string) (retMat *v1.MetricOPS, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricOPSOracleProcessorStandard.String()),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.bottom",
			ObjectId:  mat.StartEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.aggregate",
			ObjectId:  mat.AggerateLevelEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.top",
			ObjectId:  mat.EndEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.ops.attr_num_cpu",
			ObjectId:  mat.NumCPUAttrID,
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricOracleOPS"),
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
				logger.Log.Error("dgraph/CreateMetricOPS - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricOPS - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricOPS - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricOPS - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricOPS - failed to create matrix", zap.String("reason", err.Error()), zap.Any("matrix", mat))
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

// ListMetricOPS implements Licence ListMetricOPS function
func (l *MetricRepository) ListMetricOPS(ctx context.Context, scopes []string) ([]*v1.MetricOPS, error) {
	q := `{
		Data(func: eq(metric.type,oracle.processor.standard)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricOPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type oracle.processor.standard")
	}
	type Resp struct {
		Data []*metric
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricOPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAll(data.Data)
}

// GetMetricConfigOPS implements Metric GetMetricOPS function
func (l *MetricRepository) GetMetricConfigOPS(ctx context.Context, metName string, scopes []string) (*v1.MetricOPSConfig, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)){
			Name: metric.name
			BaseEqType: metric.ops.base{
				 metadata.equipment.type
			}
			TopEqType: metric.ops.top{
				metadata.equipment.type
			} 
			BottomEqType: metric.ops.bottom{
				metadata.equipment.type
			} 
			AggregateLevelEqType: metric.ops.aggregate{
				metadata.equipment.type
			}
			CoreFactorAttr: metric.ops.attr_core_factor{
				attribute.name
			}
			NumCoreAttr: metric.ops.attr_num_cores{
				attribute.name
			}
			NumCPUAttr: metric.ops.attr_num_cpu{
				attribute.name
			}
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricOPS - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type oracle.processor.standard")
	}
	type Resp struct {
		Metric []metricInfo `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricOPS - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &v1.MetricOPSConfig{
		ID:                  data.Metric[0].ID,
		Name:                data.Metric[0].Name,
		NumCPUAttr:          data.Metric[0].NumCPUAttr[0].AttributeName,
		NumCoreAttr:         data.Metric[0].NumCoreAttr[0].AttributeName,
		CoreFactorAttr:      data.Metric[0].CoreFactorAttr[0].AttributeName,
		StartEqType:         data.Metric[0].BottomEqType[0].MetadtaEquipmentType,
		BaseEqType:          data.Metric[0].BaseEqType[0].MetadtaEquipmentType,
		EndEqType:           data.Metric[0].TopEqType[0].MetadtaEquipmentType,
		AggerateLevelEqType: data.Metric[0].AggregateLevelEqType[0].MetadtaEquipmentType,
	}, nil
}

func converMetricToModelMetricAll(mts []*metric) ([]*v1.MetricOPS, error) {
	mats := make([]*v1.MetricOPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetric(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetric(m *metric) (*v1.MetricOPS, error) {
	if len(m.Bottom) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - bottom equipment level not found")
	}

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Base equipment level not found")
	}

	if len(m.Aggregate) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Aggregate equipment level not found")
	}

	if len(m.Top) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - Top equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCPU) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCPU not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCores not found")
	}

	return &v1.MetricOPS{
		ID:                    m.ID,
		Name:                  m.Name,
		StartEqTypeID:         m.Bottom[0].ID,
		BaseEqTypeID:          m.Base[0].ID,
		AggerateLevelEqTypeID: m.Aggregate[0].ID,
		EndEqTypeID:           m.Top[0].ID,
		CoreFactorAttrID:      m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:         m.AttrNumCores[0].ID,
		NumCPUAttrID:          m.AttrNumCPU[0].ID,
	}, nil
}
