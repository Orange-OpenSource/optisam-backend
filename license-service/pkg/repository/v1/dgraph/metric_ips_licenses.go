// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricIPSComputedLicenses implements Licence MetricIPSComputedLicenses function
func (l *LicenseRepository) MetricIPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricIPSComputed, scopes []string) (uint64, error) {
	q := buildQueryIPS(mat, id)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricIPSComputedLicenses - query failed")
	}

	return licenses, nil
}

// MetricIPSComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricIPSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricIPSComputed, scopes []string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricIPSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := buildQueryIPS(mat, ids...)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricIPSComputedLicensesAgg - query failed")
	}
	return licenses, nil
}
