package dgraph

import (
	"context"
	"errors"
	"math"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricEquipAttrComputedLicenses implements Licence MetricEquipAttrComputedLicenses function
func (l *LicenseRepository) MetricEquipAttrComputedLicenses(ctx context.Context, id []string, mat *v1.MetricEquipAttrStandComputed, scopes ...string) (uint64, error) {
	q := buildQueryEquipAttr(mat, scopes, id...)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricEquipAttrComputedLicenses - licensesForQueryAll", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricEquipAttrComputedLicenses - query failed")
	}
	return uint64(math.Ceil(float64(sumValue) / mat.Value)), nil
}

// MetricEquipAttrComputedLicensesAgg implements Licence MetricEquipAttrComputedLicensesAgg function
func (l *LicenseRepository) MetricEquipAttrComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricEquipAttrStandComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricEquipAttrComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricEquipAttrComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := buildQueryEquipAttr(mat, scopes, ids...)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricEquipAttrComputedLicensesAgg - licensesForQueryAll", zap.Error(err))
		return 0, errors.New("dgraph/MetricEquipAttrComputedLicensesAgg - query failed")
	}
	return uint64(math.Ceil(float64(sumValue) / mat.Value)), nil
}
