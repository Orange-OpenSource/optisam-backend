package rest

import (
	"context"
	"net/http"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	rest_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/rest"

	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3/server"

	"github.com/julienschmidt/httprouter"
	"go.opencensus.io/plugin/ochttp"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, service v1.AuthService, serv *server.Server, httpPort string, cfg config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	router := httprouter.New()

	handler := newHandler(service, serv, cfg)

	router.POST("/api/v1/token", handler.token)
	router.GET("/api/v1/activate_account", handler.activateAccount)
	router.GET("/api/v1/reset_password", handler.resetPassword)
	router.POST("/api/v1/set_password", handler.setPassword)
	router.POST("/api/v1/forgot_password", handler.forgotPassword)

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
