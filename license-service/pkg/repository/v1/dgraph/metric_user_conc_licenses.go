package dgraph

import (
	"context"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricUCSComputedLicenses implements Licence MetricUCSComputedLicenses function
func (l *LicenseRepository) MetricUCSComputedLicenses(ctx context.Context, id []string, mat *v1.MetricUCSComputed, scopes ...string) (uint64, uint64, error) {
	q := buildQueryUCS(mat, scopes, id...)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUCSComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricUCSComputedLicenses - query failed")
	}
	// log.Println("instanceNO : ", instances, " coff", mat.Profile, " lic", licenses)
	return sumValue, sumValue, nil
}

// MetricUCSComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricUCSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricUCSComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getAggregationIDByName(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricUCSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUCSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := buildQueryUCSAgg(mat, scopes, ids)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUCSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUCSComputedLicensesAgg - query failed")
	}
	return sumValue, sumValue, nil
}
