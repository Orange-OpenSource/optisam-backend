package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricSQLStandard(ctx context.Context, req *v1.MetricScopeSQL) (*v1.MetricScopeSQL, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Sugar().Errorw("service/v1 - CreateMetricSQLStandard - ListMetrics error",
			"scope", req.Scopes,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if metricNameExistsAll(metrics, req.MetricName) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricSQLStandard(ctx, serverToRepoSQLStand(req))
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - CreateMetricSQLStandard in repo - Repository function error",
			"scope", req.Scopes,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, status.Error(codes.Internal, "cannot create metric sql_standard")
	}

	return repoToServerSQLStand(met), nil
}

func serverToRepoSQLStand(met *v1.MetricScopeSQL) *repo.MetricSQLStand {
	// des := repo.MetricDescriptionInstanceNumberStandard
	// v := strings.Replace(des, "number_of_deployments_authorized_licenses", met.NumOfDeployments)

	return &repo.MetricSQLStand{
		MetricType: met.MetricType,
		MetricName: met.MetricName,
		Reference:  met.Reference,
		Core:       met.Core,
		CPU:        met.CPU,
		Default:    met.Default,
		Scope:      met.Scopes[0],
	}
}

func repoToServerSQLStand(met *repo.MetricSQLStand) *v1.MetricScopeSQL {
	return &v1.MetricScopeSQL{
		ID:         met.ID,
		MetricType: met.MetricType,
		MetricName: met.MetricName,
		Reference:  met.Reference,
		Core:       met.Core,
		CPU:        met.CPU,
		Default:    met.Default,
		Scopes:     []string{met.Scope},
	}
}
