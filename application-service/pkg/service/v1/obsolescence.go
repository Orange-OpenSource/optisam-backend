package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/repository/v1/postgres/db"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *applicationServiceServer) ApplicationDomains(ctx context.Context, req *v1.ApplicationDomainsRequest) (*v1.ApplicationDomainsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ApplicationDomains", zap.String("reason", "ScopeError"))
		return &v1.ApplicationDomainsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := s.applicationRepo.GetApplicationDomains(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ApplicationDomains - GetApplicationDomains", zap.String("reason", err.Error()))
		return &v1.ApplicationDomainsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ApplicationDomainsResponse{Domains: dbresp}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenceDomainCriticityMeta(ctx context.Context, req *v1.DomainCriticityMetaRequest) (*v1.DomainCriticityMetaResponse, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenceDomainCriticityMeta", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.applicationRepo.GetDomainCriticityMeta(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenceDomainCriticityMeta - GetDomainCriticityMeta", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.DomainCriticityMetaResponse{}
	apiresp.DomainCriticityMeta = make([]*v1.DomainCriticityMeta, len(dbresp))
	for i := range dbresp {
		apiresp.DomainCriticityMeta[i] = &v1.DomainCriticityMeta{}
		apiresp.DomainCriticityMeta[i].DomainCriticId = dbresp[i].DomainCriticID
		apiresp.DomainCriticityMeta[i].DomainCriticName = dbresp[i].DomainCriticName
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenceMaintenanceCriticityMeta(ctx context.Context, req *v1.MaintenanceCriticityMetaRequest) (*v1.MaintenanceCriticityMetaResponse, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenceMaintenanceCriticityMeta", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.applicationRepo.GetMaintenanceCricityMeta(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenceMaintenanceCriticityMeta - GetMaintenanceCricityMeta", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.MaintenanceCriticityMetaResponse{}
	apiresp.MaintenanceCriticityMeta = make([]*v1.MaintenanceCriticityMeta, len(dbresp))
	for i := range dbresp {
		apiresp.MaintenanceCriticityMeta[i] = &v1.MaintenanceCriticityMeta{}
		apiresp.MaintenanceCriticityMeta[i].MaintenanceCriticId = dbresp[i].MaintenanceLevelID
		apiresp.MaintenanceCriticityMeta[i].MaintenanceCriticName = dbresp[i].MaintenanceLevelName
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenceRiskMeta(ctx context.Context, req *v1.RiskMetaRequest) (*v1.RiskMetaResponse, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenceRiskMeta", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.applicationRepo.GetRiskMeta(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenceRiskMeta - GetRiskMeta", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.RiskMetaResponse{}
	apiresp.RiskMeta = make([]*v1.RiskMeta, len(dbresp))
	for i := range dbresp {
		apiresp.RiskMeta[i] = &v1.RiskMeta{}
		apiresp.RiskMeta[i].RiskId = dbresp[i].RiskID
		apiresp.RiskMeta[i].RiskName = dbresp[i].RiskName
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenceDomainCriticity(ctx context.Context, req *v1.DomainCriticityRequest) (*v1.DomainCriticityResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenceDomainCriticity", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ObsolescenceDomainCriticity", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := s.applicationRepo.GetDomainCriticity(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenceDomainCriticity - GetDomainCriticity", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.DomainCriticityResponse{}
	apiresp.DomainsCriticity = make([]*v1.DomainCriticity, len(dbresp))
	for i := range dbresp {
		apiresp.DomainsCriticity[i] = &v1.DomainCriticity{}
		apiresp.DomainsCriticity[i].DomainCriticId = dbresp[i].DomainCriticID
		apiresp.DomainsCriticity[i].Domains = dbresp[i].Domains
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenseMaintenanceCriticity(ctx context.Context, req *v1.MaintenanceCriticityRequest) (*v1.MaintenanceCriticityResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenseMaintenanceCriticity", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ObsolescenseMaintenanceCriticity", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := s.applicationRepo.GetMaintenanceTimeCriticity(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenseMaintenanceCriticity - GetMaintenanceTimeCriticity", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.MaintenanceCriticityResponse{}
	apiresp.MaintenanceCriticy = make([]*v1.MaintenanceCriticity, len(dbresp))
	for i := range dbresp {
		apiresp.MaintenanceCriticy[i] = &v1.MaintenanceCriticity{}
		apiresp.MaintenanceCriticy[i].MaintenanceCriticId = dbresp[i].MaintenanceCriticID
		apiresp.MaintenanceCriticy[i].MaintenanceLevelId = dbresp[i].LevelID
		apiresp.MaintenanceCriticy[i].StartMonth = dbresp[i].StartMonth
		apiresp.MaintenanceCriticy[i].EndMonth = dbresp[i].EndMonth
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ObsolescenseRiskMatrix(ctx context.Context, req *v1.RiskMatrixRequest) (*v1.RiskMatrixResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - ObsolescenseRiskMatrix", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ObsolescenseRiskMatrix", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := s.applicationRepo.GetRiskMatrixConfig(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ObsolescenseRiskMatrix - GetRiskMatrixConfig", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.RiskMatrixResponse{}
	apiresp.RiskMatrix = make([]*v1.RiskMatrix, len(dbresp))

	for i := range dbresp {
		apiresp.RiskMatrix[i] = &v1.RiskMatrix{}
		apiresp.RiskMatrix[i].ConfigurationId = dbresp[i].ConfigurationID
		apiresp.RiskMatrix[i].DomainCriticId = dbresp[i].DomainCriticID
		apiresp.RiskMatrix[i].DomainCriticName = dbresp[i].DomainCriticName
		apiresp.RiskMatrix[i].MaintenanceCriticId = dbresp[i].MaintenanceLevelID
		apiresp.RiskMatrix[i].MaintenanceCriticName = dbresp[i].MaintenanceLevelName
		apiresp.RiskMatrix[i].RiskId = dbresp[i].RiskID
		apiresp.RiskMatrix[i].RiskName = dbresp[i].RiskName
	}
	return apiresp, nil
}

func (s *applicationServiceServer) PostObsolescenceDomainCriticity(ctx context.Context, req *v1.PostDomainCriticityRequest) (*v1.PostDomainCriticityResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - PostObsolescenceDomainCriticity", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - PostObsolescenceDomainCriticity", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbrespdomains, err := s.applicationRepo.GetDomainCriticityMetaIDs(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenceDomainCriticity - GetDomainCriticityMetaIDs", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	for _, domain := range req.DomainsCriticity {
		if !helper.ContainsInts(dbrespdomains, domain.DomainCriticId) {
			logger.Log.Error("service/v1 - PostObsolescenceDomainCriticity", zap.String("reason", "DomainValidationError"))
			return nil, status.Error(codes.Unknown, "DomainValidationError")
		}
	}
	for _, domain := range req.DomainsCriticity {
		err := s.applicationRepo.InsertDomainCriticity(ctx, db.InsertDomainCriticityParams{
			Scope:          req.GetScope(),
			CreatedBy:      userClaims.UserID,
			DomainCriticID: domain.DomainCriticId,
			Domains:        domain.Domains,
		})
		if err != nil {
			logger.Log.Error("service/v1 - PostObsolescenceDomainCriticity - InsertDomainCriticity", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Unknown, "DBError")
		}
	}
	return &v1.PostDomainCriticityResponse{Success: true}, nil
}

func (s *applicationServiceServer) PostObsolescenseMaintenanceCriticity(ctx context.Context, req *v1.PostMaintenanceCriticityRequest) (*v1.PostMaintenanceCriticityResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - PostObsolescenseMaintenanceCriticity", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - PostObsolescenseMaintenanceCriticity", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}

	dbrespmaintenance, err := s.applicationRepo.GetMaintenanceCricityMetaIDs(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenseMaintenanceCriticity - GetMaintenanceCricityMetaIDs", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	for _, maintenance := range req.MaintenanceCriticy {
		if !helper.ContainsInts(dbrespmaintenance, maintenance.MaintenanceLevelId) {
			logger.Log.Error("service/v1 - PostObsolescenseMaintenanceCriticity", zap.String("reason", "MaintenanceValidationError"))
			return nil, status.Error(codes.Unknown, "MaintenanceValidationError")
		}
	}

	for _, maintenance := range req.MaintenanceCriticy {
		err := s.applicationRepo.InsertMaintenanceTimeCriticity(ctx, db.InsertMaintenanceTimeCriticityParams{
			Scope:      req.GetScope(),
			CreatedBy:  userClaims.UserID,
			LevelID:    maintenance.MaintenanceLevelId,
			StartMonth: maintenance.StartMonth,
			EndMonth:   maintenance.EndMonth,
		})
		if err != nil {
			logger.Log.Error("service/v1 - PostObsolescenseMaintenanceCriticity - InsertMaintenanceTimeCriticity", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Unknown, "DBError")
		}
	}
	return &v1.PostMaintenanceCriticityResponse{Success: true}, nil
}

func (s *applicationServiceServer) PostObsolescenseRiskMatrix(ctx context.Context, req *v1.PostRiskMatrixRequest) (*v1.PostRiskMatrixResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix", zap.String("reason", "userclaims not found"))
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}

	dbrespdomains, err := s.applicationRepo.GetDomainCriticityMetaIDs(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix - GetDomainCriticityMetaIDs", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	dbrespmaintenance, err := s.applicationRepo.GetMaintenanceCricityMetaIDs(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix - GetMaintenanceCricityMetaIDs", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	dbresprisklevels, err := s.applicationRepo.GetRiskLevelMetaIDs(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix - GetRiskLevelMetaIDs", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	for _, riskconfig := range req.RiskMatrix {
		if !helper.ContainsInts(dbrespdomains, riskconfig.DomainCriticId) || !helper.ContainsInts(dbrespmaintenance, riskconfig.MaintenanceCriticId) || !helper.ContainsInts(dbresprisklevels, riskconfig.RiskId) {
			logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix", zap.String("reason", "RickConfigValidationError"))
			return nil, status.Error(codes.Unknown, "RickConfigValidationError")
		}
	}

	configID, err := s.applicationRepo.InsertRiskMatrix(ctx, db.InsertRiskMatrixParams{
		Scope:     req.GetScope(),
		CreatedBy: userClaims.UserID,
	})
	if err != nil {
		logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix - InsertRiskMatrix", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	for _, riskconfig := range req.RiskMatrix {
		err := s.applicationRepo.InsertRiskMatrixConfig(ctx, db.InsertRiskMatrixConfigParams{
			ConfigurationID:    configID,
			DomainCriticID:     riskconfig.DomainCriticId,
			MaintenanceLevelID: riskconfig.MaintenanceCriticId,
			RiskID:             riskconfig.RiskId,
		})
		if err != nil {
			logger.Log.Error("service/v1 - PostObsolescenseRiskMatrix - InsertRiskMatrixConfig", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Unknown, "DBError")
		}
	}
	return &v1.PostRiskMatrixResponse{Success: true}, nil
}
