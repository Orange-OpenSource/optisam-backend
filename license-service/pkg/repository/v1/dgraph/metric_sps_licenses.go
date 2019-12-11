// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

// MetricSPSComputedLicenses implements Licence MetricSPSComputedLicenses function
func (l *LicenseRepository) MetricSPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricSPSComputed, scopes []string) (uint64, uint64, error) {
	q := queryBuilderSPS(id, mat)
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricSPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricSPSComputedLicenses - query failed")
	}

	type licenses struct {
		Licenses float64
	}

	type totalLicenses struct {
		Licenses        []*licenses
		LicensesNonProd []*licenses
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/MetricSPSComputedLicenses - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/MetricSPSComputedLicenses - Unmarshal failed")
	}

	if len(data.Licenses) == 0 {
		return 0, 0, v1.ErrNoData
	}

	if len(data.LicensesNonProd) == 0 {
		return 0, 0, v1.ErrNoData
	}

	return uint64(data.Licenses[0].Licenses), uint64(data.LicensesNonProd[0].Licenses), nil
}
