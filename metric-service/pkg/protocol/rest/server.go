package rest

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	rest_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/rest"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string, verifyKey *rsa.PublicKey) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	muxHTTP := http.NewServeMux()
	gw, err := newGateway(ctx, grpcPort)
	if err != nil {
		logger.Log.Fatal("failed to register GRPC gateway", zap.String("reason", err.Error()))

	}
	muxHTTP.HandleFunc("/debug/pprof/trace", pprof.Trace)
	muxHTTP.Handle("/", gw)

	srv := &http.Server{
		Addr: ":" + httpPort,
		// Handler: &ochttp.Handler{
		Handler: &ochttp.Handler{Handler: rest_middleware.AddCORS([]string{"*"},
			// rest_middleware.ValidateAuth(verifyKey,
			// rest_middleware.AddLogger(logger.Log,
			muxHTTP),
		// ))},
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

func newGateway(ctx context.Context, grpcPort string) (http.Handler, error) {
	muxGateway := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{})}
	conn, err := grpc.DialContext(ctx, "localhost:"+grpcPort, opts...)
	if err != nil {
		return nil, err
	}

	if error := v1.RegisterMetricServiceHandler(ctx, muxGateway, conn); error != nil {
		return nil, error
	}
	return muxGateway, err
}
