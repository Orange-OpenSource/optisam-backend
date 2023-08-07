package rest

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"net/http"
	"net/http/pprof"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	repo "optisam-backend/catalog-service/pkg/repository/v1/postgres"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"os"
	"os/signal"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	accService "optisam-backend/account-service/pkg/api/v1"

	redisClient "github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string, verifyKey *rsa.PublicKey, database *sql.DB, grpcServers map[string]*grpc.ClientConn, authapi string, apiKey string, appCred config.Application, redisc *redisClient.Client) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	muxHTTP := http.NewServeMux()
	gw, err := newGateway(ctx, grpcPort)
	if err != nil {
		logger.Log.Fatal("failed to register GRPC gateway", zap.String("reason", err.Error()))

	}
	repo.SetProductCatalogRepository(database, redisc)
	dbObj := repo.GetProductCatalogRepository()
	handler := handler{Db: database,
		account:     accService.NewAccountServiceClient(grpcServers["account"]),
		AuthAPI:     authapi,
		VerifyKey:   verifyKey,
		APIKey:      apiKey,
		Application: appCred,
		pCRepo:      dbObj,
	}

	muxHTTP.HandleFunc("/debug/pprof/trace", pprof.Trace)
	muxHTTP.HandleFunc("/catalog/editor", http.HandlerFunc(handler.GetEditor))
	muxHTTP.HandleFunc("/catalog/editors", http.HandlerFunc(handler.ListEditors))
	muxHTTP.HandleFunc("/catalog/editornames", http.HandlerFunc(handler.ListEditorNames))
	muxHTTP.HandleFunc("/catalog/product", http.HandlerFunc(handler.GetProduct))
	muxHTTP.HandleFunc("/catalog/products", http.HandlerFunc(handler.GetProducts))
	muxHTTP.HandleFunc("/catalog/editorfilters", http.HandlerFunc(handler.GetEditorFilters))
	muxHTTP.HandleFunc("/catalog/index", http.HandlerFunc(handler.GetTesting))
	muxHTTP.HandleFunc("/catalog/productfilters", http.HandlerFunc(handler.GetProductFilters))

	muxHTTP.Handle("/", gw)

	srv := &http.Server{
		Addr: ":" + httpPort,
		// Handler: &ochttp.Handler{
		Handler: &ochttp.Handler{Handler: rest_middleware.AddCORS([]string{"*"},
			rest_middleware.AddLogger(logger.Log, muxHTTP)),
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

	if error := v1.RegisterProductCatalogHandler(ctx, muxGateway, conn); error != nil {
		return nil, error
	}
	return muxGateway, err
}
