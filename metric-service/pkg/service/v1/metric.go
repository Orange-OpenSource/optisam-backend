package v1

import (
	"context"
	"encoding/json"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	"optisam-backend/common/optisam/token/claims"

	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// metricServiceServer is implementation of v1.authServiceServer proto interface
type metricServiceServer struct {
	metricRepo repo.Metric
	account    accv1.AccountServiceClient
}

// NewLicenseServiceServer creates License service
func NewMetricServiceServer(metricRepo repo.Metric, grpcServers map[string]*grpc.ClientConn) v1.MetricServiceServer {
	return &metricServiceServer{
		metricRepo: metricRepo,
		account:    accv1.NewAccountServiceClient(grpcServers["account"]),
	}
}

func (s *metricServiceServer) DropMetricData(ctx context.Context, req *v1.DropMetricDataRequest) (*v1.DropMetricDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClamsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return &v1.DropMetricDataResponse{Success: false}, status.Error(codes.Internal, "ScopeValidationFailure")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DropMetricDataResponse{Success: false}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	if err := s.metricRepo.DropMetrics(ctx, req.Scope); err != nil {
		logger.Log.Error("Failed to delete metrics  for", zap.Any("scope", req.Scope), zap.Error(err))
		return &v1.DropMetricDataResponse{Success: false}, status.Error(codes.Internal, err.Error())
	}
	return &v1.DropMetricDataResponse{Success: true}, nil
}

func (s *metricServiceServer) ListMetricType(ctx context.Context, req *v1.ListMetricTypeRequest) (*v1.ListMetricTypeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	scopeinfo, err := s.account.GetScope(ctx, &accv1.GetScopeRequest{Scope: req.Scopes[0]})
	if err != nil {
		logger.Log.Error("service/v1 - ListMetricType - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "unable to fetch scope info")
	}

	// metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	// if err != nil && err != repo.ErrNoData {
	// 	logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching metrics", zap.String("reason", err.Error()))
	// 	return nil, status.Error(codes.Internal, "cannot fetch metrics")
	// }
	// var opsExists, nupExists bool
	// if metricTypeExistsAll(metrics, repo.MetricOPSOracleProcessorStandard) != -1 {
	// 	opsExists = true
	// }
	// if metricTypeExistsAll(metrics, repo.MetricOracleNUPStandard) != -1 {
	// 	nupExists = true
	// }

	metricTypes, err := s.metricRepo.ListMetricTypeInfo(ctx, repo.GetScopeType(scopeinfo.ScopeType), req.GetScopes()[0])
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

	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil {
		logger.Log.Error("service/v1 - ListMetrices - fetching metric types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types")
	}
	metricsList := s.repoMetricToServiceMetricAll(ctx, metrics, req.Scopes[0])

	return &v1.ListMetricResponse{
		Metrices: metricsList, // nolint: misspell
	}, nil

}

// nolint: gocyclo,funlen
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
		logger.Log.Error("service/v1 - GetMetricConfiguration - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
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
		if !req.GetID {
			metric, err = s.metricRepo.GetMetricConfigOPS(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigOPS", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric ops config")
			}
		} else {
			metric, err = s.metricRepo.GetMetricConfigOPSID(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigOPSID", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric ops config IDs")
			}
		}
	case repo.MetricOracleNUPStandard:
		if !req.GetID {
			metric, err = s.metricRepo.GetMetricConfigNUP(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigNUP", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric nup")
			}
		} else {
			metric, err = s.metricRepo.GetMetricConfigNUPID(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigNUPID", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric nup config IDs")
			}
		}
	case repo.MetricSPSSagProcessorStandard:
		if !req.GetID {
			metric, err = s.metricRepo.GetMetricConfigSPS(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSPS", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric sps")
			}
		} else {
			metric, err = s.metricRepo.GetMetricConfigSPSID(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigSPSID", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric sps config IDs")
			}
		}
	case repo.MetricIPSIbmPvuStandard:
		if !req.GetID {
			metric, err = s.metricRepo.GetMetricConfigIPS(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricIPS", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric ips")
			}
		} else {
			metric, err = s.metricRepo.GetMetricConfigIPSID(ctx, metrics[idx].Name, req.GetScopes()[0])
			if err != nil {
				logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigIPSID", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch metric ips config IDs")
			}
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
	case repo.MetricAttrSumStandard:
		metric, err = s.metricRepo.GetMetricConfigAttrSum(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigAttrSum", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric attr sum")
		}
	case repo.MetricUserSumStandard:
		metric, err = s.metricRepo.GetMetricConfigUSS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUSS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric uss")
		}
	case repo.MetricStaticStandard:
		metric, err = s.metricRepo.GetMetricConfigSS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ss")
		}
	case repo.MetricEquipAttrStandard:
		metric, err = s.metricRepo.GetMetricConfigEquipAttr(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricEquipAttr", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric acs")
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

