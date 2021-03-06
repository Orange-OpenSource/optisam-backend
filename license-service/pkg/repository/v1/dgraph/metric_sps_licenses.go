// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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

// MetricSPSComputedLicenses implements Licence MetricSPSComputedLicenses function
func (l *LicenseRepository) MetricSPSComputedLicenses(ctx context.Context, id string, mat *v1.MetricSPSComputed, scopes []string) (uint64, uint64, error) {
	q := queryBuilderSPS(mat, id)
	prod, nonProd, err := l.licensesForSPS(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricSPSComputedLicenses - licensesForSPS", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricSPSComputedLicenses - query failed")
	}

	return prod, nonProd, nil
}

// MetricSPSComputedLicensesAgg implements Licence MetricSPSComputedLicensesAgg function
func (l *LicenseRepository) MetricSPSComputedLicensesAgg(ctx context.Context, name, metric string, mat *v1.MetricSPSComputed, scopes []string) (uint64, uint64, error) {
	ids, err := l.getProductUIDsForAggAndMetric(ctx, name, metric)
	if err != nil {
		logger.Log.Error("dgraph/MetricSPSComputedLicensesAgg - getProductUIDsForAggAndMetric", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricSPSComputedLicensesAgg - query failed")
	}
	if len(ids) == 0 {
		return 0, 0, nil
	}
	q := queryBuilderSPS(mat, ids...)
	prod, nonProd, err := l.licensesForSPS(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/MetricSPSComputedLicensesAgg - licensesForSPS", zap.Error(err))
		return 0, 0, errors.New("dgraph/MetricSPSComputedLicensesAgg - query failed")
	}

	return prod, nonProd, nil
}

func (l *LicenseRepository) getProductUIDsForAggAndMetric(ctx context.Context, name, metric string) ([]string, error) {
	q := `
	 {
		Products (func:eq(product_aggregation.name,"` + name + `"))@Normalize@cascade{
			product_aggregation.products{
			   ID: uid
			   product.acqRights@filter(eq(acqRights.metric,"` + metric + `"))
			}
		}
  
	  }
	 `
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		//	logger.Log.Error("dgraph/MetricSPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return nil, fmt.Errorf("query failed, err: %v", err)
	}

	type id struct {
		ID string
	}

	type dataTemp struct {
		Products []*id
	}

	data := &dataTemp{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		return nil, fmt.Errorf("unmarshal failed, err: %v", err)
	}

	productUIDs := make([]string, len(data.Products))
	for i := range data.Products {
		productUIDs[i] = data.Products[i].ID
	}

	return productUIDs, nil
}

func (l *LicenseRepository) licensesForSPS(ctx context.Context, q string) (uint64, uint64, error) {
	fmt.Println(q)
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		//	logger.Log.Error("dgraph/MetricSPSComputedLicenses - query failed", zap.Error(err), zap.String("query", q))
		return 0, 0, fmt.Errorf("query failed, err: %v", err)
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
		return 0, 0, fmt.Errorf("unmarshal failed, err: %v", err)
	}

	if len(data.Licenses) == 0 {
		return 0, 0, v1.ErrNoData
	}

	if len(data.LicensesNonProd) == 0 {
		return 0, 0, v1.ErrNoData
	}

	return uint64(data.Licenses[0].Licenses), uint64(data.LicensesNonProd[0].Licenses), nil
}
