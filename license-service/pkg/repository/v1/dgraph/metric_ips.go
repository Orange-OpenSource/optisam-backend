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

type metricIPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.ips.base"`
	AttrNumCores   []*id  `json:"metric.ips.attr_num_cores"`
	AtrrCoreFactor []*id  `json:"metric.ips.attr_core_factor"`
}

// ListMetricIPS implements Licence ListMetricIPS function
func (l *LicenseRepository) ListMetricIPS(ctx context.Context, scopes ...string) ([]*v1.MetricIPS, error) {
	respJson, err := l.listMetricWithMetricType(ctx, v1.MetricIPSIbmPvuStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricIPS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricIPS
	}
	var data Resp
	if err := json.Unmarshal(respJson, &data); err != nil {
		//fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricIPS - Unmarshal failed", zap.Error(err))
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
