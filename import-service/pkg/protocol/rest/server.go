// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package rest

import (
	"context"
	"crypto/rsa"
	"net/http"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"os"
	"os/signal"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, verifyKey *rsa.PublicKey, httpPort, uploadDir string) error {
	router := httprouter.New()
	h := handler{
		dir: uploadDir,
	}
	// TODO add a import handler here
	router.POST("/api/v1/import", h.uploadHandler)

	srv := &http.Server{
		Addr: ":" + httpPort,
		Handler: rest_middleware.AddCORS([]string{"*"},
			rest_middleware.AddLogger(logger.Log,
				//rest_middleware.ValidateAuth(verifyKey,&ochttp.Handler{Handler: router}),
				&ochttp.Handler{Handler: router},
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

	logger.Log.Info("starting auth-service - ", zap.String("port", httpPort))
	return srv.ListenAndServe()
}
