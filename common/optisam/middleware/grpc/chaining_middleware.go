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
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ChainedWithAdminFilter add admin rights filter along with other filters
func ChainedWithAdminFilter(logger *zap.Logger, verifyKey *rsa.PublicKey, a AdminRightsRequiredFunc) []grpc.ServerOption {

	// Shared options for the logger, with a custom gRPC code to log level function.
	o := []grpc_zap.Option{
		grpc_zap.WithLevels(codeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(logger)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_zap.UnaryServerInterceptor(logger, o...),
		grpc_auth.UnaryServerInterceptor(authHandler(verifyKey, "")),
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_validator.UnaryServerInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_zap.StreamServerInterceptor(logger, o...),
		grpc_auth.StreamServerInterceptor(authHandler(verifyKey, "")),
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_validator.StreamServerInterceptor(),
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

// Chanined returns all unary  middleware for rpc
func Chanined(logger *zap.Logger, verifyKey *rsa.PublicKey) []grpc.ServerOption {
	return ChainedWithAdminFilter(logger, verifyKey, nil)
}
