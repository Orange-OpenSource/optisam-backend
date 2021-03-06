// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"crypto/rsa"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/status"

	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

func authHandler(verifyKey *rsa.PublicKey, apiKey string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
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
			return ctxmanage.AddClaims(ctx, customClaims), nil
		}
		if xApiKey, ok := md["x-api-key"]; ok {
			if xApiKey[0] != apiKey {
				return nil, status.Error(codes.Unauthenticated, "InvalidAPIKeyError")
			}
			//TODO service to service call should manage scopes
			return ctxmanage.AddClaims(ctx, &claims.Claims{UserID: "System", Role: claims.RoleSuperAdmin, Socpes: []string{"OFR", "OSP"}}), nil
		}
		return nil, status.Error(codes.Unauthenticated, "NoAuthNError")
	}

}

// func parseToken(tokenStr string) {
// 	customClaims := claims.Claims{}
// 	token, err := jwt.ParseWithClaims(tokenStr, &customClaims, func(token *jwt.Token) (interface{}, error) {
// 		return verifyKey, nil
// 	})
// 	if err != nil {
// 		logger.Log.Error("grpc/authHandler - failed to parse token", zap.String("reason", err.Error()))
// 		return nil, status.Error(codes.Unauthenticated, "cannot parse token")
// 	}

// 	if !token.Valid {
// 		return nil, status.Error(codes.Unauthenticated, "token is not valid")
// 	}
// }
