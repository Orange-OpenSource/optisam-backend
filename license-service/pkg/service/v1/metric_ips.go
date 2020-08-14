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

func init() {
	//admin rights are required for this function
	adminRpcMap["/v1.LicenseService/CreateMetricIBMPvuStandard"] = struct{}{}
}

func metricNameExistsIPS(metrics []*repo.MetricIPS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func (s *licenseServiceServer) computedLicensesIPS(ctx context.Context, eqTypes []*repo.EquipmentType, met string, cal func(mat *repo.MetricIPSComputed) (uint64, error)) (uint64, error) {
	userClaims, _ := ctxmanage.RetrieveClaims(ctx)
	metrics, err := s.licenseRepo.ListMetricIPS(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesIPS", zap.Error(err))
		return 0, status.Error(codes.Internal, "cannot fetch metric IPS")
	}
	ind := 0
	if ind = metricNameExistsIPS(metrics, met); ind == -1 {
		return 0, status.Error(codes.Internal, "cannot find metric name")
	}

	mat, err := computedMetricIPS(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesIPS - computedMetricIPS - ", zap.Error(err))
		return 0, err
	}
	computedLicenses, err := cal(mat)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesIPS - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric OPS")

	}
	return computedLicenses, nil

}

func computedMetricIPSWithName(met *repo.MetricIPS, eqTypes []*repo.EquipmentType, name string) (*repo.MetricIPSComputed, error) {
	metric, err := computedMetricIPS(met, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricIPSWithName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute IPS metric")
	}
	metric.Name = name
	return metric, nil
}

func computedMetricIPS(met *repo.MetricIPS, eqTypes []*repo.EquipmentType) (*repo.MetricIPSComputed, error) {
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

	return &repo.MetricIPSComputed{
		BaseType:       equipBase,
		NumCoresAttr:   numOfCores,
		CoreFactorAttr: coreFactor,
	}, nil
}
