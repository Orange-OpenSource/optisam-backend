package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricOracleNUPStandard(ctx context.Context, req *v1.MetricNUP) (*v1.MetricNUP, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if req.StartEqTypeId == "" {
		return nil, status.Error(codes.InvalidArgument, "start level is empty")
	}

	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching metrics", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}

	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	// if metricTypeExistsAll(metrics, repo.MetricOracleNUPStandard) != -1 {
	// 	return nil, status.Error(codes.InvalidArgument, "metric NUP already exists")
	// }
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricOracleProcessorStandard - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	parAncestors, err := parentHierarchy(eqTypes, req.StartEqTypeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "parent hierarchy doesnt exists")
	}

	baseLevelIdx, err := validateLevelsNew(parAncestors, 0, req.BaseEqTypeId)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find base level equipment type in parent hierarchy")
	}
	aggLevelIdx, err := validateLevelsNew(parAncestors, baseLevelIdx, req.AggerateLevelEqTypeId)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find aggregate level equipment type in parent hierarchy")
	}
	_, err = validateLevelsNew(parAncestors, aggLevelIdx, req.EndEqTypeId)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find end level equipment type in parent hierarchy")
	}

	if error := validateAttributesOracleNUP(parAncestors[baseLevelIdx].Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); error != nil {
		return nil, error
	}

	// for checking TransformMetricName already exists on system or not
	if req.Transform {
		if req.TransformMetricName == "" {
			return nil, status.Error(codes.Internal, "transform metric name cannot be empty")
		}

		_, err = s.metricRepo.GetMetricConfigOPS(ctx, req.TransformMetricName, req.GetScopes()[0])
		if err != nil {
			if err == repo.ErrNoData {
				return nil, status.Error(codes.InvalidArgument, "metric does not exist")
			}
			logger.Log.Error("service/v1 -CreateMetricOracleProcessorStandard - repo/GetMetricConfigOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ops")
		}

	} else if req.TransformMetricName != "" {
		return nil, status.Error(codes.Internal, "transform metric name should be empty")
	}

	met, err := s.metricRepo.CreateMetricOracleNUPStandard(ctx, serverToRepoMetricOracleNUP(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricOracleProcessorStandard - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot create metric")
	}

	return repoToServerMetricOracleNUP(met), nil

}

func (s *metricServiceServer) UpdateMetricOracleNUPStandard(ctx context.Context, req *v1.MetricNUP) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if req.StartEqTypeId == "" {
		return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "start level is empty")
	}
	_, err := s.metricRepo.GetMetricConfigNUP(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricOracleNUPStandard - repo/GetMetricConfigNUP", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric nup")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricOracleNUPStandard - repo/EquipmentTypes - fetching equipments", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	parAncestors, err := parentHierarchy(eqTypes, req.StartEqTypeId)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "parent hierarchy doesnt exists")
	}

	baseLevelIdx, err := validateLevelsNew(parAncestors, 0, req.BaseEqTypeId)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find base level equipment type in parent hierarchy")
	}
	aggLevelIdx, err := validateLevelsNew(parAncestors, baseLevelIdx, req.AggerateLevelEqTypeId)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find aggregate level equipment type in parent hierarchy")
	}
	_, err = validateLevelsNew(parAncestors, aggLevelIdx, req.EndEqTypeId)
	if err != nil {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find end level equipment type in parent hierarchy")
	}

	if e := validateAttributesOracleNUP(parAncestors[baseLevelIdx].Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); e != nil {
		return &v1.UpdateMetricResponse{}, e
	}

	// for checking TransformMetricName already exists on system or not
	if req.Transform {

		if req.TransformMetricName == "" {
			return nil, status.Error(codes.Internal, "transform metric name cannot be empty")
		}

		_, err = s.metricRepo.GetMetricConfigOPS(ctx, req.TransformMetricName, req.GetScopes()[0])
		if err != nil {
			if err == repo.ErrNoData {
				return nil, status.Error(codes.InvalidArgument, "metric does not exist")
			}
			logger.Log.Error("service/v1 -CreateMetricOracleProcessorStandard - repo/GetMetricConfigOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ops")
		}

	} else if req.TransformMetricName != "" {
		return nil, status.Error(codes.Internal, "transform metric name should be empty")
	}

	err = s.metricRepo.UpdateMetricNUP(ctx, &repo.MetricNUPOracle{
		Name:                  req.Name,
		NumCoreAttrID:         req.NumCoreAttrId,
		NumCPUAttrID:          req.NumCPUAttrId,
		StartEqTypeID:         req.StartEqTypeId,
		BaseEqTypeID:          req.BaseEqTypeId,
		CoreFactorAttrID:      req.CoreFactorAttrId,
		AggerateLevelEqTypeID: req.AggerateLevelEqTypeId,
		EndEqTypeID:           req.EndEqTypeId,
		NumberOfUsers:         req.NumberOfUsers,
		Transform:             req.Transform,
		TransformMetricName:   req.TransformMetricName,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricOracleNUPStandard - repo/UpdateMetricNUP", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric nup")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func validateAttributesOracleNUP(attr []*repo.Attribute, numCoreAttr string, numCPUAttr string, coreFactorAttr string) error {
	// TODO : remove duplicacy (duplicate with validateAttributesOPS)
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

func serverToRepoMetricOracleNUP(met *v1.MetricNUP) *repo.MetricNUPOracle {
	return &repo.MetricNUPOracle{
		ID:                    met.ID,
		Name:                  met.Name,
		NumCoreAttrID:         met.NumCoreAttrId,
		NumCPUAttrID:          met.NumCPUAttrId,
		CoreFactorAttrID:      met.CoreFactorAttrId,
		StartEqTypeID:         met.StartEqTypeId,
		BaseEqTypeID:          met.BaseEqTypeId,
		AggerateLevelEqTypeID: met.AggerateLevelEqTypeId,
		EndEqTypeID:           met.EndEqTypeId,
		NumberOfUsers:         met.NumberOfUsers,
		Transform:             met.Transform,
		TransformMetricName:   met.TransformMetricName,
	}
}

func repoToServerMetricOracleNUP(met *repo.MetricNUPOracle) *v1.MetricNUP {
	return &v1.MetricNUP{
		ID:                    met.ID,
		Name:                  met.Name,
		NumCoreAttrId:         met.NumCoreAttrID,
		NumCPUAttrId:          met.NumCPUAttrID,
		CoreFactorAttrId:      met.CoreFactorAttrID,
		StartEqTypeId:         met.StartEqTypeID,
		BaseEqTypeId:          met.BaseEqTypeID,
		AggerateLevelEqTypeId: met.AggerateLevelEqTypeID,
		EndEqTypeId:           met.EndEqTypeID,
		NumberOfUsers:         met.NumberOfUsers,
		Transform:             met.Transform,
		TransformMetricName:   met.TransformMetricName,
	}
}

func (s *metricServiceServer) getDescriptionNUP(ctx context.Context, name, scope string) (string, error) {

	metric, err := s.metricRepo.GetMetricConfigNUP(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricNUP", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric nup")
	}
	des := repo.MetricDescriptionOracleNUPStandard.String()
	return strings.Replace(des, "given_users", strconv.FormatUint(uint64(metric.NumberOfUsers), 10), 1), nil
}
