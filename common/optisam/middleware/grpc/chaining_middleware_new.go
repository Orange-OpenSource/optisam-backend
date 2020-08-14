// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"crypto/rsa"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ChainedWithAdminFilter add admin rights filter along with other filters
func Chained(logger *zap.Logger, verifyKey *rsa.PublicKey, p *rego.PreparedEvalQuery, apiKey string) []grpc.ServerOption {
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	// Shared options for the logger, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(logger)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_zap.PayloadUnaryServerInterceptor(logger, alwaysLoggingDeciderServer),
		// grpc_zap.UnaryServerInterceptor(logger, o...),
		grpc_auth.UnaryServerInterceptor(authHandler(verifyKey, apiKey)),
		authorizationServerInterceptor(p),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_validator.UnaryServerInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_zap.StreamServerInterceptor(logger, o...),
		grpc_auth.StreamServerInterceptor(authHandler(verifyKey, apiKey)),
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_validator.StreamServerInterceptor(),
	}
	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc_middleware.WithStreamServerChain(streamInterceptors...),
	}
}
