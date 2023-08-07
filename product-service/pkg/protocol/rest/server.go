package rest

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/http/pprof"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"os"
	"os/signal"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
// nolint: funlen, gocyclo, gosec
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
			logger.Log.Info("Shutdown Signal Receieved - HTTP")
			_, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel() // nolint: staticcheck

			_ = srv.Shutdown(ctx)
		}
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
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{}), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10 * 1024 * 1024))}
	conn, err := grpc.DialContext(ctx, "localhost:"+grpcPort, opts...)
	if err != nil {
		return nil, err
	}

	if err = v1.RegisterProductServiceHandler(ctx, muxGateway, conn); err != nil {
		return nil, err
	}
	return muxGateway, err
}
