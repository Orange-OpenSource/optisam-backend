package grpc

import (
	"context"

	os_logger "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminRightsRequiredFunc returns true is admin rights are required for a perticular grpc
type AdminRightsRequiredFunc func(fullMethod string) bool

func validateAdmin(ctx context.Context) error {
	userClaims, ok := RetrieveClaims(ctx)
	if !ok {
		os_logger.Log.Error("ChaniedWithAdminFilter - validateAdmin - can not retrieve claims from context")
		return status.Error(codes.Unknown, "cannot find claims in context")
	}
	switch userClaims.Role {
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		return nil
	default:
		return status.Error(codes.PermissionDenied, "admin roles required")
	}
}
