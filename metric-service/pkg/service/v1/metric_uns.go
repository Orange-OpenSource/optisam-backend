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

func (s *metricServiceServer) CreateMetricUserNominativeStandard(ctx context.Context, req *v1.MetricUNS) (*v1.MetricUNS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricUserNominativeStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricUserNominativeStandard(ctx, serverToRepoUNS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricUserNominativeStandard  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric UNS")
	}

	return repoToServerUNS(met), nil
}

func (s *metricServiceServer) UpdateMetricUserNominativeStandard(ctx context.Context, req *v1.MetricUNS) (*v1.UpdateMetricResponse, error) {
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
	_, err := s.metricRepo.GetMetricConfigUNS(ctx, req.Name, req.GetScopes()[0])
	if err != nil {
		if err == repo.ErrNoData {
			return &v1.UpdateMetricResponse{}, status.Error(codes.InvalidArgument, "metric does not exist")
		}
		logger.Log.Error("service/v1 -UpdateMetricUserNominativeStandard - repo/GetMetricConfigUNS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot fetch metric UNS")
	}
	err = s.metricRepo.UpdateMetricUNS(ctx, &repo.MetricUNS{
		Name:    req.Name,
		Profile: req.Profile,
	}, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - UpdateMetricUserNominativeStandard - repo/UpdateMetricUNS", zap.String("reason", err.Error()))
		return &v1.UpdateMetricResponse{}, status.Error(codes.Internal, "cannot update metric UNS")
	}

	return &v1.UpdateMetricResponse{
		Success: true,
	}, nil
}

func serverToRepoUNS(met *v1.MetricUNS) *repo.MetricUNS {
	// des := repo.MetricDescriptionInstanceNumberStandard
	// v := strings.Replace(des, "number_of_deployments_authorized_licenses", met.NumOfDeployments)

	return &repo.MetricUNS{
		Name:    met.Name,
		Profile: met.Profile,
		Default: met.Default,
		//	Description: v,
	}
}

func repoToServerUNS(met *repo.MetricUNS) *v1.MetricUNS {
	return &v1.MetricUNS{
		Name:    met.Name,
		ID:      met.ID,
		Profile: met.Profile,
		Default: met.Default,
	}
}

func (s *metricServiceServer) getDescriptionUNS(ctx context.Context, name, scope string) (string, error) {
	metric, err := s.metricRepo.GetMetricConfigUNS(ctx, name, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUNS", zap.String("reason", err.Error()))
		return "", status.Error(codes.Internal, "cannot fetch metric inm")
	}
	des := repo.MetricDescriptionUserNomStandard.String()
	return strings.Replace(des, "[profile]", metric.Profile, 1), nil
}
