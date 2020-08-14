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
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

type metricINM struct {
	ID          string  `json:"uid"`
	Name        string  `json:"metric.name"`
	Coefficient float32 `json:"metric.instancenumber.coefficient"`
}

// ListMetricINM implements Licence ListMetricINM function
func (l *LicenseRepository) ListMetricINM(ctx context.Context, scopes []string) ([]*v1.MetricINM, error) {
	respJson, err := l.listMetricWithMetricType(ctx, v1.MetricInstanceNumberStandard, scopes)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricINM - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricINM
	}
	log.Println("INM metrics ", string(respJson))
	var data Resp
	if err := json.Unmarshal(respJson, &data); err != nil {
		//fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricINM - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllINM(data.Data)
}

func converMetricToModelMetricAllINM(mts []*metricINM) ([]*v1.MetricINM, error) {
	mats := make([]*v1.MetricINM, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricINM(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricINM(m *metricINM) (*v1.MetricINM, error) {

	return &v1.MetricINM{
		ID:          m.ID,
		Name:        m.Name,
		Coefficient: m.Coefficient,
	}, nil
}
