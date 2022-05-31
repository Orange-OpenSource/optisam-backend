package dgraph

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricACSComputedLicenses implements Licence MetricACSComputedLicenses function
func (l *LicenseRepository) MetricACSComputedLicenses(ctx context.Context, id string, mat *v1.MetricACSComputed, scopes ...string) (uint64, error) {
	q := buildQueryACS(mat, scopes, id)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricACSComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricACSComputedLicenses - query failed")
	}
	return licenses, nil
}

// MetricACSComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricACSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricACSComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricACSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricACSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := buildQueryACS(mat, scopes, ids...)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricACSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricACSComputedLicensesAgg - query failed")
	}
	return licenses, nil
}
