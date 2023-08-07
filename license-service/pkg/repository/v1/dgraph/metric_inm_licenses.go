package dgraph

import (
	"context"
	"errors"
	"math"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricINMComputedLicenses implements Licence MetricINMComputedLicenses function
func (l *LicenseRepository) MetricINMComputedLicenses(ctx context.Context, id []string, mat *v1.MetricINMComputed, scopes ...string) (uint64, uint64, error) {
	q := buildQueryINM(id...)
	instances, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricINMComputedLicenses - query failed")
	}
	licenses := uint64(math.Ceil(float64(instances) / float64(mat.Coefficient)))
	// log.Println("instanceNO : ", instances, " coff", mat.Coefficient, " lic", licenses)
	return licenses, instances, nil
}

// MetricINMComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricINMComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricINMComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricINMComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := buildQueryINM(ids...)
	instances, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricINMComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricINMComputedLicensesAgg - query failed")
	}
	licenses := uint64(math.Ceil(float64(instances) / float64(mat.Coefficient)))
	return licenses, instances, nil
}
