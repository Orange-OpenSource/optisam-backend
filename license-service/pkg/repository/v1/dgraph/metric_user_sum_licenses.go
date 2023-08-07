package dgraph

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricUserSumComputedLicenses implements Licence MetricUserSumComputedLicenses function
func (l *LicenseRepository) MetricUserSumComputedLicenses(ctx context.Context, id []string, scopes ...string) (uint64, uint64, error) {
	q := buildQueryUsersForNUP(scopes, "", id...)
	sumValue, err := l.userLicenesForQueryNUP(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUserSumComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricUserSumComputedLicenses - query failed")
	}
	return sumValue, sumValue, nil
}

// MetricUserSumComputedLicensesAgg implements Licence MetricUserSumComputedLicensesAgg function
func (l *LicenseRepository) MetricUserSumComputedLicensesAgg(ctx context.Context, name, metric string, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricUserSumComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUserSumComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := buildQueryUsersForNUP(scopes, "", ids...)
	sumValue, err := l.userLicenesForQueryNUP(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUserSumComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUserSumComputedLicensesAgg - query failed")
	}
	return sumValue, sumValue, nil
}
