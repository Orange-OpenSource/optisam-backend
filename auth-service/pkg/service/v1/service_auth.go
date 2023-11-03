package v1

import (
	"context"
	"database/sql"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/errors"
	repoV1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

// AuthServiceServer is implementation of v1.AuthServiceServer proto interface
type AuthServiceServer struct {
	rep           repoV1.Repository
	cfg           config.Config
	notification  v1.NotificationServiceClient
	kafkaProducer *kafka.Producer
}

// NewAuthServiceServer creates Auth service
func NewAuthServiceServer(rep repoV1.Repository, cfg config.Config, grpcConnections map[string]*grpc.ClientConn, kafkaProducer *kafka.Producer) *AuthServiceServer {
	return &AuthServiceServer{
		rep:           rep,
		cfg:           cfg,
		notification:  v1.NewNotificationServiceClient(grpcConnections["notification"]),
		kafkaProducer: kafkaProducer,
	}
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
	// if ui.FailedLogins >= 3 {
	// 	return nil, errors.ErrLoginBlockedAccount
	// }

	// Check if password is correct or not

	if err := bcrypt.CompareHashAndPassword([]byte(ui.Password), []byte(req.Password)); err != nil {
		// Now increase failed login counts
		if err := s.rep.IncreaseFailedLoginCount(ctx, req.Username); err != nil {
			return nil, fmt.Errorf("service/v1 login failed to increase unsuccessful login count: %v", err)
		}
		// check if user is blocked
		// if ui.FailedLogins == 2 {
		// 	return nil, errors.ErrAccountBlocked
		// }
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
