package dgraph

import (
	"context"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

// MetricMSSComputedLicenses implements Licence MetricMSSComputedLicenses function
func (l *LicenseRepository) MetricMSSComputedLicenses(ctx context.Context, id []string, mat *v1.MetricMSSComputed, scopes ...string) (uint64, error) {

	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, id...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}

	q := queryBuilderMSS(mat, scopes, id...)
	licensesMSS, err := l.licensesQuery(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricMSSComputedLicenses - query failed",
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"matInfo", mat,
		)
		return 0, errors.New("dgraph/MetricMSSComputedLicenses - query failed")
	}

	return licensesMSS, nil
}

// MetricMSSComputedLicensesAgg implements Licence MetricMSSComputedLicensesAgg function
func (l *LicenseRepository) MetricMSSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricMSSComputed, scopes ...string) (uint64, error) {
	idsMSSAgg, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricMSSComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", err.Error(),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, errors.New("dgraph/MetricMSSComputedLicensesAgg - query failed while getting products of aggregations")
	}
	if len(idsMSSAgg) == 0 {
		logger.Log.Sugar().Errorw("dgraph/MetricMSSComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", errors.New("No data found"),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, nil
	}

	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, idsMSSAgg...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}

	q := queryBuilderMSS(mat, scopes, idsMSSAgg...)
	licensesMSSWSD, err := l.licensesQuery(ctx, q)
	queryFailedMsg := "dgraph/MetricMSSComputedLicensesAgg - query failed"
	if err != nil {
		logger.Log.Sugar().Errorw(queryFailedMsg,
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"prodID", idsMSSAgg,
		)
		return 0, errors.New(queryFailedMsg)
	}

	return licensesMSSWSD, nil
}
