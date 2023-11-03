package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricOracleNUP struct {
	ID                  string `json:"uid"`
	Name                string `json:"metric.name"`
	Bottom              []*id  `json:"metric.oracle_nup.bottom"`
	Base                []*id  `json:"metric.oracle_nup.base"`
	Aggregate           []*id  `json:"metric.oracle_nup.aggregate"`
	Top                 []*id  `json:"metric.oracle_nup.top"`
	AttrNumCores        []*id  `json:"metric.oracle_nup.attr_num_cores"`
	AttrNumCPU          []*id  `json:"metric.oracle_nup.attr_num_cpu"`
	AtrrCoreFactor      []*id  `json:"metric.oracle_nup.attr_core_factor"`
	NumberOfUsers       uint32 `json:"metric.oracle_nup.num_users"`
	Transform           bool   `json:"metric.oracle_nup.transform"`
	TransformMetricName string `json:"metric.oracle_nup.transform_metric_name"`
}

// ListMetricNUP implements Licence ListMetricNUP function
func (l *LicenseRepository) ListMetricNUP(ctx context.Context, scopes ...string) ([]*v1.MetricNUPOracle, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricOracleNUPStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricNUP - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricOracleNUP
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricNUP - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricNUPAll(data.Data)
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
		Transform:             m.Transform,
		TransformMetricName:   m.TransformMetricName,
	}, nil
}

func (l *LicenseRepository) GetMetricConfigNUPID(ctx context.Context, metName string, scope string) (*v1.MetricNUPOracle, error) {
	q := `{
		Data(func: eq(metric.name,` + metName + `)) @filter(eq(scopes,` + scope + `)){
			 uid
			 metric.name
			 metric.oracle_nup.base{uid}
			 metric.oracle_nup.top{uid}
			 metric.oracle_nup.bottom{uid}
			 metric.oracle_nup.aggregate{uid}
			 metric.oracle_nup.attr_core_factor{uid}
			 metric.oracle_nup.attr_num_cores{uid}
		     metric.oracle_nup.attr_num_cpu{uid}
			 metric.oracle_nup.num_users
			 metric.oracle_nup.transform
			 metric.oracle_nup.transform_metric_name
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricConfigNUP - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type nup")
	}
	type Resp struct {
		Metric []metricOracleNUP `json:"Data"`
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
	return converMetricToModelMetricNUP(&data.Metric[0])
}

// GetMetricNUPByTransformMetricName implements Metric GetMetricNUPByTransformMetricName function
func (l *LicenseRepository) GetMetricNUPByTransformMetricName(ctx context.Context, transformMetricName string, scope string) (*v1.MetricNUPOracle, error) {
	q := `{
		Data(func: eq(metric.oracle_nup.transform_metric_name,` + transformMetricName + `)) @filter(eq(scopes,` + scope + `)){
			 uid
			 metric.name
			 metric.oracle_nup.base{uid}
			 metric.oracle_nup.top{uid}
			 metric.oracle_nup.bottom{uid}
			 metric.oracle_nup.aggregate{uid}
			 metric.oracle_nup.attr_core_factor{uid}
			 metric.oracle_nup.attr_num_cores{uid}
		     metric.oracle_nup.attr_num_cpu{uid}
			 metric.oracle_nup.num_users
			 metric.oracle_nup.transform
			 metric.oracle_nup.transform_metric_name
		} 
	}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricNUPByTransformMetricName - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot get metrics of type nup")
	}
	type Resp struct {
		Metric []metricOracleNUP `json:"Data"`
	}
	var data Resp
	if err := json.Unmarshal(resp.Json, &data); err != nil {
		fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/GetMetricNUPByTransformMetricName - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("cannot Unmarshal")
	}
	if data.Metric == nil {
		return nil, v1.ErrNoData
	}
	if len(data.Metric) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricNUP(&data.Metric[0])
}
