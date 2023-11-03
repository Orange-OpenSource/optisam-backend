package dgraph

import (
	"context"
	"errors"
	"math"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricAttrSumComputedLicenses implements Licence MetricAttrSumComputedLicenses function
func (l *LicenseRepository) MetricAttrSumComputedLicenses(ctx context.Context, id []string, mat *v1.MetricAttrSumStandComputed, scopes ...string) (uint64, uint64, error) {
	q := buildQueryAttrSum(mat, scopes, id...)
	sumValue, err := l.licensesForQueryAll(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricAttrSumComputedLicenses - licensesForQueryAll", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricAttrSumComputedLicenses - query failed")
	}
	return uint64(math.Ceil(sumValue.LicensesNoCeil / mat.ReferenceValue)), uint64(sumValue.LicensesNoCeil), nil
}

// MetricAttrSumComputedLicensesAgg implements Licence MetricAttrSumComputedLicensesAgg function
func (l *LicenseRepository) MetricAttrSumComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricAttrSumStandComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricAttrSumComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricAttrSumComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := buildQueryAttrSum(mat, scopes, ids...)
	sumValue, err := l.licensesForQueryAll(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricAttrSumComputedLicensesAgg - licensesForQueryAll", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricAttrSumComputedLicensesAgg - query failed")
	}
	return uint64(math.Ceil(sumValue.LicensesNoCeil / mat.ReferenceValue)), uint64(sumValue.LicensesNoCeil), nil
}
