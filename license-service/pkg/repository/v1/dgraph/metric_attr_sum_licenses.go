package dgraph

import (
	"context"
	"errors"
	"math"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricAttrSumComputedLicenses implements Licence MetricAttrSumComputedLicenses function
func (l *LicenseRepository) MetricAttrSumComputedLicenses(ctx context.Context, id string, mat *v1.MetricAttrSumStandComputed, scopes ...string) (uint64, uint64, error) {
	q := buildQueryAttrSum(mat, scopes, id)
	sumValue, err := l.licensesForQueryAll(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricAttrSumComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricAttrSumComputedLicenses - query failed")
	}
	return uint64(math.Ceil(sumValue.LicensesNoCeil / mat.ReferenceValue)), uint64(sumValue.LicensesNoCeil), nil
}

// MetricAttrSumComputedLicensesAgg implements Licence MetricAttrSumComputedLicensesAgg function
func (l *LicenseRepository) MetricAttrSumComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricAttrSumStandComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
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
		logger.Log.Error("dgraph/MetricAttrSumComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricAttrSumComputedLicensesAgg - query failed")
	}
	return uint64(math.Ceil(sumValue.LicensesNoCeil / mat.ReferenceValue)), uint64(sumValue.LicensesNoCeil), nil
}
