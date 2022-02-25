package grpc

import (
	"context"
	"crypto/rsa"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ChainedWithAdminFilter add admin rights filter along with other filters
func ChainedWithAdminFilter(logger *zap.Logger, verifyKey *rsa.PublicKey, apiKey string, a AdminRightsRequiredFunc) []grpc.ServerOption {

	// Shared options for the logger, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(logger)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(logger, o...),
		LoggingUnaryServerInterceptor(),
		grpc_auth.UnaryServerInterceptor(authHandler(verifyKey, apiKey)),
		//authorizationServerInterceptor(p),
		grpc_validator.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_zap.StreamServerInterceptor(logger, o...),
		grpc_auth.StreamServerInterceptor(authHandler(verifyKey, apiKey)),
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_validator.StreamServerInterceptor(),
		grpc_recovery.StreamServerInterceptor(),
	}

	if a != nil {
		u := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			if a(info.FullMethod) {
				if err := validateAdmin(ctx); err != nil {
					return nil, err
				}
			}
			return handler(ctx, req)
		}
		unaryInterceptors = append(unaryInterceptors, u)
		s := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if a(info.FullMethod) {
				if err := validateAdmin(ss.Context()); err != nil {
					return err
				}
			}
			return handler(srv, ss)
		}
		streamInterceptors = append(streamInterceptors, s)
	}

	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc_middleware.WithStreamServerChain(streamInterceptors...),
	}
}

// // Chanined returns all unary  middleware for rpc
// func Chanined(logger *zap.Logger, verifyKey *rsa.PublicKey) []grpc.ServerOption {
// 	return ChainedWithAdminFilter(logger, verifyKey, nil)
// }
