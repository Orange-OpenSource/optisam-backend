package dgraph

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricIPSComputedLicenses implements Licence MetricIPSComputedLicenses function
func (l *LicenseRepository) MetricIPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricIPSComputed, scopes ...string) (uint64, error) {
	q := buildQueryIPS(mat, scopes, id)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricIPSComputedLicenses - query failed")
	}

	return licenses, nil
}

// MetricIPSComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricIPSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricIPSComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricIPSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := buildQueryIPS(mat, scopes, ids...)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricIPSComputedLicensesAgg - query failed")
	}
	return licenses, nil
}
