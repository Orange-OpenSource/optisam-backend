package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// MetricOPSComputedLicenses implements Licence MetricOPSComputedLicenses function
func (l *LicenseRepository) MetricOPSComputedLicenses(ctx context.Context, id []string, mat *v1.MetricOPSComputed, scopes ...string) (uint64, error) {
	prodAllocatMetricEquipment, err := l.GetProdAllocatedMetric(ctx, id, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicenses - unable to get allocated equipments", zap.Error(err))
		return 0, errors.New("dgraph/MetricOPSComputedLicenses - unable to get allocated equipments")
	}

	opsTransformNUPMetricNamed := ""
	// Get NUP metric if this ops metric exists as transform metric name
	transformNUPMetric, _ := l.GetMetricNUPByTransformMetricName(ctx, mat.Name, scopes[0])
	if transformNUPMetric != nil {
		opsTransformNUPMetricNamed = transformNUPMetric.Name
	}

	equipIDs := filterMetricEquipments(prodAllocatMetricEquipment, mat.Name, opsTransformNUPMetricNamed)
	q := queryBuilder(mat, scopes, equipIDs, id...)
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
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metirc, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, errors.New("dgraph/MetricOPSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, nil
	}

	prodAllocatMetricEquipment, err := l.GetProdAllocatedMetric(ctx, ids, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricOPSComputedLicenses - unable to get allocated equipments", zap.Error(err))
		return 0, errors.New("dgraph/MetricOPSComputedLicenses - unable to get allocated equipments")
	}

	opsTransformNUPMetricNamed := ""
	// Get NUP metric if this ops metric exists as transform metric name
	transformNUPMetric, _ := l.GetMetricNUPByTransformMetricName(ctx, metirc, scopes[0])
	if transformNUPMetric != nil {
		opsTransformNUPMetricNamed = transformNUPMetric.Name
	}

	equipIDs := filterMetricEquipments(prodAllocatMetricEquipment, metirc, opsTransformNUPMetricNamed)
	q := queryBuilder(mat, scopes, equipIDs, ids...)
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
		return &license{
			Licenses:       0,
			LicensesNoCeil: 0,
		}, nil
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
