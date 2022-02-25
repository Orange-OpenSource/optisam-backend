package v1

import (
	"context"
	"errors"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"

	equipment "optisam-backend/equipment-service/pkg/api/v1"

	pTypes "github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *accountServiceServer) CreateScope(ctx context.Context, req *v1.CreateScopeRequest) (*v1.CreateScopeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "only superadmin user can create scope")
	}

	_, err := s.accountRepo.ScopeByCode(ctx, req.ScopeCode)

	if err != nil {
		if err != repo.ErrNoData {
			logger.Log.Error("service/v1 - CreateScope - Repo: CreateScope", zap.Error(err))
			return nil, status.Error(codes.Internal, "Can not fetch scopes")
		}
	} else {
		return nil, status.Error(codes.AlreadyExists, "Scope already exists")
	}

	if err = s.accountRepo.CreateScope(ctx, req.ScopeName, req.ScopeCode, userClaims.UserID, req.ScopeType.String()); err != nil {
		logger.Log.Error("service/v1 - CreateScope - Repo: CreateScope", zap.Error(err))
		return nil, status.Error(codes.Internal, "Unable to create new scope")
	}

	if req.ScopeType == v1.ScopeType_GENERIC {
		if _, err := s.equipment.CreateGenericScopeEquipmentTypes(ctx, &equipment.CreateGenericScopeEquipmentTypesRequest{Scope: req.ScopeCode}); err != nil {
			logger.Log.Error("service/v1 - Create Generic Scope Metadata & eqType ", zap.Error(err))
			return nil, status.Error(codes.Internal, "Unable to create Metadata & EqTypes")
		}
	}
	return &v1.CreateScopeResponse{Success: true}, nil

}

func (s *accountServiceServer) ListScopes(ctx context.Context, req *v1.ListScopesRequest) (*v1.ListScopesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	// If there are no scopes available to user.
	if len(userClaims.Socpes) == 0 {
		return &v1.ListScopesResponse{}, nil
	}

	// Fetch Scopes from user claims
	scopeCodes := userClaims.Socpes

	// Call ListScopes
	scopes, err := s.accountRepo.ListScopes(ctx, scopeCodes)

	if err != nil {
		logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes", zap.Error(err))
		return nil, status.Error(codes.Internal, "Unable to fetch scopes")
	}

	if len(scopes) == 0 {
		return &v1.ListScopesResponse{}, nil
	}

	scopeList, err := repoScopeListToServrepoList(scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListScopes - ListScopes  - timestampProto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &v1.ListScopesResponse{
		Scopes: scopeList,
	}, nil

}

func (s *accountServiceServer) GetScope(ctx context.Context, req *v1.GetScopeRequest) (*v1.Scope, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - GetScope ", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	scopeInfo, err := s.accountRepo.ScopeByCode(ctx, req.Scope)
	if err != nil {
		if errors.Is(err, repo.ErrNoData) {
			logger.Log.Error("service/v1 - GetScope - repo/ScopeByCode - ", zap.Error(err))
			return &v1.Scope{}, status.Error(codes.NotFound, "scope does not exist")
		}
		logger.Log.Error("service/v1 - GetScope - repo/ScopeByCode - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable to get scope info")
	}
	protoTime, err := pTypes.TimestampProto(scopeInfo.CreatedOn)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return repoScopeToListScope(scopeInfo, protoTime), nil
}

func repoScopeListToServrepoList(scopes []*repo.Scope) ([]*v1.Scope, error) {

	res := make([]*v1.Scope, 0)

	for _, scope := range scopes {
		protoTime, err := pTypes.TimestampProto(scope.CreatedOn)
		if err != nil {
			return nil, err
		}
		servScope := repoScopeToListScope(scope, protoTime)
		res = append(res, servScope)
	}

	return res, nil
}

func repoScopeToListScope(scope *repo.Scope, time *tspb.Timestamp) *v1.Scope {
	return &v1.Scope{
		ScopeCode:  scope.ScopeCode,
		ScopeName:  scope.ScopeName,
		CreatedBy:  scope.CreatedBy,
		CreatedOn:  time,
		GroupNames: scope.GroupNames,
		ScopeType:  scope.ScopeType,
	}
}
