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

// MetricIPSComputedLicenses implements Licence MetricIPSComputedLicenses function
func (l *LicenseRepository) MetricIPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricIPSComputed, scopes []string) (uint64, error) {
	q := buildQueryIPS(id, mat)
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricIPSComputedLicenses - query failed")
	}

	type licenses struct {
		Licenses float64
	}

	type totalLicenses struct {
		Licenses []*licenses
	}

	data := &totalLicenses{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/MetricIPSComputedLicenses - Unmarshal failed", zap.Error(err), zap.String("query", q))
		return 0, errors.New("dgraph/MetricIPSComputedLicenses - Unmarshal failed")
	}

	if len(data.Licenses) == 0 {
		return 0, v1.ErrNoData
	}

	return uint64(data.Licenses[0].Licenses), nil
}
