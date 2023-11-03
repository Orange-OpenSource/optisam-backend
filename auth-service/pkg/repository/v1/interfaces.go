package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1 Repository

// TODO fix this reflect files should be removed automatically go:generate rm -r gomock_reflect_*

// Repository interface has all the methods we need to operate on database
type Repository interface {
	// UserInfo return a Users information based on users id
	UserInfo(ctx context.Context, userID string) (*UserInfo, error)

	// IncreaseFailedLoginCount increases the count of failed login attempts
	// We should only call this function when user's password do not match with what
	// have stored in database.
	// Note: Don't call this function if user is already blocked, i.e. unsuccessful
	// attempts are already three(for now attempts limit is three this may change).
	IncreaseFailedLoginCount(ctx context.Context, userID string) error

	// ResetLoginCount reset Failed login counts in case of successful login.
	// Note: Don't call this function if user is already blocked even if he provides
	// correct credentials this time.
	ResetLoginCount(ctx context.Context, userID string) error

	// UserOwnedGroupsDirect return the groups directly owned by user
	UserOwnedGroupsDirect(ctx context.Context, userID string) ([]*Group, error)
	GetToken(ctx context.Context, acc helper.EmailParams) error
	SetToken(ctx context.Context, acc helper.EmailParams, ttl int) error
	DelToken(ctx context.Context, acc helper.EmailParams) error
	GenerateMailBody(ctx context.Context, acc helper.EmailParams, cfg config.Config) (string, error)
	AccountInfo(ctx context.Context, userID string) (*v1.AccountInfo, error)
	ChangeUserFirstLogin(ctx context.Context, userID string) error
	ChangePassword(ctx context.Context, userID, password string) error
	CreateAuthContext(cfg config.Config) (context.Context, error)
	// // CheckPassword check for users password in database
	// CheckPassword(ctx context.Context, userID, password string) (bool, error)
}
