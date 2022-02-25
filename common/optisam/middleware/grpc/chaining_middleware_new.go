package grpc

import (
	"crypto/rsa"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Chained for linking all grpc interceptor
func Chained(logger *zap.Logger, verifyKey *rsa.PublicKey, p *rego.PreparedEvalQuery, apiKey string) []grpc.ServerOption {

	// alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	// Shared options for the logger, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.

	// grpc_zap.ReplaceGrpcLoggerV2WithVerbosity(logger, 1)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		// grpc_zap.PayloadUnaryServerInterceptor(logger, alwaysLoggingDeciderServer),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(logger, o...),
		// UserLoggingUnaryServerInterceptor(),
		LoggingUnaryServerInterceptor(),
		grpc_auth.UnaryServerInterceptor(authHandler(verifyKey, apiKey)),
		authorizationServerInterceptor(p),
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
	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc_middleware.WithStreamServerChain(streamInterceptors...),
	}
}
