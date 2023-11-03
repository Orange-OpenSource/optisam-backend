package v1

import (
	"context"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricOracleProcessorStandard(ctx context.Context, req *v1.MetricOPS) (*v1.MetricOPS, error) {
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

	// if metricTypeExistsAll(metrics, repo.MetricOPSOracleProcessorStandard) != -1 {
	// 	return nil, status.Error(codes.InvalidArgument, "metric ops already exists")
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

	if error := validateAttributesOPS(parAncestors[baseLevelIdx].Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); error != nil {
		return nil, error
	}

	met, err := s.metricRepo.CreateMetricOPS(ctx, serverToRepoMetricOPS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricOracleProcessorStandard - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot create metric")
	}

	return repoToServerMetricOPS(met), nil

}

func (s *metricServiceServer) UpdateMetricOracleProcessorStandard(ctx context.Context, req *v1.MetricOPS) (*v1.UpdateMetricResponse, error) {
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
	if req.StartEqTypeId == "" {
		return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "start level is empty")
	}
	_, err := s.metricRepo.GetMetricConfigOPS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricOracleProcessorStandard - repo/GetMetricConfigOPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric ops")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricOracleProcessorStandard - repo/EquipmentTypes - fetching equipments", zap.String("reason", err.Error()))
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

	if e := validateAttributesOPS(parAncestors[baseLevelIdx].Attributes, req.NumCoreAttrId, req.NumCPUAttrId, req.CoreFactorAttrId); e != nil {
		return &v1.UpdateMetricResponse{}, e
	}
	err = s.metricRepo.UpdateMetricOPS(ctx, &repo.MetricOPS{
		Name:                  req.Name,
		NumCoreAttrID:         req.NumCoreAttrId,
		NumCPUAttrID:          req.NumCPUAttrId,
		StartEqTypeID:         req.StartEqTypeId,
		BaseEqTypeID:          req.BaseEqTypeId,
		CoreFactorAttrID:      req.CoreFactorAttrId,
		AggerateLevelEqTypeID: req.AggerateLevelEqTypeId,
		EndEqTypeID:           req.EndEqTypeId,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricOracleProcessorStandard - repo/UpdateMetricOPS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric ops")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func parentHierarchy(eqTypes []*repo.EquipmentType, startID string) ([]*repo.EquipmentType, error) {
	equip, err := equipmentTypeExistsByID(startID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}
	ancestors := []*repo.EquipmentType{}
	ancestors = append(ancestors, equip)
	parID := equip.ParentID
	for parID != "" {
		equipAnc, err := equipmentTypeExistsByID(parID, eqTypes)
		if err != nil {
			logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "parent hierarchy not found")
		}
		ancestors = append(ancestors, equipAnc)
		parID = equipAnc.ParentID
	}
	return ancestors, nil
}

func attributeExists(attributes []*repo.Attribute, attrID string) (*repo.Attribute, error) {
	for _, attr := range attributes {
		if attr.ID == attrID {
			return attr, nil
		}
	}
	return nil, status.Errorf(codes.Unknown, "attribute not exists")
}

func validateAttributesOPS(attr []*repo.Attribute, numCoreAttr string, numCPUAttr string, coreFactorAttr string) error {

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

func serverToRepoMetricOPS(met *v1.MetricOPS) *repo.MetricOPS {
	return &repo.MetricOPS{
		ID:                    met.ID,
		Name:                  met.Name,
		NumCoreAttrID:         met.NumCoreAttrId,
		NumCPUAttrID:          met.NumCPUAttrId,
		CoreFactorAttrID:      met.CoreFactorAttrId,
		StartEqTypeID:         met.StartEqTypeId,
		BaseEqTypeID:          met.BaseEqTypeId,
		AggerateLevelEqTypeID: met.AggerateLevelEqTypeId,
		EndEqTypeID:           met.EndEqTypeId,
		Default:               met.Default,
	}
}

func repoToServerMetricOPS(met *repo.MetricOPS) *v1.MetricOPS {
	return &v1.MetricOPS{

		ID:                    met.ID,
		Name:                  met.Name,
		NumCoreAttrId:         met.NumCoreAttrID,
		NumCPUAttrId:          met.NumCPUAttrID,
		CoreFactorAttrId:      met.CoreFactorAttrID,
		StartEqTypeId:         met.StartEqTypeID,
		BaseEqTypeId:          met.BaseEqTypeID,
		AggerateLevelEqTypeId: met.AggerateLevelEqTypeID,
		EndEqTypeId:           met.EndEqTypeID,
		Default:               met.Default,
	}
}

func validateLevelsNew(levels []*repo.EquipmentType, startIdx int, base string) (int, error) {
	for i := startIdx; i < len(levels); i++ {
		if levels[i].ID == base {
			return i, nil
		}
	}
	return -1, errors.New("not found")
}
