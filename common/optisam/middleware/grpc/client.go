// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"crypto/rsa"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AddAuthNClientInterceptor(key string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", key)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AddContextSharingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		mdgrpc, _ := metadata.FromIncomingContext(ctx)
		mdrest, _ := metadata.FromOutgoingContext(ctx)
		if mdgrpc != nil {
			ctx = metadata.NewOutgoingContext(ctx, mdgrpc)
		}
		if mdrest != nil {
			ctx = metadata.NewOutgoingContext(ctx, mdrest)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

//To call self service rpc Need to add claims or add claims when you have metadata
func AddClaimsInContext(ctx context.Context, verifyKey *rsa.PublicKey) (context.Context, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if _, ok := md["authorization"]; ok {
		tokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			logger.Log.Error("grpc/authHandler - failed to get token, AuthFromMD", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Unauthenticated, "NoTokenError")
		}
		token, err := jwt.ParseWithClaims(tokenStr, &claims.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})
		if err != nil {
			logger.Log.Error("grpc/authHandler - failed to parse token", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Unauthenticated, "ParseTokenError")
		}

		if !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "InvalidTokenError")
		}

		customClaims, ok := token.Claims.(*claims.Claims)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "InvalidClaimsError")
		}
		return AddClaims(ctx, customClaims), nil
	} else {
		return nil, status.Error(codes.Unauthenticated, "NoAuthorizationFound")
	}
}
