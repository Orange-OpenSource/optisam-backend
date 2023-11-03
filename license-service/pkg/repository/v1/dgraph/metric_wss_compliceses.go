package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

// MetricWSSComputedLicenses implements Licence MetricWSSComputedLicenses function
func (l *LicenseRepository) MetricWSSComputedLicenses(ctx context.Context, id []string, mat *v1.MetricWSSComputed, scopes ...string) (uint64, error) {

	q := queryBuilderWSS(mat, scopes, id...)
	if mat.IsSA {
		q = queryBuilderWithSAWSS(mat, scopes, id...)
	}
	licenses, err := l.licensesForWSS(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricWSSComputedLicenses - query failed",
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"matInfo", mat,
		)
		return 0, errors.New("dgraph/MetricWSSComputedLicenses - query failed")
	}

	return licenses, nil
}

// MetricWSSComputedLicensesAgg implements Licence MetricWSSComputedLicensesAgg function
func (l *LicenseRepository) MetricWSSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricWSSComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/MetricWSSComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", err.Error(),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, errors.New("dgraph/MetricWSSComputedLicensesAgg - query failed while getting products of aggregations")
	}
	if len(ids) == 0 {
		logger.Log.Sugar().Errorw("dgraph/MetricWSSComputedLicensesAgg - getProductUIDsForAggAndMetric",
			"error", errors.New("No data found"),
			"scope", scopes,
			"aggName", name,
			"metric", metric,
		)
		return 0, nil
	}
	q := queryBuilderWSS(mat, scopes, ids...)
	if mat.IsSA {
		q = queryBuilderWithSAWSS(mat, scopes, ids...)
	}
	prod, err := l.licensesForWSS(ctx, q)
	queryFailedMsg := "dgraph/MetricWSSComputedLicensesAgg - query failed"
	if err != nil {
		logger.Log.Sugar().Errorw(queryFailedMsg,
			"error", err.Error(),
			"scope", scopes,
			"query", q,
			"prodID", ids,
		)
		return 0, errors.New(queryFailedMsg)
	}

	return prod, nil
}

func (l *LicenseRepository) licensesForWSS(ctx context.Context, q string) (uint64, error) {
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSS - query failed",
			"error", err.Error(),
			"query", q,
		)
		return 0, fmt.Errorf("query failed, err: %v", err)
	}

	type licenses struct {
		Licenses float64
	}

	type totalLicenses struct {
		Licenses []*licenses
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSS - Unmarshal failed",
			"error", err.Error(),
			"response", resp.Json,
		)
		return 0, fmt.Errorf("unmarshal failed, err: %v", err)
	}

	if len(data.Licenses) == 0 {
		logger.Log.Sugar().Errorw("dgraph/licensesForWSS -"+v1.ErrNoData.Error(),
			"error", v1.ErrNoData.Error(),
			"response", resp.Json,
		)
		return 0, v1.ErrNoData
	}

	return uint64(data.Licenses[0].Licenses), nil
}
