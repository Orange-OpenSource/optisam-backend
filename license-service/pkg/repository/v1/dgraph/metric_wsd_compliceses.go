package dgraph

import (
	"context"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

// MetricWSDComputedLicenses implements Licence MetricWSDComputedLicenses function
func (l *LicenseRepository) MetricWSDComputedLicenses(ctx context.Context, id []string, mat *v1.MetricWSDComputed, scopes ...string) (uint64, error) {

	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, id...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}

	q := queryBuilderWSD(mat, scopes, id...)
	if mat.IsSA {
		q = queryBuilderWithSAWSD(mat, scopes, id...)
	}
	licensesIndWSD, err := l.licensesQuery(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricWSDComputedLicenses - query failed",
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"matInfo", mat,
		)
		return 0, errors.New("dgraph/MetricWSDComputedLicenses - query failed")
	}

	return licensesIndWSD, nil
}

// MetricWSDComputedLicensesAgg implements Licence MetricWSDComputedLicensesAgg function
func (l *LicenseRepository) MetricWSDComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricWSDComputed, scopes ...string) (uint64, error) {
	idsWSDAgg, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricWSDComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", err.Error(),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, errors.New("dgraph/MetricWSDComputedLicensesAgg - query failed while getting products of aggregations")
	}
	if len(idsWSDAgg) == 0 {
		logger.Log.Sugar().Errorw("dgraph/MetricWSDComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", errors.New("No data found"),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, nil
	}
	totalVMCount, totalServerCount := l.productEquipmentCount(ctx, idsWSDAgg...)
	if totalVMCount == 0 && totalServerCount == 0 {
		return 0, nil
	}

	q := queryBuilderWSD(mat, scopes, idsWSDAgg...)
	if mat.IsSA {
		q = queryBuilderWithSAWSD(mat, scopes, idsWSDAgg...)
	}
	licensesAggWSD, err := l.licensesQuery(ctx, q)
	queryFailedMsg := "dgraph/MetricWSDComputedLicensesAgg - query failed"
	if err != nil {
		logger.Log.Sugar().Errorw(queryFailedMsg,
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"prodID", idsWSDAgg,
		)
		return 0, errors.New(queryFailedMsg)
	}

	return licensesAggWSD, nil
}
