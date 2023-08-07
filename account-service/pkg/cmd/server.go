package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	redisClient "github.com/go-redis/redis/v8"

	// "sample-service/pkg/middleware/logger"
	"optisam-backend/account-service/pkg/config"
	"optisam-backend/account-service/pkg/protocol/grpc"
	"optisam-backend/account-service/pkg/protocol/rest"
	repo "optisam-backend/account-service/pkg/repository/v1/postgres"
	v1 "optisam-backend/account-service/pkg/service/v1"
	"optisam-backend/common/optisam/buildinfo"
	"optisam-backend/common/optisam/healthcheck"
	"optisam-backend/common/optisam/iam"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"optisam-backend/common/optisam/prometheus"
	"optisam-backend/common/optisam/redis"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"go.uber.org/zap"

	gconn "optisam-backend/common/optisam/grpc"

	"contrib.go.opencensus.io/integrations/ocsql"
	// pq driver
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// nolint: gochecknoglobals
var (
	version    string
	commitHash string
	buildDate  string
)

// nolint: gochecknoinits
func init() {
	pflag.Bool("version", false, "Show version information")
}

// RunServer runs gRPC server and HTTP gateway
// nolint: funlen, gocyclo
func RunServer() error {
	config.Configure(viper.GetViper(), pflag.CommandLine)
	isLocal := false
	pflag.Parse()
	if os.Getenv("ENV") == "prod" { // nolint: gocritic
		viper.SetConfigName("config-prod")
	} else if os.Getenv("ENV") == "performance" {
		viper.SetConfigName("config-performance")
	} else if os.Getenv("ENV") == "int" {
		viper.SetConfigName("config-int")
	} else if os.Getenv("ENV") == "dev" {
		viper.SetConfigName("config-dev")
	} else if os.Getenv("ENV") == "pc" {
		viper.SetConfigName("config-pc")
	} else {
		viper.SetConfigName("config-local")
		isLocal = true
	}

	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal configuration: %v", err)
	}

	buildInfo := buildinfo.New(version, commitHash, buildDate)
	// Instumentation Handler
	instrumentationRouter := http.NewServeMux()
	instrumentationRouter.Handle("/version", buildinfo.Handler(buildInfo))

	// configure health checker
	healthChecker := healthcheck.New()
	instrumentationRouter.Handle("/healthz", healthcheck.Handler(healthChecker))

	// initialize logger
	if error := logger.Init(cfg.Log.LogLevel, cfg.Log.LogTimeFormat); error != nil {
		return fmt.Errorf("failed to initialize logger: %v", error)
	}

	err = cfg.Validate()
	if err != nil {
		logger.Log.Error(err.Error())

		os.Exit(3)
	}

	ctx := context.Background()

	// Register SQL stat views
	ocsql.RegisterAllViews()

	// Create database connection.and exec migratrions
	db, err := postgres.ConnectDBExecMig(cfg.Database)
	if err != nil {
		logger.Log.Error("failed to ConnectDBExecMig error: %v", zap.Any("", err.Error()))
		return fmt.Errorf("failed to ConnectDBExecMig error: %v", err.Error())
	}
	// defer db.Close()
	defer func() {
		db.Close()
		// Wait to 4 seconds so that the traces can be exported
		waitTime := 2 * time.Second
		log.Printf("Waiting for %s seconds to ensure all traces are exported before exiting", waitTime)
		<-time.After(waitTime)
	}()

	// Record DB stats every 5 seconds until we exit
	defer ocsql.RecordStats(db, 5*time.Second)()

	// Register database health check

	check, error := checkers.NewSQL(&checkers.SQLConfig{Pinger: db})
	if error != nil {
		error = healthChecker.AddCheck(&health.Config{
			Name:     "postgres",
			Checker:  check,
			Interval: time.Duration(3) * time.Second,
			Fatal:    true,
		})
		if error != nil {
			return fmt.Errorf("failed to add health checker: %v", error.Error())
		}
	}

	// GRPC Connections
	grpcClientMap, err := gconn.GetGRPCConnections(ctx, cfg.GrpcServers)
	if err != nil {
		logger.Log.Fatal("Failed to initialize GRPC client")
	}
	logger.Log.Info("grpc Connections list", zap.Any("grpcConnections", grpcClientMap))
	for _, conn := range grpcClientMap {
		defer conn.Close()
	}

	// Register http health check
	{
		check, error := checkers.NewHTTP(&checkers.HTTPConfig{URL: &url.URL{Scheme: "http", Host: "localhost:8080"}})
		if error != nil {
			return fmt.Errorf("failed to create health checker: %v", error.Error())
		}
		error = healthChecker.AddCheck(&health.Config{
			Name:     "Http Server",
			Checker:  check,
			Interval: time.Duration(3) * time.Second,
			Fatal:    true,
		})
		if error != nil {
			return fmt.Errorf("failed to add health checker: %v", error.Error())
		}
	}
	redisC := &redisClient.Client{}
	if isLocal {
		redisC = redis.NewConnection(*cfg.Redis)
		_, err = redisC.Ping(ctx).Result()
		if err != nil {
			logger.Log.Fatal("Failed to connect redis")
		} else {
			logger.Log.Info("redis server connected at:" + cfg.Redis.RedisHost)
		}
		//defer redisClient.Close()
	} else {
		redisC = redis.NewConnectionSentinel(*cfg.Redis)
		_, err = redisC.Ping(ctx).Result()
		if err != nil {
			logger.Log.Fatal("Failed to connect redis sentinel")
		} else {
			logger.Log.Info("redis server connected")
		}
	}
	defer redisC.Close()
	// defer db.Close()
	defer func() {
		//	db.Close()
		// Wait to 4 seconds so that the traces can be exported
		// TODO: Investigate if 4 seconds is a reasonable time.
		waitTime := 4 * time.Second
		log.Printf("Waiting for %s seconds to ensure all traces are exported before exiting", waitTime)
		<-time.After(waitTime)
	}()
	// Configure Prometheus
	if cfg.Instrumentation.Prometheus.Enabled {
		logger.Log.Info("prometheus exporter enabled")

		exporter, error := prometheus.NewExporter(cfg.Instrumentation.Prometheus.Config)
		if error != nil {
			logger.Log.Fatal("Prometheus Exporter Error")
		}
		view.RegisterExporter(exporter)
		instrumentationRouter.Handle("/metrics", exporter)
	}

	// Trace everything in development environment or when debugging is enabled
	if cfg.Environment == "DEVELOPMENT" || cfg.Environment == "INTEGRATION" || cfg.Debug {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	// Configure Jaeger
	if cfg.Instrumentation.Jaeger.Enabled {
		logger.Log.Info("jaeger exporter enabled")

		exporter, error := jaeger.NewExporter(cfg.Instrumentation.Jaeger.Config)
		if error != nil {
			logger.Log.Fatal("Jaeger Exporter Error")
		}
		trace.RegisterExporter(exporter)
	}

	// Register stat views
	err = view.Register(
		// HTTP
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,

		// GRPC
		ocgrpc.ServerReceivedBytesPerRPCView,
		ocgrpc.ServerSentBytesPerRPCView,
		ocgrpc.ServerLatencyView,
		ocgrpc.ServerCompletedRPCsView,
	)
	if err != nil {
		logger.Log.Error("Failed to register server stats view")
	}

	// Run Instumentation Server
	instrumentationServer := &http.Server{
		Addr:    cfg.Instrumentation.Addr,
		Handler: instrumentationRouter,
	}
	go func() {
		_ = instrumentationServer.ListenAndServe()
	}()

	v1API := v1.NewAccountServiceServer(repo.NewAccountRepository(db, redisC), grpcClientMap)
	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(cfg.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}
	// get Authorization Policy
	// authZPolicies, err := iam.NewOPA(ctx, cfg.IAM.RegoPath)
	// if err != nil {
	// 	logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	// }

	// run HTTP gateway
	fmt.Printf("%s - grpc port,%s - http port", cfg.GRPCPort, cfg.HTTPPort)
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, verifyKey)
	}()
	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, verifyKey, cfg.IAM.APIKey, v1.AdminRightsRequired)
}
