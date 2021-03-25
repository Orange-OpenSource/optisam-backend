// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"

	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricAttrCounterStandard(ctx context.Context, req *v1.CreateMetricACS) (*v1.CreateMetricACS, error) {
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

func metricNameExistsACS(metrics []*repo.MetricACS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func serverToRepoMetricACS(met *v1.CreateMetricACS) *repo.MetricACS {
	return &repo.MetricACS{
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Value:         met.Value,
	}
}

func repoToServerMetricACS(met *repo.MetricACS) *v1.CreateMetricACS {
	return &v1.CreateMetricACS{
		ID:            met.ID,
		Name:          met.Name,
		EqType:        met.EqType,
		AttributeName: met.AttributeName,
		Value:         met.Value,
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
