package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateMetricIBMPvuStandard will create an IBM.pvu.standard metric
func (s *metricServiceServer) CreateMetricIBMPvuStandard(ctx context.Context, req *v1.MetricIPS) (*v1.MetricIPS, error) {
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
	if error := validateAttributesIPS(equipBase.Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); error != nil {
		return nil, error
	}

	met, err := s.metricRepo.CreateMetricIPS(ctx, serverToRepoMetricIPS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricSAGProcessorStandard - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot create metric")
	}

	return repoToServerMetricIPS(met), nil

}

func (s *metricServiceServer) UpdateMetricIBMPvuStandard(ctx context.Context, req *v1.MetricIPS) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	_, err := s.metricRepo.GetMetricConfigIPS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricIBMPvuStandard - repo/GetMetricConfigIPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric ips")
	}
	eqTypes, er := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if er != nil {
		logger.Log.Error("service/v1 - UpdateMetricIBMPvuStandard - repo/EquipmentTypes - fetching equipments", zap.String("reason", er.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	equipbase, err := equipmentTypeExistsByID(req.BaseEqTypeId, eqTypes)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.NotFound, "cannot find equipment type")
	}
	if e := validateAttributesIPS(equipbase.Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); e != nil {
		return &v1.UpdateMetricResponse{}, e
	}
	if err := s.metricRepo.UpdateMetricIPS(ctx, &repo.MetricIPS{
		Name:             req.Name,
		NumCoreAttrID:    req.NumCoreAttrId,
		NumCPUAttrID:     req.NumCPUAttrId,
		BaseEqTypeID:     req.BaseEqTypeId,
		CoreFactorAttrID: req.CoreFactorAttrId,
	}, req.GetScopes()[0]); err != nil {
		logger.Log.Error("service/v1 - UpdateMetricIBMPvuStandard - repo/UpdateMetricIPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric ips")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoMetricIPS(met *v1.MetricIPS) *repo.MetricIPS {
	return &repo.MetricIPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrID:    met.NumCoreAttrId,
		NumCPUAttrID:     met.NumCPUAttrId,
		CoreFactorAttrID: met.CoreFactorAttrId,
		BaseEqTypeID:     met.BaseEqTypeId,
	}
}

func repoToServerMetricIPS(met *repo.MetricIPS) *v1.MetricIPS {
	return &v1.MetricIPS{
		ID:               met.ID,
		Name:             met.Name,
		NumCoreAttrId:    met.NumCoreAttrID,
		NumCPUAttrId:     met.NumCPUAttrID,
		CoreFactorAttrId: met.CoreFactorAttrID,
		BaseEqTypeId:     met.BaseEqTypeID,
	}
}

func validateAttributesIPS(attr []*repo.Attribute, numCoreAttr string, numCPUAttr string, coreFactorAttr string) error {

	if numCoreAttr == "" {
		return status.Error(codes.InvalidArgument, "num of cores attribute is empty")
	}

	if numCPUAttr == "" {
		return status.Error(codes.InvalidArgument, "num of cpu attribute is empty")
	}

	if coreFactorAttr == "" {
		return status.Error(codes.InvalidArgument, "core factor attribute is empty")
	}

	numOfCores, err := attributeExists(attr, numCoreAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "numofcores attribute doesnt exists")
	}
	if numOfCores.Type != repo.DataTypeInt && numOfCores.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "numofcores attribute doesnt have valid data type")
	}

	numOfCPU, err := attributeExists(attr, numCPUAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "numofcpu attribute doesnt exists")
	}
	if numOfCPU.Type != repo.DataTypeInt && numOfCPU.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "numofcpu attribute doesnt have valid data type")
	}

	coreFactor, err := attributeExists(attr, coreFactorAttr)
	if err != nil {

		return status.Error(codes.InvalidArgument, "corefactor attribute doesnt exists")
	}

	if coreFactor.Type != repo.DataTypeInt && coreFactor.Type != repo.DataTypeFloat {
		return status.Error(codes.InvalidArgument, "corefactor attribute doesnt have valid data type")
	}
	return nil
}

func equipmentTypeExistsByID(id string, eqTypes []*repo.EquipmentType) (*repo.EquipmentType, error) {
	for _, eqt := range eqTypes {
		if eqt.ID == id {
			return eqt, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "equipment not exists")
}
