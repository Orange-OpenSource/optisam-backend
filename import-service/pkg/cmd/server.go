package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"

	"net/http"
	"net/url"
	"optisam-backend/common/optisam/buildinfo"
	"optisam-backend/common/optisam/healthcheck"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/prometheus"
	"optisam-backend/import-service/pkg/config"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/import-service/pkg/protocol/rest"

	"time"

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
	}

	viper.AddConfigPath("/opt/config/")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
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
	err = logger.Init(cfg.Log.LogLevel, cfg.Log.LogTimeFormat)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		logger.Log.Error(err.Error())
		os.Exit(3)
	}

	ctx := context.Background()

	// Register http health check
	{
		// TODO change the port
		check, err := checkers.NewHTTP(&checkers.HTTPConfig{URL: &url.URL{Scheme: "http", Host: "localhost:8080"}})
		if err != nil {
			return fmt.Errorf("failed to create health checker: %v", err.Error())
		}
		err = healthChecker.AddCheck(&health.Config{
			Name:     "Http Server",
			Checker:  check,
			Interval: time.Duration(3) * time.Second,
			Fatal:    true,
		})
		if err != nil {
			return fmt.Errorf("failed to add health checker: %v", err.Error())
		}
	}

	// Configure Prometheus
	if cfg.Instrumentation.Prometheus.Enabled {
		logger.Log.Info("prometheus exporter enabled")

		exporter, err := prometheus.NewExporter(cfg.Instrumentation.Prometheus.Config)
		if err != nil {
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

		exporter, err := jaeger.NewExporter(cfg.Instrumentation.Jaeger.Config)
		if err != nil {
			logger.Log.Fatal("Jaeger Exporter Error")
		}
		trace.RegisterExporter(exporter)
	}

	// Register stat views
	if err := view.Register(
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
	); err != nil {
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

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()

	return rest.RunServer(ctx, cfg)
}
