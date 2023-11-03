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

func (s *metricServiceServer) CreateMetricUserConcurentStandard(ctx context.Context, req *v1.MetricUCS) (*v1.MetricUCS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricUserConcurentStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}
	met, err := s.metricRepo.CreateMetricUserConcurentStandard(ctx, serverToRepoUCS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricUserConcurentStandard  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric uns")
	}
	return repoToServerUCS(met), nil
}

func (s *metricServiceServer) UpdateMetricUserConcurentStandard(ctx context.Context, req *v1.MetricUCS) (*v1.UpdateMetricResponse, error) {
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
	_, err := s.metricRepo.GetMetricConfigConcurentUser(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricUserConcurentStandard - repo/GetMetricConfigConcurentUser", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric UCS")
	}
	err = s.metricRepo.UpdateMetricUCS(ctx, &repo.MetricUCS{
		Name:    req.Name,
		Profile: req.Profile,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricUserConcurentStandard - repo/UpdateMetricUCS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric UCS")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoUCS(met *v1.MetricUCS) *repo.MetricUCS {
	// des := repo.MetricDescriptionInstanceNumberStandard
	// v := strings.Replace(des, "number_of_deployments_authorized_licenses", met.NumOfDeployments)

	return &repo.MetricUCS{
		Name:    met.Name,
		Profile: met.Profile,
		Default: met.Default,
		//	Description: v,
	}
}

func repoToServerUCS(met *repo.MetricUCS) *v1.MetricUCS {
	return &v1.MetricUCS{
		Name:    met.Name,
		ID:      met.ID,
		Profile: met.Profile,
		Default: met.Default,
	}
}

func (s *metricServiceServer) getDescriptionUCS(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigConcurentUser(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUCS", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric User_conc")
	}
	des := repo.MetricDescriptionUserConcurentStandard.String()
	return strings.Replace(des, "[profile]", metric.Profile, 1), nil
}
