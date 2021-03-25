// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesINM(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricINM(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesINM", zap.Error(err))
		return 0, status.Error(codes.Internal, "cannot fetch metric INM")
	}
	ind := metricNameExistsINM(metrics, input[METRIC_NAME].(string))
	if ind == -1 {
		return 0, status.Error(codes.NotFound, "cannot find metric name")
	}

	mat, err := computedMetricINM(metrics[ind])
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesINM - computedMetricINM - ", zap.Error(err))
		return 0, err
	}
	computedLicenses := uint64(0)

	if input[IS_AGG].(bool) {
		computedLicenses, err = s.licenseRepo.MetricINMComputedLicensesAgg(ctx, input[PROD_AGG_NAME].(string), input[METRIC_NAME].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricINMComputedLicenses(ctx, input[PROD_ID].(string), mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesINM - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric INM")

	}
	return computedLicenses, nil
}

func computedMetricINM(met *repo.MetricINM) (*repo.MetricINMComputed, error) {
	return &repo.MetricINMComputed{
		Name:        met.Name,
		Coefficient: met.Coefficient,
	}, nil
}

func metricNameExistsINM(metrics []*repo.MetricINM, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
