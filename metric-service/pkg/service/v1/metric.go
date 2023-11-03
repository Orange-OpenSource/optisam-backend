package v1

import (
	"context"
	"encoding/json"

	equipv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/equipment-service/pkg/api/v1"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// metricServiceServer is implementation of v1.authServiceServer proto interface
type metricServiceServer struct {
	metricRepo repo.Metric
	account    accv1.AccountServiceClient
	equipments equipv1.EquipmentServiceClient
}

// NewLicenseServiceServer creates License service
func NewMetricServiceServer(metricRepo repo.Metric, grpcServers map[string]*grpc.ClientConn) v1.MetricServiceServer {
	return &metricServiceServer{
		metricRepo: metricRepo,
		account:    accv1.NewAccountServiceClient(grpcServers["account"]),
		equipments: equipv1.NewEquipmentServiceClient(grpcServers["equipment"]),
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

	metrics, err := s.metricRepo.ListMetrices(ctx, req.GetScopes()[0])
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricSAGProcessorStandard - fetching metrics", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	// var opsExists, nupExists bool
	// if metricTypeExistsAll(metrics, repo.MetricOPSOracleProcessorStandard) != -1 {
	// 	opsExists = true
	// }
	// if metricTypeExistsAll(metrics, repo.MetricOracleNUPStandard) != -1 {
	// 	nupExists = true
	// }

	metricTypes, err := s.metricRepo.ListMetricTypeInfo(ctx, repo.GetScopeType(scopeinfo.ScopeType), req.GetScopes()[0], req.IsImport)
	if err != nil {
		logger.Log.Error("service/v1 - ListMetricType - fetching metric types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metric types")
	}
	types := repoMetricTypeToServiceMetricTypeAll(metricTypes, req.GetScopes()[0])

	for _, val := range metrics {
		if val.Default == false {
			continue
		} else {
			for _, data := range types {
				if data.Name == string(val.Type) {
					data.IsExist = true
				}
			}
		}
	}
	return &v1.ListMetricTypeResponse{
		Types: types,
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

func (s *metricServiceServer) CreateMetric(ctx context.Context, req *v1.CreateMetricRequest) (*v1.CreateMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if req.Metric == nil || req.Metric.Name == "" || req.Metric.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "metric name and type can not be empty")
	}
	if !helper.Contains(userClaims.Socpes, req.GetSenderScope()) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	switch req.Metric.Type {
	case repo.MetricOPSOracleProcessorStandard.String():
		metric, err := s.metricRepo.GetMetricConfigOPSID(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ops config")
		}
		_, err = s.metricRepo.CreateMetricOPS(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricOracleNUPStandard.String():
		metric, err := s.metricRepo.GetMetricConfigNUPID(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigNUP", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric nup")
		}
		_, err = s.metricRepo.CreateMetricOracleNUPStandard(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricSPSSagProcessorStandard.String():
		metric, err := s.metricRepo.GetMetricConfigSPSID(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric sps")
		}
		_, err = s.metricRepo.CreateMetricSPS(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricIPSIbmPvuStandard.String():
		metric, err := s.metricRepo.GetMetricConfigIPSID(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricIPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ips")
		}
		_, err = s.metricRepo.CreateMetricIPS(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricAttrCounterStandard.String():
		metric, err := s.metricRepo.GetMetricConfigACS(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricACS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric acs")
		}
		eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetRecieverScope())
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricEquipAttrStand - fetching equipments", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch equipment types")
		}
		idx := equipmentTypeExistsByType(metric.EqType, eqTypes)
		if idx == -1 {
			return nil, status.Error(codes.NotFound, "cannot find equipment type")
		}
		attr, err := validateAttributeACSMetric(eqTypes[idx].Attributes, metric.AttributeName)
		if err != nil {
			return nil, err
		}
		_, err = s.metricRepo.CreateMetricACS(ctx, metric, attr, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric acs")
		}
	case repo.MetricInstanceNumberStandard.String():
		metric, err := s.metricRepo.GetMetricConfigINM(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricINM", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric inm")
		}
		_, err = s.metricRepo.CreateMetricInstanceNumberStandard(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric inm")
		}
	case repo.MetricAttrSumStandard.String():
		metric, err := s.metricRepo.GetMetricConfigAttrSum(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigAttrSum", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric attr sum")
		}
		eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetRecieverScope())
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricEquipAttrStand - fetching equipments", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch equipment types")
		}
		idx := equipmentTypeExistsByType(metric.EqType, eqTypes)
		if idx == -1 {
			return nil, status.Error(codes.NotFound, "cannot find equipment type")
		}
		attr, err := validateAttributeASSMetric(eqTypes[idx].Attributes, metric.AttributeName)
		if err != nil {
			return nil, err
		}
		_, err = s.metricRepo.CreateMetricAttrSum(ctx, metric, attr, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricUserSumStandard.String():
		metric, err := s.metricRepo.GetMetricConfigUSS(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUSS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric uss")
		}
		_, err = s.metricRepo.CreateMetricUSS(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric uss")
		}
	case repo.MetricStaticStandard.String():
		metric, err := s.metricRepo.GetMetricConfigSS(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricSS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric ss")
		}
		_, err = s.metricRepo.CreateMetricStaticStandard(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricEquipAttrStandard.String():
		metric, err := s.metricRepo.GetMetricConfigEquipAttr(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricEquipAttr", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric acs")
		}
		eqTypes, err := s.metricRepo.EquipmentTypes(ctx, req.GetRecieverScope())
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricEquipAttrStand - fetching equipments", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch equipment types")
		}
		idx := equipmentTypeExistsByType(metric.EqType, eqTypes)
		if idx == -1 {
			return nil, status.Error(codes.NotFound, "cannot find equipment type")
		}
		attr, err := validateEquipAttStandardMetric(eqTypes[idx].Attributes, metric.AttributeName)
		if err != nil {
			return nil, err
		}
		_, err = s.metricRepo.CreateMetricEquipAttrStandard(ctx, metric, attr, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric ss")
		}
	case repo.MetricUserNomStandard.String():
		metric, err := s.metricRepo.GetMetricConfigUNS(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUNS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric UNS")
		}
		_, err = s.metricRepo.CreateMetricUserNominativeStandard(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricStaticStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric UNS")
		}
	case repo.MetricUserConcurentStandard.String():
		metric, err := s.metricRepo.GetMetricConfigConcurentUser(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUCS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric UCS")
		}
		_, err = s.metricRepo.CreateMetricUserConcurentStandard(ctx, metric, req.RecieverScope)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricUserConcurentStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric UCS")
		}
	case repo.MetricMicrosoftSQLStandard.String():
		metric, err := s.metricRepo.GetMetricConfigSQLStandard(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigSQLStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric sql_standard")
		}
		metric.Scope = req.RecieverScope
		_, err = s.metricRepo.CreateMetricSQLStandard(ctx, metric)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricSQLStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric sql_standard")
		}
	case repo.MetricMicrosoftSQLEnterprise.String():
		metric, err := s.metricRepo.GetMetricConfigSQLForScope(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigSQLForScope", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric sql")
		}
		metric.Scope = req.RecieverScope
		_, err = s.metricRepo.CreateMetricSQLForScope(ctx, metric)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricSQLForScope  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric sql")
		}
	case repo.MetricWindowsServerDataCenter.String():
		metric, err := s.metricRepo.GetMetricConfigDataCenterForScope(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigDataCenterForScope", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric win_dcenter")
		}
		metric.Scope = req.RecieverScope
		_, err = s.metricRepo.CreateMetricDataCenterForScope(ctx, metric)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricDataCenterForScope  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric win_dcenter")
		}
	case repo.MetricWindowsServerStandard.String():
		metric, err := s.metricRepo.GetMetricConfigWindowServerStandard(ctx, req.Metric.Name, req.SenderScope)
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricConfigWindowServerStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric win_server_stand")
		}
		metric.Scope = req.RecieverScope
		_, err = s.metricRepo.CreateMetricWindowServerStandard(ctx, metric)
		if err != nil {
			logger.Log.Error("service/v1 - CreateMetricWindowServerStandard  in repo", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot create metric win_server_stand")
		}
	}
	return &v1.CreateMetricResponse{Success: true}, nil
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
	case repo.MetricUserNomStandard:
		metric, err = s.metricRepo.GetMetricConfigUNS(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUNS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric UNS")
		}
	case repo.MetricUserConcurentStandard:
		metric, err = s.metricRepo.GetMetricConfigConcurentUser(ctx, metrics[idx].Name, req.GetScopes()[0])
		if err != nil {
			logger.Log.Error("service/v1 - GetMetricConfiguration - GetMetricUCS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch metric UNS")
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
	if metric.Default == true {
		return &v1.DeleteMetricResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "metric imported, can't be deleted")
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

func repoMetricTypeToServiceMetricTypeAll(met []*repo.MetricTypeInfo, scope string) []*v1.MetricType {
	servMetrics := make([]*v1.MetricType, len(met))
	myMap := make(map[string][]string)
	metadata := repo.GlobalMetricMetadata(scope)
	dataScope := repo.GetScopeMetric(scope)
	for i, val := range metadata {
		switch i {
		case "oracle.processor.standard":
			myMap[i] = append(myMap[i], val.MetadataOPS.Name)
		case "oracle.nup.standard":
			myMap[i] = append(myMap[i], val.MetadataNUP.Name)
		case "instance.number.standard":
			myMap[i] = append(myMap[i], val.MetadataINM.Name)
		case "user.sum.standard":
			myMap[i] = append(myMap[i], val.MetadataUSS.Name)
		case "sag.processor.standard":
			myMap[i] = append(myMap[i], val.MetadataSPS.Name)
		case "ibm.pvu.standard":
			myMap[i] = append(myMap[i], val.MetadataSPS.Name)
		case "microsoft.sql.standard":
			myMap[i] = append(myMap[i], val.MetadataSQL.MetricName)
		case "microsoft.sql.enterprise":
			myMap[i] = append(myMap[i], val.MetadataSQL.MetricName)
		case "windows.server.datacenter":
			for _, v := range dataScope {
				if v.MetricType == "windows.server.datacenter" {
					myMap[i] = append(myMap[i], v.MetricName)
				}
			}
		case "user.nominative.standard":
			myMap[i] = append(myMap[i], val.MetadataUNS.Name)
		case "user.concurrent.standard":
			myMap[i] = append(myMap[i], val.MetadataUNS.Name)
		case "equipment.attribute.standard":
			for _, data := range val.MetadataEquipAttr {
				myMap[i] = append(myMap[i], data.Name)
			}
		case "windows.server.standard":
			myMap[i] = append(myMap[i], val.MetadataSQL.MetricName)
		case "static.standard":
			myMap[i] = append(myMap[i], val.MetadataSS.Name)
		case "attribute.sum.standard":
			myMap[i] = append(myMap[i], val.MetadataAttrSum.Name)
		case "attribute.counter.standard":
			myMap[i] = append(myMap[i], val.MetadataACS.Name)
		}
	}
	for i := range met {
		servMetrics[i] = repoMetricTypeToServiceMetricType(met[i], myMap)
	}
	return servMetrics
}

func repoMetricTypeToServiceMetricType(met *repo.MetricTypeInfo, myMap map[string][]string) *v1.MetricType {
	return &v1.MetricType{
		Name:           string(met.Name),
		Description:    met.Description,
		Href:           met.Href,
		TypeId:         v1.MetricType_Type(met.MetricType),
		IsExist:        met.Exist,
		DefaultMetrics: myMap[string(met.Name)],
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
		Default:     met.Default,
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
	case repo.MetricUserNomStandard:
		return s.getDescriptionUNS(ctx, met.Name, scope)
	case repo.MetricUserConcurentStandard:
		return s.getDescriptionUCS(ctx, met.Name, scope)
	case repo.MetricMicrosoftSQLEnterprise:
		return s.getDescriptionSQLEnterprise(ctx, met.Name, scope)
	case repo.MetricWindowsServerDataCenter:
		return s.getDescriptionWinDCenter(ctx, met.Name, scope)
	case repo.MetricMicrosoftSQLStandard:
		return s.getDescriptionSQLStandard(ctx, met.Name, scope)
	case repo.MetricWindowsServerStandard:
		return s.getDescriptionWinStandard(ctx, met.Name, scope)
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

func (s *metricServiceServer) getDescriptionSQLStandard(ctx context.Context, name, scope string) (string, error) {
	des := repo.MetricDescriptionMicrosoftSQLStandard.String()
	return des, nil
}

func (s *metricServiceServer) getDescriptionSQLEnterprise(ctx context.Context, name, scope string) (string, error) {
	des := repo.MetricDescriptionMicrosoftSQLEnterprise.String()
	return des, nil
}

func (s *metricServiceServer) getDescriptionWinDCenter(ctx context.Context, name, scope string) (string, error) {
	des := repo.MetricDescriptionWindowsServerDataCenter.String()
	return des, nil
}

func (s *metricServiceServer) getDescriptionWinStandard(ctx context.Context, name, scope string) (string, error) {
	des := repo.MetricDescriptionWindowsServerStandard.String()
	return des, nil
}
