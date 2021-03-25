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

type metricOracleNUP struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Bottom         []*id  `json:"metric.oracle_nup.bottom"`
	Base           []*id  `json:"metric.oracle_nup.base"`
	Aggregate      []*id  `json:"metric.oracle_nup.aggregate"`
	Top            []*id  `json:"metric.oracle_nup.top"`
	AttrNumCores   []*id  `json:"metric.oracle_nup.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.oracle_nup.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.oracle_nup.attr_core_factor"`
	NumberOfUsers  uint32 `json:"metric.oracle_nup.num_users"`
}

// CreateMetricOracleNUPStandard implements Licence CreateMetricOracleNUPStandard function
func (l *MetricRepository) CreateMetricOracleNUPStandard(ctx context.Context, mat *v1.MetricNUPOracle, scope string) (retMat *v1.MetricNUPOracle, retErr error) {
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
			ObjectValue: stringObjectValue(v1.MetricOracleNUPStandard.String()),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "metric.name",
			ObjectValue: stringObjectValue(mat.Name),
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.bottom",
			ObjectId:  mat.StartEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.base",
			ObjectId:  mat.BaseEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.aggregate",
			ObjectId:  mat.AggerateLevelEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.top",
			ObjectId:  mat.EndEqTypeID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.attr_core_factor",
			ObjectId:  mat.CoreFactorAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.attr_num_cores",
			ObjectId:  mat.NumCoreAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.attr_num_cpu",
			ObjectId:  mat.NumCPUAttrID,
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "metric.oracle_nup.num_users",
			ObjectValue: &api.Value{
				Val: &api.Value_IntVal{
					IntVal: int64(mat.NumberOfUsers),
				},
			},
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetricOracleNUP"),
		},
		&api.NQuad{
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
				logger.Log.Error("dgraph/CreateMetricOracleNUPStandard - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateMetricOracleNUPStandard - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateMetricOracleNUPStandard - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateMetricOracleNUPStandard - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateMetricOracleNUPStandard - failed to create matrix", zap.String("reason", err.Error()), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	id, ok := assigned.Uids[mat.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateMetricOracleNUPStandard - failed to create matrix", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("matrix", mat))
		return nil, errors.New("cannot create matrix")
	}
	mat.ID = id
	return mat, nil
}

// ListMetricNUP implements Licence ListMetricNUP function
func (l *MetricRepository) ListMetricNUP(ctx context.Context, scope string) ([]*v1.MetricNUPOracle, error) {
	q := `{
		Data(func: eq(metric.type,oracle.nup.standard))@filter(eq(scopes,` + scope + `)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricNUP - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type oracle.nup.standard")
	}
	type Resp struct {
		Data []*metricOracleNUP
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricNUP - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricNUPAll(data.Data)
}

// GetMetricConfigNUP implements Metric GetMetricConfigNUP function
func (l *MetricRepository) GetMetricConfigNUP(ctx context.Context, metName string, scope string) (*v1.MetricNUPConfig, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scope + `)){
			Name: metric.name
			BaseEqType: metric.oracle_nup.base{
				 metadata.equipment.type
			}
			TopEqType: metric.oracle_nup.top{
				metadata.equipment.type
			} 
			BottomEqType: metric.oracle_nup.bottom{
				metadata.equipment.type
			} 
			AggregateLevelEqType: metric.oracle_nup.aggregate{
				metadata.equipment.type
			}
			CoreFactorAttr: metric.oracle_nup.attr_core_factor{
				attribute.name
			}
			NumCoreAttr: metric.oracle_nup.attr_num_cores{
				attribute.name
			}
			NumCPUAttr: metric.oracle_nup.attr_num_cpu{
				attribute.name
			}
			NumOfUsers: metric.oracle_nup.num_users
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigNUP - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrices of type nup")
	}
	type Resp struct {
		Metric []metricInfo `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricConfigNUP - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return &v1.MetricNUPConfig{
		ID:                  data.Metric[0].ID,
		Name:                data.Metric[0].Name,
		NumCPUAttr:          data.Metric[0].NumCPUAttr[0].AttributeName,
		NumCoreAttr:         data.Metric[0].NumCoreAttr[0].AttributeName,
		CoreFactorAttr:      data.Metric[0].CoreFactorAttr[0].AttributeName,
		StartEqType:         data.Metric[0].BottomEqType[0].MetadtaEquipmentType,
		BaseEqType:          data.Metric[0].BaseEqType[0].MetadtaEquipmentType,
		EndEqType:           data.Metric[0].TopEqType[0].MetadtaEquipmentType,
		AggerateLevelEqType: data.Metric[0].AggregateLevelEqType[0].MetadtaEquipmentType,
		NumberOfUsers:       data.Metric[0].NumOfUsers,
	}, nil
}

func converMetricToModelMetricNUPAll(mts []*metricOracleNUP) ([]*v1.MetricNUPOracle, error) {
	mats := make([]*v1.MetricNUPOracle, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricNUP(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricNUP(m *metricOracleNUP) (*v1.MetricNUPOracle, error) {
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

	return &v1.MetricNUPOracle{
		ID:                    m.ID,
		Name:                  m.Name,
		StartEqTypeID:         m.Bottom[0].ID,
		BaseEqTypeID:          m.Base[0].ID,
		AggerateLevelEqTypeID: m.Aggregate[0].ID,
		EndEqTypeID:           m.Top[0].ID,
		CoreFactorAttrID:      m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:         m.AttrNumCores[0].ID,
		NumCPUAttrID:          m.AttrNumCPU[0].ID,
		NumberOfUsers:         m.NumberOfUsers,
	}, nil
}
