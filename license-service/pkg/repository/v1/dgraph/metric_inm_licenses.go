// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	"log"
	"math"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricINMComputedLicenses implements Licence MetricINMComputedLicenses function
func (l *LicenseRepository) MetricINMComputedLicenses(ctx context.Context, id string, mat *v1.MetricINMComputed, scopes ...string) (uint64, error) {
	q := buildQueryINM(mat, id)
	instances, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricINMComputedLicenses - query failed")
	}
	licenses := uint64(math.Ceil(float64(instances) * float64(mat.Coefficient)))
	log.Println("instanceNO : ", instances, " coff", mat.Coefficient, " lic", licenses)
	return licenses, nil
}

// MetricINMComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricINMComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricINMComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricINMComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := buildQueryINM(mat, ids...)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricINMComputedLicensesAgg - query failed")
	}
	return licenses, nil
}
