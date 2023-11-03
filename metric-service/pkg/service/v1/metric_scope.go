package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateScopeMetric(ctx context.Context, req *v1.CreateScopeMetricRequest) (*v1.CreateScopeMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if userClaims.Role == claims.RoleUser {
		return nil, status.Error(codes.PermissionDenied, "only superadmin and Admin user can create metric")
	}
	metrics := repo.GetScopeMetric(req.Scope)
	for _, val := range metrics {
		var err error
		if val.MetricType == "microsoft.sql.enterprise" && req.Type == "microsoft.sql.enterprise" {
			if _, err = s.metricRepo.CreateMetricSQLForScope(ctx, &val); err != nil { //nolint
				logger.Log.Error("Failed to create SQLEnterprise metric", zap.String("reason", err.Error()), zap.Any("scope", req.Scope))
				return nil, status.Error(codes.Internal, "cannot SQLEnterprise metric")
			}
		}
		if val.MetricType == "windows.server.datacenter" && req.Type == "windows.server.datacenter" {
			if _, err = s.metricRepo.CreateMetricDataCenterForScope(ctx, &val); err != nil { //nolint
				logger.Log.Error("Failed to create WSDataCenter metric", zap.String("reason", err.Error()), zap.Any("scope", req.Scope))
				return nil, status.Error(codes.Internal, "cannot WSDataCenter metric")
			}
		}
	}

	return &v1.CreateScopeMetricResponse{
		Success: true,
	}, nil
}
