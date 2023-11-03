package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1 AuthService

// AuthService is collection of all the methods required by AuthService
type AuthService interface {
	// Login will return LoginResponse. Error if it is not able to fetch user,
	// user does not exist or if user is blocked after three unsuccessful atemps.
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	TokenValidation(ctx context.Context, req *TokenRequest) error
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, email string) error
}
