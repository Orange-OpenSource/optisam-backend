// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateMetricSAGProcessorStandard will create an sag.processor.standard metric
func (s *licenseServiceServer) CreateMetricSAGProcessorStandard(ctx context.Context, req *v1.CreateMetricSPS) (*v1.CreateMetricSPS, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching metrics", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}

	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	equipBase, err := equipmentTypeExistsByID(req.BaseEqTypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot find base level equipment type")
	}
	if err := validateAttributesSPS(equipBase.Attributes, req.NumCoreAttrId, req.CoreFactorAttrId); err != nil {
		return nil, err
	}

	met, err := s.licenseRepo.CreateMetricSPS(ctx, serverToRepoMetricSPS(req), userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricSAGProcessorStandard - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot create metric")
	}

	return repoToServerMetricSPS(met), nil

}

func serverToRepoMetricSPS(met *v1.CreateMetricSPS) *repo.MetricSPS {
	return &repo.MetricSPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrID:    met.NumCoreAttrId,
		CoreFactorAttrID: met.CoreFactorAttrId,
		BaseEqTypeID:     met.BaseEqTypeId,
	}
}

func repoToServerMetricSPS(met *repo.MetricSPS) *v1.CreateMetricSPS {
	return &v1.CreateMetricSPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrId:    met.NumCoreAttrID,
		CoreFactorAttrId: met.CoreFactorAttrID,
		BaseEqTypeId:     met.BaseEqTypeID,
	}
}

func validateAttributesSPS(attrs []*repo.Attribute, numCoreAttr string, coreFactorAttr string) error {

	if numCoreAttr == "" {
		return status.Error(codes.InvalidArgument, "num of cores attribute is empty")
	}

	if coreFactorAttr == "" {
		return status.Error(codes.InvalidArgument, "core factor attribute is empty")
	}

	numOfCores, err := attributeExists(attrs, numCoreAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "numofcores attribute doesnt exists")
	}
	if numOfCores.Type != repo.DataTypeInt && numOfCores.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "numofcores attribute doesnt have valid data type")
	}

	coreFactor, err := attributeExists(attrs, coreFactorAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "corefactor attribute doesnt exists")
	}

	if coreFactor.Type != repo.DataTypeInt && coreFactor.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "corefactor attribute doesnt have valid data type")
	}
	return nil
}
func metricNameExistsSPS(metrics []*repo.MetricSPS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func (s *licenseServiceServer) computedLicensesSPS(ctx context.Context, eqTypes []*repo.EquipmentType, ID string, met string) (uint64, uint64, error) {
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

	equipBase, err := equipmentTypeExistsByID(metrics[ind].BaseEqTypeID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot find base level equipment type")
	}
	numOfCores, err := attributeExists(equipBase.Attributes, metrics[ind].NumCoreAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	coreFactor, err := attributeExists(equipBase.Attributes, metrics[ind].CoreFactorAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}

	mat := &repo.MetricSPSComputed{
		BaseType:       equipBase,
		NumCoresAttr:   numOfCores,
		CoreFactorAttr: coreFactor,
	}
	computedLicensesProd, computedLicensesNonProd, err := s.licenseRepo.MetricSPSComputedLicenses(ctx, ID, mat, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot compute licenses for metric SPS")

	}
	return computedLicensesProd, computedLicensesNonProd, nil

}
