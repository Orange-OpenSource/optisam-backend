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
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

type predMetric string

// String implements string.Stringer
func (p predMetric) String() string {
	return string(p)
}

const (
	predMetricName predMetric = "metric.name"
)

// ListMetrices implements Licence ListMetrices function
func (l *LicenseRepository) ListMetrices(ctx context.Context, scopes []string) ([]*v1.Metric, error) {

	q := `   {
             Metrics(func:eq(type_name,"metric")){
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

func (l *LicenseRepository) listMetricWithMetricType(ctx context.Context, metType v1.MetricType, scopes []string) (json.RawMessage, error) {
	q := `{
		Data(func: eq(metric.type,` + metType.String() + `)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/listMetricWithMetricType - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New(fmt.Sprintf("cannot get metrices of %s", metType.String()))
	}
	return resp.Json, nil
}
