// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"crypto/rsa"
	"net/http"
	v1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string, verifyKey *rsa.PublicKey) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{})}
	if err := v1.RegisterAccountServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger.Log.Fatal("failed to start HTTP gateway", zap.String("reason", err.Error()))
	}

	srv := &http.Server{
		Addr: ":" + httpPort,
		Handler: &ochttp.Handler{Handler: rest_middleware.AddCORS([]string{"*"},
			// rest_middleware.ValidateAuth(verifyKey,
			// rest_middleware.AddLogger(logger.Log,
			mux),
		// )},
		},
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
		}

		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	logger.Log.Info("starting HTTP/REST gateway...")
	return srv.ListenAndServe()
}
