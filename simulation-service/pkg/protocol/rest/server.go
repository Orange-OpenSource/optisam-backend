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
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	"optisam-backend/simulation-service/pkg/config"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string, verifyKey *rsa.PublicKey) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	m := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{})}
	if err := v1.RegisterSimulationServiceHandlerFromEndpoint(ctx, m, "localhost:"+grpcPort, opts); err != nil {
		logger.Log.Fatal("failed to start HTTP gateway", zap.String("reason", err.Error()))
	}

	r := mux.NewRouter()

	r.NotFoundHandler = m
	conn, err := openGrpcConnection(ctx)
	if err != nil {
		logger.Log.Fatal("failed to start internal grpc connection", zap.String("reason", err.Error()))
	}
	hdlr := &handler{
		client: v1.NewSimulationServiceClient(conn),
	}
	r.HandleFunc("/api/v1/config", makeHandler(ctx, hdlr.CreateConfigHandler)).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.Method == "POST"
	})
	r.HandleFunc("/api/v1/config/{config_id}", makeHandler(ctx, hdlr.UpdateConfigHandler)).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.Method == "PUT"
	})

	srv := &http.Server{
		Addr: ":" + httpPort,
		Handler: &ochttp.Handler{
			Handler: rest_middleware.AddCORS([]string{"*"},
				rest_middleware.ValidateAuth(verifyKey,
					rest_middleware.AddLogger(logger.Log, r),
				),
			),
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

func makeHandler(ctx context.Context, myHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		myHandler(ctx, w, r)
	}
}

func openGrpcConnection(ctx context.Context) (*grpc.ClientConn, error) {
	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		logger.Log.Error("failed to unmarshal configuration ", zap.Error(err))
		return nil, err
	}
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{})}
	conn, err := grpc.Dial("localhost:"+cfg.GRPCPort, opts...)
	if err != nil {
		logger.Log.Error("failed to start grpc conn ", zap.Error(err))
		return nil, err
	}
	return conn, nil
}
