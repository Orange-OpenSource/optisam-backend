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

type metricSPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.sps.base"`
	AttrNumCores   []*id  `json:"metric.sps.attr_num_cores"`
	AtrrCoreFactor []*id  `json:"metric.sps.attr_core_factor"`
}

// ListMetricSPS implements Licence ListMetricSPS function
func (l *LicenseRepository) ListMetricSPS(ctx context.Context, scopes []string) ([]*v1.MetricSPS, error) {
	respJson, err := l.listMetricWithMetricType(ctx, v1.MetricSPSSagProcessorStandard, scopes)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricSPS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricSPS
	}
	var data Resp
	if err := json.Unmarshal(respJson, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricSPS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllSPS(data.Data)
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
