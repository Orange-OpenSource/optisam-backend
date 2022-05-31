package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricNUPComputedLicenses implements Licence MetricNUPComputedLicenses function
func (l *LicenseRepository) MetricNUPComputedLicenses(ctx context.Context, id string, mat *v1.MetricNUPComputed, scopes ...string) (uint64, uint64, error) {
	// templ, ok := l.templates[nupTemplate]
	// if !ok {
	// 	return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - cannot find template for:  " + string(nupTemplate))
	// }
	opsq := queryBuilderOPSForNUP(mat, scopes, id)
	usersq := buildQueryUsersForNUP(scopes, id)
	opsLicenses, err := l.licensesForQueryAll(ctx, opsq)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicenses - query failed", zap.Error(err), zap.String("query", opsq))
		return 0, 0, errors.New("dgraph/MetricNUPComputedLicenses - query failed")
	}
	userLicenses, err := l.userLicenesForQueryNUP(ctx, usersq)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicenses - query failed", zap.Error(err), zap.String("query", usersq))
		return 0, 0, errors.New("dgraph/MetricNUPComputedLicenses - query failed")
	}
	maxLicenses := math.Max(opsLicenses.Licenses*float64(mat.NumOfUsers), float64(userLicenses))
	return uint64(maxLicenses), userLicenses, nil
}

// MetricNUPComputedLicensesAgg implements Licence MetricNUPComputedLicensesAgg function
func (l *LicenseRepository) MetricNUPComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricNUPComputed, scopes ...string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("getProductUIDsForAggAndMetric query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	// templ, ok := l.templates[nupTemplate]
	// if !ok {
	// 	return 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - cannot find template for:  " + string(nupTemplate))
	// }
	opsq := queryBuilderOPSForNUP(mat, scopes, ids...)
	usersq := buildQueryUsersForNUP(scopes, ids...)
	opsLicenses, err := l.licensesForQueryAll(ctx, opsq)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - query failed", zap.Error(err), zap.String("query", opsq))
		return 0, 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - query failed")
	}
	userLicenses, err := l.userLicenesForQueryNUP(ctx, usersq)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicensesAgg - query failed", zap.Error(err), zap.String("query", usersq))
		return 0, 0, errors.New("dgraph/MetricNUPComputedLicensesAgg - query failed")
	}
	maxLicenses := math.Max(opsLicenses.Licenses*float64(mat.NumOfUsers), float64(userLicenses))
	return uint64(maxLicenses), userLicenses, nil
}

func (l *LicenseRepository) userLicenesForQueryNUP(ctx context.Context, userq string) (uint64, error) {
	usersresp, err := l.dg.NewTxn().Query(ctx, userq)
	if err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicenses - query failed", zap.Error(err), zap.String("users nup query", userq))
		return 0, err
	}
	type users struct {
		TotalUserCount int32
	}
	type totalUsers struct {
		Users []*users
	}
	data := &totalUsers{}
	if err := json.Unmarshal(usersresp.Json, data); err != nil {
		logger.Log.Error("dgraph/MetricNUPComputedLicenses - Unmarshal failed", zap.Error(err), zap.String("users nup query", userq))
		return 0, errors.New("unmarshal failed")
	}
	if len(data.Users) == 0 {
		return 0, nil
	}
	return uint64(data.Users[0].TotalUserCount), nil
}
