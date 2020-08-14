// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"net/http"
	"optisam-backend/common/optisam/grpc"
	"optisam-backend/common/optisam/iam"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"optisam-backend/import-service/pkg/config"
	v1 "optisam-backend/import-service/pkg/service/v1"
	"os"
	"os/signal"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, config *config.Config) error {
	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(config.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	// get Authorization Policy
	authZPolicies, err := iam.NewOPA(ctx, config.IAM.RegoPath)
	if err != nil {
		logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	}
	router := httprouter.New()
	grpcClientMap, err := grpc.GetGRPCConnections(ctx, config.GRPCServers)
	if err != nil {
		logger.Log.Fatal("Failed to initialize GRPC client")
	}
	h := v1.NewImportServiceServer(grpcClientMap, config)
	// TODO add a import handler here
	router.POST("/api/v1/import/data", h.UploadDataHandler)
	router.POST("/api/v1/import/metadata", h.UploadMetaDataHandler)

	srv := &http.Server{
		Addr: ":" + config.HTTPPort,
		Handler: rest_middleware.AddCORS([]string{"*"},
			rest_middleware.AddLogger(logger.Log,
				rest_middleware.ValidateAuth(verifyKey,
					rest_middleware.ValidateAuthZ(authZPolicies, &ochttp.Handler{Handler: router})),
			)),
	}
	//   Handler:router,

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	logger.Log.Info("starting import-service - ", zap.String("port", config.HTTPPort))
	return srv.ListenAndServe()
}
