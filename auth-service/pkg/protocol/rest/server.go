// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"net/http"
	"optisam-backend/auth-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"

	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3/server"

	"github.com/julienschmidt/httprouter"
	"go.opencensus.io/plugin/ochttp"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, service v1.AuthService, serv *server.Server, httpPort string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	router := httprouter.New()

	handler := newHandler(service, serv)

	router.POST("/api/v1/token", handler.token)

	srv := &http.Server{
		Addr: ":" + httpPort,
		Handler: rest_middleware.AddCORS([]string{"*"},
			rest_middleware.AddLogger(logger.Log, &ochttp.Handler{Handler: router})),
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

	logger.Log.Info("starting auth-service - ", zap.String("port", httpPort))
	return srv.ListenAndServe()
}
