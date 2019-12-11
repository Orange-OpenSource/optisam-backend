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
	"optisam-backend/common/optisam/strcomp"

	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListMetricType(ctx context.Context, req *v1.ListMetricTypeRequest) (*v1.ListMetricTypeResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	metricTypes, err := s.licenseRepo.ListMetricTypeInfo(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListMetricType - fetching metric types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types")
	}

	return &v1.ListMetricTypeResponse{
		Types: repoMetricTypeToServiceMetricTypeAll(metricTypes),
	}, nil
}

func (s *licenseServiceServer) ListMetrices(ctx context.Context, req *v1.ListMetricRequest) (*v1.ListMetricResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	metricTypes, err := s.licenseRepo.ListMetricTypeInfo(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListMetrices - fetching metric types info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types info")
	}

	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListMetrices - fetching metric types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types")
	}

	metricsList := repoMetricToServiceMetricAll(metrics)
	for _, met := range metricsList {
		desc, err := discriptionMetric(met.Type, metricTypes)
		if err != nil {
			logger.Log.Error("service/v1 - GetEquipment - fetching equipment", zap.String("reason", err.Error()))
			continue
		}
		met.Description = desc
	}

	return &v1.ListMetricResponse{
		Metrices: metricsList,
	}, nil

}

func repoMetricTypeToServiceMetricTypeAll(met []*repo.MetricTypeInfo) []*v1.MetricType {
	servMetrics := make([]*v1.MetricType, len(met))
	for i := range met {
		servMetrics[i] = repoMetricTypeToServiceMetricType(met[i])
	}
	return servMetrics
}

func repoMetricTypeToServiceMetricType(met *repo.MetricTypeInfo) *v1.MetricType {
	return &v1.MetricType{
		Name:        string(met.Name),
		Description: met.Description,
		Href:        met.Href,
		TypeId:      v1.MetricType_Type(met.MetricType),
	}
}

func repoMetricToServiceMetricAll(met []*repo.Metric) []*v1.Metric {
	servMetric := make([]*v1.Metric, len(met))
	for i := range met {
		servMetric[i] = repoMetricToServiceMetric(met[i])
	}
	return servMetric
}

func repoMetricToServiceMetric(met *repo.Metric) *v1.Metric {
	return &v1.Metric{
		Name: met.Name,
		Type: string(met.Type),
	}
}

func discriptionMetric(typ string, metrics []*repo.MetricTypeInfo) (string, error) {
	for _, met := range metrics {
		if (met.Name).String() == typ {
			return met.Description, nil
		}
	}
	return "", status.Error(codes.Internal, "description not found - "+typ)
}

func metricNameExistsAll(metrics []*repo.Metric, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
