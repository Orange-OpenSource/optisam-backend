package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricUNSComputedLicenses implements Licence MetricUNSComputedLicenses function
func (l *LicenseRepository) MetricUNSComputedLicenses(ctx context.Context, id []string, mat *v1.MetricUNSComputed, scopes ...string) (uint64, uint64, error) {
	q := buildQueryUNS(mat, scopes, id...)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUNSComputedLicenses - licensesForQuery", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricUNSComputedLicenses - query failed")
	}
	// log.Println("instanceNO : ", instances, " coff", mat.Profile, " lic", licenses)
	return sumValue, sumValue, nil
}

// MetricUNSComputedLicensesAgg implements Licence MetricIPSComputedLicensesAgg function
func (l *LicenseRepository) MetricUNSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricUNSComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getAggregationIDByName(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricUNSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUNSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := buildQueryUNSAgg(mat, scopes, ids)
	sumValue, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricUNSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricUNSComputedLicensesAgg - query failed")
	}
	return sumValue, sumValue, nil
}

// nolint: unparam
func (l *LicenseRepository) getAggregationIDByName(ctx context.Context, name, metric string, scopes ...string) (string, error) {
	q := `{
		Agg (func:eq(aggregation.name,"` + name + `"))  ` + agregateFilters(scopeFilters(scopes)) + ` @Normalize @cascade{
			ID:  uid
		}
	  }
	`

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		//	logger.Log.Error("dgraph/MetricUNSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return "", fmt.Errorf("query failed, err: %v", err)
	}

	type id struct {
		ID string
	}

	type data struct {
		Agg []*id
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		return "", fmt.Errorf("unmarshal failed, err: %v", err)
	}
	aggUID := d.Agg[0].ID

	return aggUID, nil
}
