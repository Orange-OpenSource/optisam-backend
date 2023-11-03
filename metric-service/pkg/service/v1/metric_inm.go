package v1

import (
	"context"
	"strconv"
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

func (s *metricServiceServer) CreateMetricInstanceNumberStandard(ctx context.Context, req *v1.MetricINM) (*v1.MetricINM, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricInstanceNumberStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricInstanceNumberStandard(ctx, serverToRepoINM(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricInstanceNumberStandard  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric inm")
	}

	return repoToServerINM(met), nil
}

func (s *metricServiceServer) UpdateMetricInstanceNumberStandard(ctx context.Context, req *v1.MetricINM) (*v1.UpdateMetricResponse, error) {
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
	_, err := s.metricRepo.GetMetricConfigINM(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricInstanceNumberStandard - repo/GetMetricConfigINM", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric inm")
	}
	err = s.metricRepo.UpdateMetricINM(ctx, &repo.MetricINM{
		Name:        req.Name,
		Coefficient: req.NumOfDeployments,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricInstanceNumberStandard - repo/UpdateMetricINM", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric inm")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoINM(met *v1.MetricINM) *repo.MetricINM {
	// des := repo.MetricDescriptionInstanceNumberStandard
	// v := strings.Replace(des, "number_of_deployments_authorized_licenses", met.NumOfDeployments)

	return &repo.MetricINM{
		Name:        met.Name,
		Coefficient: met.NumOfDeployments,
		Default:     met.Default,
		//	Description: v,
	}
}

func repoToServerINM(met *repo.MetricINM) *v1.MetricINM {
	return &v1.MetricINM{
		Name:             met.Name,
		ID:               met.ID,
		NumOfDeployments: met.Coefficient,
		Default:          met.Default,
	}
}

func (s *metricServiceServer) getDescriptionINM(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigINM(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricINM", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric inm")
	}
	des := repo.MetricDescriptionInstanceNumberStandard.String()
	return strings.Replace(des, "number_of_deployments_authorized_licenses", strconv.Itoa(int(metric.Coefficient)), 1), nil
}
