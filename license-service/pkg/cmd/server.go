package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"optisam-backend/common/optisam/buildinfo"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/healthcheck"
	"optisam-backend/common/optisam/iam"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/prometheus"
	"optisam-backend/license-service/pkg/config"
	"optisam-backend/license-service/pkg/protocol/grpc"
	"optisam-backend/license-service/pkg/protocol/rest"
	repo "optisam-backend/license-service/pkg/repository/v1/dgraph"
	v1 "optisam-backend/license-service/pkg/service/v1"
	"os"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"go.uber.org/zap"

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

	pflag.Parse()
	if os.Getenv("ENV") == "prod" { // nolint: gocritic
		viper.SetConfigName("config-prod")
	} else if os.Getenv("ENV") == "pprod" {
		viper.SetConfigName("config-pprod")
	} else if os.Getenv("ENV") == "int" {
		viper.SetConfigName("config-int")
	} else if os.Getenv("ENV") == "dev" {
		viper.SetConfigName("config-dev")
	} else {
		viper.SetConfigName("config-local")
	}

	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var cfg *config.Config
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

	// Create Dgraph Connection
	dg, err := dgraph.NewDgraphConnection(cfg.Dgraph)
	if err != nil {
		return fmt.Errorf("failed to open connection with dgraph: %v", err)
	}
	fmt.Println("Dgraph connection verified to", cfg.Dgraph.Hosts)

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

	rep, err := repo.NewLicenseRepositoryWithTemplates(dg)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	v1API := v1.NewLicenseServiceServer(rep)

	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(cfg.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	// get Authorization Policy
	authZPolicies, err := iam.NewOPA(ctx, cfg.IAM.RegoPath)
	if err != nil {
		logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	}

	// run HTTP gateway
	fmt.Printf("%s - grpc port,%s - http port", cfg.GRPCPort, cfg.HTTPPort)
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, verifyKey)
	}()
	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, verifyKey, authZPolicies, cfg.IAM.APIKey)
}
