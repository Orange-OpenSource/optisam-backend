// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/auth-service/pkg/api/v1"
	"optisam-backend/auth-service/pkg/oauth2/errors"
	repoV1 "optisam-backend/auth-service/pkg/repository/v1"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	"go.uber.org/zap"
)

// AuthServiceServer is implementation of v1.AuthServiceServer proto interface
type AuthServiceServer struct {
	rep repoV1.Repository
}

// NewAuthServiceServer creates Auth service
func NewAuthServiceServer(rep repoV1.Repository) *AuthServiceServer {
	return &AuthServiceServer{rep: rep}
}

// Login implements AuthService Login function
func (s *AuthServiceServer) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	ui, err := s.rep.UserInfo(ctx, req.Username)
	if err != nil {
		// check if user exists or not
		if err == sql.ErrNoRows {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, err
	}

	// check if user is blocked
	if ui.FailedLogins >= 3 {
		return nil, errors.ErrLoginBlockedAccount
	}

	// Check if password is correct or not
	correct, err := s.rep.CheckPassword(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	if !correct {
		// Now increase failed login counts
		if err := s.rep.IncreaseFailedLoginCount(ctx, req.Username); err != nil {
			return nil, fmt.Errorf("service/v1 login failed to increase unsuccessful login count: %v", err)
		}
		// check if user is blocked
		if ui.FailedLogins == 2 {
			return nil, errors.ErrAccountBlocked
		}
		return nil, errors.ErrInvalidCredentials
	}

	// User has validated his credentials now rest failed login attempts to zero
	if err := s.rep.ResetLoginCount(ctx, req.Username); err != nil {
		return nil, fmt.Errorf("service/v1 login failed to reset unsuccessful login count: %v", err)
	}

	return &v1.LoginResponse{
		UserID: ui.UserID,
	}, nil
}

// UserClaims implements access.ClaimsFetcher UserClaims Function
func (s *AuthServiceServer) UserClaims(ctx context.Context, userID string) (*claims.Claims, error) {
	info, err := s.rep.UserInfo(ctx, userID)
	if err != nil {
		logger.Log.Error("service/v1 - UserClaims cannot fetch user info", zap.Error(err))
		return nil, fmt.Errorf("cannot get claims for user: %v", userID)
	}

	role, err := translateRole(info.Role)
	if err != nil {
		logger.Log.Error("service/v1 - UserClaims cannot tranlate user role", zap.Error(err))
		return nil, fmt.Errorf("cannot get claims for user: %v", userID)
	}

	grps, err := s.rep.UserOwnedGroupsDirect(ctx, userID)
	if err != nil {
		logger.Log.Error("service/v1 - UserClaims cannot fetch user info", zap.Error(err))
		return nil, fmt.Errorf("cannot get claims for user: %v", userID)
	}
	var scopes []string
	for _, grp := range grps {
		for _, s := range grp.Scopes {
			if !elementExists(scopes, s) {
				scopes = append(scopes, s)
			}
		}
	}
	return &claims.Claims{
		UserID: userID,
		Role:   role,
		Locale: info.Locale,
		Socpes: scopes,
	}, nil
}

func elementExists(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func translateRole(role repoV1.Role) (claims.Role, error) {
	switch role {
	case repoV1.RoleSuperAdmin:
		return claims.RoleSuperAdmin, nil
	case repoV1.RoleAdmin:
		return claims.RoleAdmin, nil
	case repoV1.RoleUser:
		return claims.RoleUser, nil
	default:
		return "", fmt.Errorf("service - v1 - translateRole unknow role from databnase: %v", role)
	}
}
