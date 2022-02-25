package grpc

import (
	"context"
	"optisam-backend/common/optisam/logger"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type LoggerKey struct{}
type LoggerUserDetails struct {
	UserID string
	Role   string
}

// codeToLevel
func codeToLevel(code codes.Code) zapcore.Level {
	if code != codes.OK {
		// It is Error
		return zap.ErrorLevel
	}
	return grpc_zap.DefaultCodeToLevel(code)
}

// LoggingUnaryServerInterceptor to add custom logging fields
func LoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sc := trace.FromContext(ctx).SpanContext()
		ctxzap.AddFields(
			ctx,
			zap.String("trace_id", sc.TraceID.String()),
			zap.String("span_id", sc.SpanID.String()),
			zap.Bool("is_sampled", sc.IsSampled()),
		)
		if logger.GlobalLevel == zap.DebugLevel {
			ctxzap.AddFields(
				ctx,
				zap.Any("req_payload", req),
			)
		}
		resp, err := handler(ctx, req)
		if err == nil {
			if logger.GlobalLevel == zap.DebugLevel {
				ctxzap.AddFields(
					ctx,
					zap.Any("res_payload", resp),
				)
			}
		}

		return resp, err
	}
}
