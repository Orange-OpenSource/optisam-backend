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
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

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

// ListMetricNUP implements Licence ListMetricNUP function
func (l *LicenseRepository) ListMetricNUP(ctx context.Context, scopes []string) ([]*v1.MetricNUPOracle, error) {
	respJson, err := l.listMetricWithMetricType(ctx, v1.MetricOracleNUPStandard, scopes)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricNUP - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricOracleNUP
	}
	var data Resp
	if err := json.Unmarshal(respJson, &data); err != nil {
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
	}, nil
}
