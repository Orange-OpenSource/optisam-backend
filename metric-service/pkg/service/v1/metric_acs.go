package v1

import (
	"context"
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricAttrCounterStandard(ctx context.Context, req *v1.MetricACS) (*v1.MetricACS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricAttrCounterStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 -CreateMetricAttrCounterStandard - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return nil, status.Error(codes.NotFound, "cannot find equipment type")
	}
	attr, err := validateAttributeACSMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return nil, err
	}
	err = attr.ValidateAttrValFromString(req.Value)
	if err != nil {
		return nil, err
	}
	met, err := s.metricRepo.CreateMetricACS(ctx, serverToRepoMetricACS(req), attr, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricAttrCounterStandard - CreateMetricACS", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric acs")
	}

	return repoToServerMetricACS(met), nil
}

func (s *metricServiceServer) UpdateMetricAttrCounterStandard(ctx context.Context, req *v1.MetricACS) (*v1.UpdateMetricResponse, error) {
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
	_, err := s.metricRepo.GetMetricConfigACS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricACS - repo/GetMetricConfigACS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric attrcounter")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricACS - fetching equipments", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return &v1.UpdateMetricResponse{}, status.Error(codes.NotFound, "cannot find equipment type")
	}
	attr, err := validateAttributeACSMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return &v1.UpdateMetricResponse{}, err
	}
	err = attr.ValidateAttrValFromString(req.Value)
	if err != nil {
		return &v1.UpdateMetricResponse{}, err
	}
	err = s.metricRepo.UpdateMetricACS(ctx, &repo.MetricACS{
		Name:          req.Name,
		EqType:        req.EqType,
		AttributeName: req.AttributeName,
		Value:         req.Value,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricAttributeSum - repo/UpdateMetricAttrSum", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric attrsum")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func validateAttributeACSMetric(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	if attrName == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute name is empty")
	}
	attr, err := attributeExistsByName(attributes, attrName)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

func attributeExistsByName(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	for _, attr := range attributes {
		if attr.Name == attrName {
			return attr, nil
		}
	}
	return nil, status.Error(codes.InvalidArgument, "attribute does not exists")
}

func serverToRepoMetricACS(met *v1.MetricACS) *repo.MetricACS {
	return &repo.MetricACS{
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Value:         met.Value,
		Default:       met.Default,
	}
}

func repoToServerMetricACS(met *repo.MetricACS) *v1.MetricACS {
	return &v1.MetricACS{
		ID:            met.ID,
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Value:         met.Value,
		Default:       met.Default,
	}
}

func equipmentTypeExistsByType(eqType string, eqTypes []*repo.EquipmentType) int {
	for i := 0; i < len(eqTypes); i++ {
		if eqTypes[i].Type == eqType {
			return i
		}
	}
	return -1
}

func (s *metricServiceServer) getDescriptionACS(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigACS(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricACS", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric acs")
	}
	des := repo.MetricDescriptionAttrCounterStandard.String()
	v := strings.Replace(des, "specific_type", metric.EqType, 1)
	v = strings.Replace(v, "specific_attribute", metric.AttributeName, 1)
	v = strings.Replace(v, "value", metric.Value, 1)
	return v, nil
}