// DeleteMetric deletes metric that is not being used with name and scope
func (s *metricServiceServer) DeleteMetric(ctx context.Context, req *v1.DeleteMetricRequest) (*v1.DeleteMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.Scope))
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	metric, err := s.metricRepo.MetricInfoWithAcqAndAgg(ctx, req.MetricName, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - DeleteMetric - MetricInfoWithAcqAndAgg", zap.String("reason", err.Error()))
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.Internal, "can not get metric info")
	}
	if metric.Name == "" {
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "metric does not exist")
	}
	if metric.TotalAggregations != 0 || metric.TotalAcqRights != 0 {
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "metric is being used by acquired right/aggregation")
	}

	// Check if metric exists as transform metric name
	metricTransformMetric, _ := s.metricRepo.GetMetricNUPByTransformMetricName(ctx, req.MetricName, req.Scope)
	if metricTransformMetric != nil {
		logger.Log.Error("service/v1 - DeleteMetric - GetMetricNUPByTransformMetricName", zap.String("reason", "metric is being used for transform"))
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.Internal, "metric cannot be deleted it's alloted as transform metric ")
	}

	if err := s.metricRepo.DeleteMetric(ctx, req.MetricName, req.Scope); err != nil {
		logger.Log.Error("service/v1 - DeleteMetric - DeleteMetric", zap.String("reason", err.Error()))
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.Internal, "unable to delete metric")
	}
	return &v1.DeleteMetricResponse{
		Success: true,
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

func (s *metricServiceServer) repoMetricToServiceMetricAll(ctx context.Context, met []*repo.MetricInfo, scope string) []*v1.Metric {
	servMetric := make([]*v1.Metric, len(met))
	for i := range met {
		servMetric[i] = s.repoMetricToServiceMetric(ctx, met[i], scope)
	}
	return servMetric
}

func (s *metricServiceServer) repoMetricToServiceMetric(ctx context.Context, met *repo.MetricInfo, scope string) *v1.Metric {
	desc, err := s.discriptionMetric(ctx, met, scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment", zap.String("reason", err.Error()))
	}
	return &v1.Metric{
		Name:        met.Name,
		Type:        met.Type.String(),
		Description: desc,
	}
}

func (s *metricServiceServer) discriptionMetric(ctx context.Context, met *repo.MetricInfo, scope string) (string, error) {
	switch met.Type {
	case repo.MetricOPSOracleProcessorStandard:
		return repo.MetricDescriptionOracleProcessorStandard.String(), nil
	case repo.MetricOracleNUPStandard:
		return s.getDescriptionNUP(ctx, met.Name, scope)
	case repo.MetricSPSSagProcessorStandard:
		return repo.MetricDescriptionSagProcessorStandard.String(), nil
	case repo.MetricIPSIbmPvuStandard:
		return repo.MetricDescriptionIbmPvuStandard.String(), nil
	case repo.MetricAttrCounterStandard:
		return s.getDescriptionACS(ctx, met.Name, scope)
	case repo.MetricInstanceNumberStandard:
		return s.getDescriptionINM(ctx, met.Name, scope)
	case repo.MetricAttrSumStandard:
		return s.getDescriptionAttSum(ctx, met.Name, scope)
	case repo.MetricUserSumStandard:
		return repo.MetricDescriptionUserSumStandard.String(), nil
	case repo.MetricStaticStandard:
		return s.getDescriptionSS(ctx, met.Name, scope)
	case repo.MetricEquipAttrStandard:
		return s.getDescriptionEquipAttr(ctx, met.Name, scope)
	default:
		return "", status.Error(codes.Internal, "description not found - "+met.Type.String())
	}

}

func metricNameExistsAll(metrics []*repo.MetricInfo, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

// func metricTypeExistsAll(metrics []*repo.MetricInfo, metricType repo.MetricType) int {
// 	for i, met := range metrics {
// 		if met.Type == metricType {
// 			return i
// 		}
// 	}
// 	return -1
// }
