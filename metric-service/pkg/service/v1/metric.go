// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"encoding/json"
	"optisam-backend/common/optisam/helper"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"

	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// metricServiceServer is implementation of v1.authServiceServer proto interface
type metricServiceServer struct {
	metricRepo repo.Metric
}

// NewLicenseServiceServer creates License service
func NewMetricServiceServer(metricRepo repo.Metric) v1.MetricServiceServer {
	return &metricServiceServer{metricRepo: metricRepo}
}

func (s *metricServiceServer) ListMetricType(ctx context.Context, req *v1.ListMetricTypeRequest) (*v1.ListMetricTypeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metricTypes, err := s.metricRepo.ListMetricTypeInfo(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - ListMetricType - fetching metric types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types")
	}

	return &v1.ListMetricTypeResponse{
		Types: repoMetricTypeToServiceMetricTypeAll(metricTypes),
	}, nil
}

func (s *metricServiceServer) ListMetrices(ctx context.Context, req *v1.ListMetricRequest) (*v1.ListMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metricTypes, err := s.metricRepo.ListMetricTypeInfo(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - ListMetrices - fetching metric types info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types info")
	}

	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
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

func (s *metricServiceServer) GetMetricConfiguration(ctx context.Context, req *v1.GetMetricConfigurationRequest) (*v1.GetMetricConfigurationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if req.MetricInfo == nil || req.MetricInfo.Name == "" || req.MetricInfo.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "metric name and type can not be empty")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - GetMetricConfiguration - ListMetricOPS", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch OPS metrics")
	}
	idx := metricNameExistsAll(metrics, req.MetricInfo.Name)
	if idx == -1 {
		return nil, status.Error(codes.InvalidArgument, "metric does not exist")
	}
	if metrics[idx].Type.String() != req.MetricInfo.Type {
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}
	var metric interface{}
	switch metrics[idx].Type {
	case repo.MetricOPSOracleProcessorStandard:
		metric, err = s.metricRepo.GetMetricConfigOPS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ops")
		}
	case repo.MetricOracleNUPStandard:
		metric, err = s.metricRepo.GetMetricConfigNUP(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigNUP", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric nup")
		}
	case repo.MetricSPSSagProcessorStandard:
		metric, err = s.metricRepo.GetMetricConfigSPS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric sps")
		}
	case repo.MetricIPSIbmPvuStandard:
		metric, err = s.metricRepo.GetMetricConfigIPS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricIPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ips")
		}
	case repo.MetricAttrCounterStandard:
		metric, err = s.metricRepo.GetMetricConfigACS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricACS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric acs")
		}
	case repo.MetricInstanceNumberStandard:
		metric, err = s.metricRepo.GetMetricConfigINM(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricINM", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric inm")
		}
	}
	resMetric, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot marshal metric")
	}
	return &v1.GetMetricConfigurationResponse{
		MetricConfig: string(resMetric),
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

func repoMetricToServiceMetricAll(met []*repo.MetricInfo) []*v1.Metric {
	servMetric := make([]*v1.Metric, len(met))
	for i := range met {
		servMetric[i] = repoMetricToServiceMetric(met[i])
	}
	return servMetric
}

func repoMetricToServiceMetric(met *repo.MetricInfo) *v1.Metric {
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

func metricNameExistsAll(metrics []*repo.MetricInfo, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
