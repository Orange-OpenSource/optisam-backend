// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	os_logger "optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminRightsRequiredFunc returns true is admin rights are required for a perticular grpc
type AdminRightsRequiredFunc func(fullMethod string) bool

func validateAdmin(ctx context.Context) error {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
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
