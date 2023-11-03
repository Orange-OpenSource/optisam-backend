package v1

import (
	"context"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricUserSumStandard(ctx context.Context, req *v1.MetricUSS) (*v1.MetricUSS, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	scopeinfo, err := s.account.GetScope(ctx, &accv1.GetScopeRequest{Scope: req.Scopes[0]})
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricUserSumStandard - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "unable to fetch scope info")
	}
	if scopeinfo.ScopeType == accv1.ScopeType_SPECIFIC.String() {
		return nil, status.Error(codes.PermissionDenied, "can not user metric for specific scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricUserSumStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricUSS(ctx, serverToRepoUSS(req), req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricUSS  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric uss")
	}

	return repoToServerUSS(met), nil
}

func serverToRepoUSS(met *v1.MetricUSS) *repo.MetricUSS {
	return &repo.MetricUSS{
		Name:    met.Name,
		Default: met.Default,
	}
}

func repoToServerUSS(met *repo.MetricUSS) *v1.MetricUSS {
	return &v1.MetricUSS{
		Name:    met.Name,
		ID:      met.ID,
		Default: met.Default,
	}
}
