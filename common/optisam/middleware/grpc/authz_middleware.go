// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"optisam-backend/common/optisam/opa"

	"github.com/open-policy-agent/opa/rego"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func authorizationServerInterceptor(p *rego.PreparedEvalQuery) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userClaims, ok := RetrieveClaims(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid claims")
		}
		// Authorize
		authorized, err := opa.EvalAuthZ(ctx, p, opa.AuthzInput{Role: string(userClaims.Role), MethodFullName: info.FullMethod})
		if err != nil || !authorized {
			return nil, status.Errorf(codes.PermissionDenied, "Access to %s denied: %v", info.FullMethod, err)
		}
		return handler(ctx, req)

	}
}
