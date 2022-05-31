package v1

import (
	"context"
	"strconv"
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

func (s *metricServiceServer) CreateMetricStaticStandard(ctx context.Context, req *v1.MetricSS) (*v1.MetricSS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricStaticStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricStaticStandard(ctx, serverToRepoSS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric ss")
	}

	return repoToServerSS(met), nil
}

func (s *metricServiceServer) UpdateMetricStaticStandard(ctx context.Context, req *v1.MetricSS) (*v1.UpdateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return &v1.UpdateMetricResponse{}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	_, err := s.metricRepo.GetMetricConfigSS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricStaticStandard - repo/GetMetricConfigSS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric ss")
	}
	err = s.metricRepo.UpdateMetricSS(ctx, &repo.MetricSS{
		Name:           req.Name,
		ReferenceValue: req.ReferenceValue,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricStaticStandard - repo/UpdateMetricSS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric ss")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoSS(met *v1.MetricSS) *repo.MetricSS {
	// des := repo.MetricDescriptionInstanceNumberStandard
	// v := strings.Replace(des, "number_of_deployments_authorized_licenses", met.NumOfDeployments)

	return &repo.MetricSS{
		Name:           met.Name,
		ReferenceValue: met.ReferenceValue,
		//	Description: v,
	}
}

func repoToServerSS(met *repo.MetricSS) *v1.MetricSS {
	return &v1.MetricSS{
		Name:           met.Name,
		ID:             met.ID,
		ReferenceValue: met.ReferenceValue,
	}
}

func (s *metricServiceServer) getDescriptionSS(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigSS(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSS", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric ss")
	}
	des := repo.MetricDescriptionStaticStandard.String()
	return strings.Replace(des, "Reference_value", strconv.Itoa(int(metric.ReferenceValue)), 1), nil
}
