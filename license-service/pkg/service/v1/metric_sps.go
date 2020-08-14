// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func metricNameExistsSPS(metrics []*repo.MetricSPS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func (s *licenseServiceServer) computedLicensesSPS(ctx context.Context, eqTypes []*repo.EquipmentType, met string, cal func(mat *repo.MetricSPSComputed) (uint64, uint64, error)) (uint64, uint64, error) {
	userClaims, _ := ctxmanage.RetrieveClaims(ctx)
	metrics, err := s.licenseRepo.ListMetricSPS(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot fetch metric SPS")

	}
	ind := 0
	if ind = metricNameExistsSPS(metrics, met); ind == -1 {
		return 0, 0, status.Error(codes.Internal, "cannot find metric name")
	}

	mat, err := computedMetricSPS(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - computedMetricSPS - ", zap.Error(err))
		return 0, 0, err
	}
	computedLicensesProd, computedLicensesNonProd, err := cal(mat)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot compute licenses for metric SPS")

	}
	return computedLicensesProd, computedLicensesNonProd, nil

}

func computedMetricSPSWithName(met *repo.MetricSPS, eqTypes []*repo.EquipmentType, name string) (*repo.MetricSPSComputed, error) {
	metric, err := computedMetricSPS(met, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricSPSWithName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute SPS metric")
	}
	metric.Name = name
	return metric, nil
}

func computedMetricSPS(met *repo.MetricSPS, eqTypes []*repo.EquipmentType) (*repo.MetricSPSComputed, error) {
	equipBase, err := equipmentTypeExistsByID(met.BaseEqTypeID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - equipmentTypeExistsByID - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot find base level equipment type")
	}
	numOfCores, err := attributeExists(equipBase.Attributes, met.NumCoreAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - attributeExists - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	coreFactor, err := attributeExists(equipBase.Attributes, met.CoreFactorAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - attributeExists - corefactor", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}

	return &repo.MetricSPSComputed{
		BaseType:       equipBase,
		NumCoresAttr:   numOfCores,
		CoreFactorAttr: coreFactor,
	}, nil
}
