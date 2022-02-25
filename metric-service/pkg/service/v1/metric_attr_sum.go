package v1

import (
	"context"
	"fmt"
	"strings"

	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricAttrSumStandard(ctx context.Context, req *v1.MetricAttrSum) (*v1.MetricAttrSum, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - MetricAttrSumStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricAttrSumStandard - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return nil, status.Error(codes.NotFound, "cannot find equipment type")
	}
	attr, err := validateAttributeASSMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return nil, err
	}
	met, err := s.metricRepo.CreateMetricAttrSum(ctx, serverToRepoMetricAttrSum(req), attr, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricAttrSumStandard - CreateMetricAttrSum", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric acs")
	}

	return repoToServerMetricAttrSum(met), nil
}

func validateAttributeASSMetric(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	if attrName == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute name is empty")
	}
	attr, err := attributeExistsByName(attributes, attrName)
	if err != nil {
		return nil, err
	}
	if attr.Type == repo.DataTypeString {
		return nil, status.Error(codes.InvalidArgument, "only float and integer attributes are supported")
	}
	return attr, nil
}

func (s *metricServiceServer) UpdateMetricAttrSumStandard(ctx context.Context, req *v1.MetricAttrSum) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	_, err := s.metricRepo.GetMetricConfigAttrSum(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricAttributeSum - repo/GetMetricConfigAttrSum", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric attrsum")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricAttributeSum - fetching equipments", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return &v1.UpdateMetricResponse{}, status.Error(codes.NotFound, "cannot find equipment type")
	}
	_, err = validateAttributeASSMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return &v1.UpdateMetricResponse{}, err
	}
	err = s.metricRepo.UpdateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
		Name:           req.Name,
		EqType:         req.EqType,
		AttributeName:  req.AttributeName,
		ReferenceValue: req.ReferenceValue,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricAttributeSum - repo/UpdateMetricAttrSum", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric attrsum")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoMetricAttrSum(met *v1.MetricAttrSum) *repo.MetricAttrSumStand {
	return &repo.MetricAttrSumStand{
		Name:           met.Name,
		EqType:         met.EqType,
		AttributeName:  met.AttributeName,
		ReferenceValue: met.ReferenceValue,
	}
}

func repoToServerMetricAttrSum(met *repo.MetricAttrSumStand) *v1.MetricAttrSum {
	return &v1.MetricAttrSum{
		ID:             met.ID,
		Name:           met.Name,
		EqType:         met.EqType,
		AttributeName:  met.AttributeName,
		ReferenceValue: met.ReferenceValue,
	}
}

func (s *metricServiceServer) getDescriptionAttSum(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigAttrSum(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricAttrSum", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric att")
	}
	des := repo.MetricDescriptionAttrSumStandard.String()
	v := strings.Replace(des, "Equipment_type", metric.EqType, 1)
	v = strings.Replace(v, "attribute_value", metric.AttributeName, 1)
	v = strings.Replace(v, "Reference_value", fmt.Sprintf("%.2f", metric.ReferenceValue), 1)
	return v, nil
}
