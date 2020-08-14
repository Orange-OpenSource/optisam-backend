// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"context"
	"time"

	v1 "optisam-backend/application-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//RunClient to get a GPRC connection to the destined server
//Hanles X-API-KEY
//Needs to check from where to call this
func RunClient(apiKey string) error {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("localhost:8090", grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: time.Second * 5}),
	)
	if err != nil {
		logger.Log.Error("did not connect:", zap.Error(err))
		return err
	}
	defer conn.Close()
	c := v1.NewApplicationServiceClient(conn)
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
	resp, err := c.ListApplications(ctx, &v1.ListApplicationsRequest{PageNum: 1, PageSize: 10})
	if err != nil {
		logger.Log.Error("Error when calling ListApplications:", zap.Error(err))
		return err
	}
	logger.Log.Sugar().Infof("Response from server: %v", resp.GetTotalRecords())
	return nil
}
