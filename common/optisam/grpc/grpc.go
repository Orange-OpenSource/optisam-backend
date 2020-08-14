// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"optisam-backend/common/optisam/logger"
	middleware "optisam-backend/common/optisam/middleware/grpc"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func GetGRPCConnections(ctx context.Context, c Config) (map[string]*grpc.ClientConn, error) {
	grpcGRPCConnections := make(map[string]*grpc.ClientConn)
	if c.Timeout == 0 {
		c.Timeout = 10
	}
	logger.Log.Sugar().Info("config :", c)
	for key, val := range c.Address {
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(val, grpc.WithInsecure(),
			grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: c.Timeout * time.Second}),
			grpc.WithChainUnaryInterceptor(middleware.AddAuthNClientInterceptor(c.ApiKey)),
		)
		if err != nil {
			logger.Log.Error("did not connect:", zap.String(key, val), zap.Error(err))
			return nil, err
		}
		logger.Log.Info("grpc connection created with ", zap.String(key, val))
		grpcGRPCConnections[key] = conn
	}

	return grpcGRPCConnections, nil
}
