package grpc

import (
	"context"
	"crypto/rsa"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	rest_middleware "optisam-backend/common/optisam/middleware/rest"

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
		if userClaims, ok := rest_middleware.RetrieveClaims(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "user-id", userClaims.UserID)
			ctx = metadata.AppendToOutgoingContext(ctx, "user-role", string(userClaims.Role))
			for _, scope := range userClaims.Socpes {
				ctx = metadata.AppendToOutgoingContext(ctx, "user-scopes", scope)
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		if userClaims, ok := RetrieveClaims(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "user-id", userClaims.UserID)
			ctx = metadata.AppendToOutgoingContext(ctx, "user-role", string(userClaims.Role))
			for _, scope := range userClaims.Socpes {
				ctx = metadata.AppendToOutgoingContext(ctx, "user-scopes", scope)
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}
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

// To call self service rpc Need to add claims or add claims when you have metadata
func AddClaimsInContext(ctx context.Context, verifyKey *rsa.PublicKey, apiKey string) (context.Context, error) {
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
		delete(md, "authorization")
		md["x-api-key"] = []string{apiKey}
		md["user-id"] = []string{customClaims.UserID}
		md["user-role"] = []string{string(customClaims.Role)}
		md["user-scopes"] = customClaims.Socpes
		// md["Access-Control-Allow-Headers"] = []string{"X-Requested-With,content-type, Accept"}
		// ctx = metadata.NewIncomingContext(ctx, md)
		return AddClaims(ctx, customClaims), nil
	}
	return nil, status.Error(codes.Unauthenticated, "NoAuthorizationFound")
}
