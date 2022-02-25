package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricOPSComputedLicenses implements Licence MetricOPSComputedLicenses function
func (l *LicenseRepository) MetricOPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricOPSComputed, scopes ...string) (uint64, error) {
	q := queryBuilder(mat, scopes, id)
	// fmt.Println(q)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricOPSComputedLicenses - query failed")
	}

	return licenses, nil
}

// MetricOPSComputedLicensesForAppProduct implements Licence MetricOPSComputedLicensesForAppProduct function
func (l *LicenseRepository) MetricOPSComputedLicensesForAppProduct(ctx context.Context, prodID, appID string, mat *v1.MetricOPSComputed, scopes ...string) (uint64, error) {
	q := queryBuilderForAppProduct(mat, appID, scopes, prodID)
	// fmt.Println(q)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicensesForAppProduct - query failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricOPSComputedLicensesForAppProduct - query failed")
	}

	return licenses, nil
}

// MetricOPSComputedLicensesAgg implements Licence MetricOPSComputedLicensesAgg function
func (l *LicenseRepository) MetricOPSComputedLicensesAgg(ctx context.Context, name, metirc string, mat *v1.MetricOPSComputed, scopes ...string) (uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metirc)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricOPSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}
	q := queryBuilder(mat, scopes, ids...)
	// fmt.Println(q)
	// fmt.Println("we will sleep now")
	// time.Sleep(1 * time.Minute)
	licenses, err := l.licensesForQuery(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicensesAgg - licensesForQuery", zap.Error(err))
		return 0, errors.New("dgraph/MetricOPSComputedLicensesAgg - query failed")
	}
	return licenses, nil
}

type license struct {
	Licenses       float64
	LicensesNoCeil float64
}

// licensesForQueryAll return both licenses both ceiled and normal
func (l *LicenseRepository) licensesForQueryAll(ctx context.Context, q string) (*license, error) {
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return nil, err
	}

	type totalLicenses struct {
		Licenses []*license
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicenses - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("unmarshal failed")
	}

	if len(data.Licenses) == 0 {
		return nil, v1.ErrNoData
	}

	if len(data.Licenses) == 2 {
		data.Licenses[0].LicensesNoCeil = data.Licenses[1].LicensesNoCeil
	}

	return data.Licenses[0], nil
}

// depricated use licensesForQueryAll in future
func (l *LicenseRepository) licensesForQuery(ctx context.Context, q string) (uint64, error) {
	lic, err := l.licensesForQueryAll(ctx, q)
	if err != nil && err == v1.ErrNoData {
		logger.Log.Error("repo-dgraph/licensesForQuery licensesForQueryAll - no licesnse were found for query")
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return uint64(lic.Licenses), nil
}
