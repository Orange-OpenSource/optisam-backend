package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateMetricSAGProcessorStandard will create an sag.processor.standard metric
func (s *metricServiceServer) CreateMetricSAGProcessorStandard(ctx context.Context, req *v1.MetricSPS) (*v1.MetricSPS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching metrics", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}

	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	equipBase, err := equipmentTypeExistsByID(req.BaseEqTypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot find base level equipment type")
	}
	if error := validateAttributesSPS(equipBase.Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); error != nil {
		return nil, error
	}

	met, err := s.metricRepo.CreateMetricSPS(ctx, serverToRepoMetricSPS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricSAGProcessorStandard - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot create metric")
	}

	return repoToServerMetricSPS(met), nil
}

func (s *metricServiceServer) UpdateMetricSAGProcessorStandard(ctx context.Context, req *v1.MetricSPS) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if req.Default == true {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "Default Value True, Metric created by import can't be updated")
	}
	_, err := s.metricRepo.GetMetricConfigSPS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricSAGProcessorStandard - repo/GetMetricConfigSPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric sps")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricSAGProcessorStandard - repo/EquipmentTypes - fetching equipments", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	equipbase, err := equipmentTypeExistsByID(req.BaseEqTypeId, eqTypes)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.NotFound, "cannot find equipment type")
	}
	if e := validateAttributesSPS(equipbase.Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); e != nil {
		return &v1.UpdateMetricResponse{}, e
	}
	err = s.metricRepo.UpdateMetricSPS(ctx, &repo.MetricSPS{
		Name:             req.Name,
		NumCoreAttrID:    req.NumCoreAttrId,
		NumCPUAttrID:     req.NumCPUAttrId,
		BaseEqTypeID:     req.BaseEqTypeId,
		CoreFactorAttrID: req.CoreFactorAttrId,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricSAGProcessorStandard - repo/UpdateMetricSPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric sps")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoMetricSPS(met *v1.MetricSPS) *repo.MetricSPS {
	return &repo.MetricSPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrID:    met.NumCoreAttrId,
		NumCPUAttrID:     met.NumCPUAttrId,
		CoreFactorAttrID: met.CoreFactorAttrId,
		BaseEqTypeID:     met.BaseEqTypeId,
		Default:          met.Default,
	}
}

func repoToServerMetricSPS(met *repo.MetricSPS) *v1.MetricSPS {
	return &v1.MetricSPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrId:    met.NumCoreAttrID,
		NumCPUAttrId:     met.NumCPUAttrID,
		CoreFactorAttrId: met.CoreFactorAttrID,
		BaseEqTypeId:     met.BaseEqTypeID,
		Default:          met.Default,
	}
}

func validateAttributesSPS(attrs []*repo.Attribute, numCoreAttr string, numCPUAttr string, coreFactorAttr string) error {

	if numCoreAttr == "" {
		return status.Error(codes.InvalidArgument, "num of cores attribute is empty")
	}
	if numCPUAttr == "" {
		return status.Error(codes.InvalidArgument, "num of cpu attribute is empty")
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

	numOfCPU, err := attributeExists(attrs, numCPUAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "numofcpu attribute doesnt exists")
	}
	if numOfCPU.Type != repo.DataTypeInt && numOfCPU.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "numofcpu attribute doesnt have valid data type")
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
