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

func (s *metricServiceServer) CreateMetricEquipAttrStandard(ctx context.Context, req *v1.MetricEquipAtt) (*v1.MetricEquipAtt, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - MetricEquipAttrStand - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricEquipAttrStand - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return nil, status.Error(codes.NotFound, "cannot find equipment type")
	}
	attr, err := validateEquipAttStandardMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return nil, err
	}
	met, err := s.metricRepo.CreateMetricEquipAttrStandard(ctx, serverToRepoMetricEquipAttr(req), attr, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricEquipAttrStandard - CreateMetricEquipAttrStand", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric acs")
	}

	return repoToServerMetricEquipAttr(met), nil
}

func validateEquipAttStandardMetric(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	if attrName == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute name is empty")
	}
	attr, err := attributeExistsByName(attributes, attrName)
	if err != nil {
		return nil, err
	}
	if attr.Type != repo.DataTypeInt {
		return nil, status.Error(codes.InvalidArgument, "only integer attributes are supported")
	}
	return attr, nil
}

func (s *metricServiceServer) UpdateMetricEquipAttrStandard(ctx context.Context, req *v1.MetricEquipAtt) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	_, err := s.metricRepo.GetMetricConfigEquipAttr(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricEquipAttr - repo/GetMetricConfigEquipAttr", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric attrsum")
	}
	eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricEquipAttr - fetching equipments", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EqType, eqTypes)
	if idx == -1 {
		return &v1.UpdateMetricResponse{}, status.Error(codes.NotFound, "cannot find equipment type")
	}
	_, err = validateEquipAttStandardMetric(eqTypes[idx].Attributes, req.AttributeName)
	if err != nil {
		return &v1.UpdateMetricResponse{}, err
	}
	err = s.metricRepo.UpdateMetricEquipAttr(ctx, &repo.MetricEquipAttrStand{
		Name:          req.Name,
		EqType:        req.EqType,
		AttributeName: req.AttributeName,
		Environment:   req.Environment,
		Value:         req.Value,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricEquipAttr - repo/UpdateMetricEquipAttr", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric attrsum")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoMetricEquipAttr(met *v1.MetricEquipAtt) *repo.MetricEquipAttrStand {
	return &repo.MetricEquipAttrStand{
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Environment:   met.Environment,
		Value:         met.Value,
	}
}

func repoToServerMetricEquipAttr(met *repo.MetricEquipAttrStand) *v1.MetricEquipAtt {
	return &v1.MetricEquipAtt{
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Environment:   met.Environment,
		Value:         met.Value,
	}
}

func (s *metricServiceServer) getDescriptionEquipAttr(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigEquipAttr(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricAttrSum", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric equip_attr")
	}
	des := repo.MetricDescriptionEquipAttrStandard.String()
	v := strings.Replace(des, "[equipment]", metric.EqType, 1)
	v = strings.Replace(v, "[attribute]", metric.AttributeName, 1)
	v = strings.Replace(v, "[environment]", metric.Environment, 1)
	v = strings.Replace(v, "[number]", fmt.Sprint(metric.Value), 1)
	return v, nil
}
