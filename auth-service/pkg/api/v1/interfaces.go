package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/auth-service/pkg/api/v1 AuthService

// AuthService is collection of all the methods required by AuthService
type AuthService interface {
	// Login will return LoginResponse. Error if it is not able to fetch user,
	// user does not exist or if user is blocked after three unsuccessful atemps.
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}
