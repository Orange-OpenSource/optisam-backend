package v1

import (
	"context"
	"database/sql"
	"errors"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	equipment "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/thirdparty/equipment-service/pkg/api/v1"

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

	scope, err := s.accountRepo.ListScopes(ctx, []string{req.ScopeCode})
	if err != nil {
		if err != repo.ErrNoData {
			logger.Log.Error("service/v1 - CreateScope - Repo: ListScopes", zap.Error(err))
		}
	}
	//set scope in redis
	err = s.accountRepo.SetScope(ctx, scope)
	if err != nil {
		logger.Log.Error("service/v1 - CreateScope - Repo: SetScope", zap.Error(err))
	}
	return &v1.CreateScopeResponse{Success: true}, nil

}

func (s *accountServiceServer) ListScopes(ctx context.Context, req *v1.ListScopesRequest) (*v1.ListScopesResponse, error) {
	logger.Log.Info("List Scopes", zap.Any("list scopes called", time.Now()))
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
	var scopes []*repo.Scope
	var err error
	//redis call
	scopes, err = s.accountRepo.GetScopes(ctx, scopeCodes)
	if err != nil {
		logger.Log.Info("service/v1 - ListScopes - Repo: ListScopes : unable to find scope details in redis", zap.Error(err))
	}
	// Call ListScopes
	if len(scopes) != len(scopeCodes) {
		scopes, err = s.accountRepo.ListScopes(ctx, scopeCodes)
		if err != nil {
			logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes", zap.Error(err))
			return nil, status.Error(codes.Internal, "Unable to fetch scopes")
		}
		err = s.accountRepo.SetScope(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes : unable to set scope details in redis", zap.Error(err))
		}
	}
	if len(scopes) == 0 {
		return &v1.ListScopesResponse{}, nil
	}

	scopeList, err := repoScopeListToServrepoList(scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListScopes - ListScopes  - timestampProto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}
	logger.Log.Info("List Scopes", zap.Any("end", time.Now()))

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
	logger.Log.Info("repoScopeListToServrepoList", zap.Any("before parsing", time.Now()))

	res := make([]*v1.Scope, 0)

	for _, scope := range scopes {
		protoTime, err := pTypes.TimestampProto(scope.CreatedOn)
		if err != nil {
			return nil, err
		}
		servScope := repoScopeToListScope(scope, protoTime)
		res = append(res, servScope)
	}
	logger.Log.Info("repoScopeListToServrepoList", zap.Any("after parsing", time.Now()))
	return res, nil
}

func repoScopeToListScope(scope *repo.Scope, time *tspb.Timestamp) *v1.Scope {
	return &v1.Scope{
		ScopeCode:   scope.ScopeCode,
		ScopeName:   scope.ScopeName,
		CreatedBy:   scope.CreatedBy,
		CreatedOn:   time,
		GroupNames:  scope.GroupNames,
		ScopeType:   scope.ScopeType,
		Expenditure: scope.Expenses.Float64,
	}
}

func (s *accountServiceServer) UpsertScopeExpenses(ctx context.Context, req *v1.UpsertScopeExpensesRequest) (*v1.CreateScopeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !(userClaims.Role == claims.RoleSuperAdmin || userClaims.Role == claims.RoleAdmin) {
		return nil, status.Error(codes.PermissionDenied, "only superadmin or admin user can create/update scope expenses")
	}

	scope, err := s.accountRepo.ScopeByCode(ctx, req.ScopeCode)
	if err != nil {
		logger.Log.Error("service/v1 - UpsertScopeExpenses", zap.Error(err))
		return nil, status.Error(codes.Internal, "Can not fetch scopes")
	}

	err = s.accountRepo.UpsertScopeExpenses(ctx, req.GetScopeCode(), userClaims.UserID, userClaims.UserID, req.GetExpenses(), time.Now().Year()-1)
	if err != nil {
		return nil, status.Error(codes.Internal, "Can not update scope expenses")
	}
	scope.Expenses = sql.NullFloat64{Float64: float64(req.Expenses)}
	err = s.accountRepo.SetScope(ctx, []*repo.Scope{scope})
	if err != nil {
		logger.Log.Error("service/v1 - UpsertScopeExpenses ", zap.String("reason", "Update scope expenses in redis"))
		return nil, status.Error(codes.Internal, "Can not update scope expenses")
	}
	return &v1.CreateScopeResponse{Success: true}, nil

}

func (s *accountServiceServer) GetScopeExpenses(ctx context.Context, req *v1.GetScopeRequest) (*v1.ScopeExpenses, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - GetScopeExpenses ", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	scopeExpenses, err := s.accountRepo.ScopeExpensesByScopeCode(ctx, req.Scope)
	if err != nil {
		if errors.Is(err, repo.ErrNoData) {
			logger.Log.Error("service/v1 - GetScopeExpenses - repo/ScopeExpensesByScopeCode - ", zap.Error(err))
			return &v1.ScopeExpenses{}, status.Error(codes.NotFound, "no expenses exists for previous year")
		}
		logger.Log.Error("service/v1 - GetScopeExpenses - repo/ScopeExpensesByScopeCode - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "unable to get scope info")
	}
	return &v1.ScopeExpenses{ScopeCode: req.GetScope(), Expenses: float64(scopeExpenses)}, nil
}
func (s *accountServiceServer) GetScopeLists(ctx context.Context, req *v1.GetScopeListRequest) (*v1.ScopeListResponse, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	// If there are no scopes available to user.
	if len(req.Scopes) == 0 {
		return &v1.ScopeListResponse{}, nil
	}

	// Fetch Scopes from request body
	scopeCodes := req.Scopes
	//scopeCodes := userClaims.Socpes
	var scopes []*repo.Scope
	var err error
	//redis call
	scopes, err = s.accountRepo.GetScopes(ctx, scopeCodes)
	if err != nil {
		logger.Log.Info("service/v1 - ListScopes - Repo: ListScopes : unable to find scope details in redis", zap.Error(err))
	}
	// Call ListScopes
	if len(scopes) != len(scopeCodes) {
		scopes, err = s.accountRepo.ListScopes(ctx, scopeCodes)
		if err != nil {
			logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes", zap.Error(err))
			return nil, status.Error(codes.Internal, "Unable to fetch scopes")
		}
		err = s.accountRepo.SetScope(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes : unable to set scope details in redis", zap.Error(err))
		}
	}
	// Call ListScopes
	//scopes, err = s.accountRepo.ListScopes(ctx, scopeCodes)

	if err != nil {
		logger.Log.Error("service/v1 - ListScopes - Repo: ListScopes", zap.Error(err))
		return nil, status.Error(codes.Internal, "Unable to fetch scopes")
	}

	if len(scopes) == 0 {
		return &v1.ScopeListResponse{}, nil
	}
	var scopeList []string

	for _, scpData := range scopes {
		scopeList = append(scopeList, scpData.ScopeName)
	}
	return &v1.ScopeListResponse{
		ScopeNames: scopeList,
	}, nil

}
