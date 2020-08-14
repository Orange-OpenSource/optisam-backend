// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	pTypes "github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *accountServiceServer) CreateScope(ctx context.Context, req *v1.CreateScopeRequest) (*v1.CreateScopeResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
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

	err = s.accountRepo.CreateScope(ctx, req.ScopeName, req.ScopeCode, userClaims.UserID)

	if err != nil {
		logger.Log.Error("service/v1 - CreateScope - Repo: CreateScope", zap.Error(err))
		return nil, status.Error(codes.Internal, "Unable to create new scope")
	}

	return &v1.CreateScopeResponse{}, nil

}

func (s *accountServiceServer) ListScopes(ctx context.Context, req *v1.ListScopesRequest) (*v1.ListScopesResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	// If there are no scopes available to user.
	if len(userClaims.Socpes) == 0 {
		return &v1.ListScopesResponse{}, nil
	}

	//Fetch Scopes from user claims
	scopeCodes := userClaims.Socpes

	//Call ListScopes
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

func repoScopeListToServrepoList(scopes []*repo.Scope) ([]*v1.Scope, error) {

	var res []*v1.Scope

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
	}
}
