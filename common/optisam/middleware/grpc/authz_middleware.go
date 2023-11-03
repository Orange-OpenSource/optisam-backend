package grpc

import (
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/opa"

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
