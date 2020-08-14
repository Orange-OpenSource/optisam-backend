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

// MetricNUPComputedLicenses implements Licence MetricNUPComputedLicenses function
func (l *LicenseRepository) MetricNUPComputedLicenses(ctx context.Context, id string, mat *v1.MetricNUPComputed, scopes []string) (uint64, error) {
	templ, ok := l.templates[nupTemplate]
	if !ok {
		return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - cannot find template for:  " + string(nupTemplate))
	}

	q, err := queryBuilderNUP(mat, templ, id)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - queryBuilderNUP", zap.Error(err))
		return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - query cannot be built")
	}
	//fmt.Println(q)
	licenses, err := l.licensesForQueryNUP(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricNUPComputedLicenses - query failed")
	}

	return licenses, nil
}

// MetricNUPComputedLicensesAgg implements Licence MetricNUPComputedLicensesAgg function
func (l *LicenseRepository) MetricNUPComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricNUPComputed, scopes []string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("getProductUIDsForAggAndMetric query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}

	templ, ok := l.templates[nupTemplate]
	if !ok {
		return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - cannot find template for:  " + string(nupTemplate))
	}

	q, err := queryBuilderNUP(mat, templ, ids...)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - queryBuilderNUP", zap.Error(err))
		return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - query cannot be built")
	}

	licenses, err := l.licensesForQueryNUP(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - query failed")
	}
	return licenses, nil
}

func (l *LicenseRepository) licensesForQueryNUP(ctx context.Context, q string) (uint64, error) {
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		//	logger.Log.Error("dgraph/MetricNUPComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, err
	}

	type licenses struct {
		Licenses float64
	}

	type totalLicenses struct {
		Licenses []*licenses
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		//	logger.Log.Error("dgraph/MetricNUPComputedLicenses - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("unmarshal failed")
	}

	if len(data.Licenses) == 0 {
		return 0, nil
	}

	return uint64(data.Licenses[0].Licenses), nil
}
