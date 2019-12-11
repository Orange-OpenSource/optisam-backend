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
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// ListMetricTypeInfo implements Licence ListMetricTypeInfo function
func (l *LicenseRepository) ListMetricTypeInfo(ctx context.Context, scopes []string) ([]*v1.MetricTypeInfo, error) {
	return v1.MetricTypes, nil
}

// ListMetrices implements Licence ListMetrices function
func (l *LicenseRepository) ListMetrices(ctx context.Context, scopes []string) ([]*v1.Metric, error) {

	q := `   {
             Metrics(func:eq(type,"metric")){
			   ID  : uid
			   Name: metric.name
			   Type: metric.type
		   }
		}

		  `

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListMetrices - cannot complete query transaction")
	}

	type Data struct {
		Metrics []*v1.Metric
	}
	var metricList Data
	if err := json.Unmarshal(resp.GetJson(), &metricList); err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListMetrices - cannot unmarshal Json object")
	}

	return metricList.Metrics, nil
}
