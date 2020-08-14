// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"log"

	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesINM(ctx context.Context, eqTypes []*repo.EquipmentType, met string, cal func(mat *repo.MetricINMComputed) (uint64, error)) (uint64, error) {
	userClaims, _ := ctxmanage.RetrieveClaims(ctx)
	metrics, err := s.licenseRepo.ListMetricINM(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesINM", zap.Error(err))
		return 0, status.Error(codes.Internal, "cannot fetch metric INM")
	}
	ind := metricNameExistsINM(metrics, met)
	if ind == -1 {
		return 0, status.Error(codes.NotFound, "cannot find metric name")
	}

	mat, err := computedMetricINM(metrics[ind])
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesINM - computedMetricINM - ", zap.Error(err))
		return 0, err
	}
	log.Printf("METRIC to be computated %+v", mat)
	computedLicenses, err := cal(mat)
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
