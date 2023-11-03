package grpc

import (
	"context"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func GetGRPCConnections(ctx context.Context, c Config) (map[string]*grpc.ClientConn, error) {
	grpcGRPCConnections := make(map[string]*grpc.ClientConn)
	if c.Timeout == 0 {
		c.Timeout = 10
	}
	for key, val := range c.Address {
		var conn *grpc.ClientConn
		opts := []grpc.DialOption{grpc.WithInsecure(),
			grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: c.Timeout * time.Millisecond * 10}),
			grpc.WithChainUnaryInterceptor(middleware.AddAuthNClientInterceptor(c.APIKey)),
			grpc.WithStatsHandler(&ocgrpc.ClientHandler{})}
		conn, err := grpc.Dial(val, opts...)
		if err != nil {
			logger.Log.Error("did not connect:", zap.String(key, val), zap.Error(err))
			return nil, err
		}
		grpcGRPCConnections[key] = conn
	}

	return grpcGRPCConnections, nil
}
