// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"optisam-backend/auth-service/pkg/oauth2/generators/access"
	"optisam-backend/auth-service/pkg/oauth2/server"
	"optisam-backend/auth-service/pkg/oauth2/stores/client"
	"optisam-backend/auth-service/pkg/oauth2/stores/token"
	repv1_postgres "optisam-backend/auth-service/pkg/repository/v1/postgres"
	"optisam-backend/common/optisam/token/generator"
	"os"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"go.uber.org/zap"

	"optisam-backend/auth-service/pkg/config"
	"optisam-backend/auth-service/pkg/protocol/rest"
	v1 "optisam-backend/auth-service/pkg/service/v1"
	"optisam-backend/common/optisam/buildinfo"
	"optisam-backend/common/optisam/healthcheck"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/logger"
	postgres "optisam-backend/common/optisam/postgres"
	"optisam-backend/common/optisam/prometheus"

	"contrib.go.opencensus.io/integrations/ocsql"

	//postgres library
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
func RunServer() error {
	config.Configure(viper.GetViper(), pflag.CommandLine)

	pflag.Parse()
	if os.Getenv("ENV") == "prod" {
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
	if err := logger.Init(cfg.Log.LogLevel, cfg.Log.LogTimeFormat); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	err = cfg.Validate()
	if err != nil {
		logger.Log.Error(err.Error())

		os.Exit(3)
	}

	ctx := context.Background()

	// Register SQL stat views
	ocsql.RegisterAllViews()

	// Create database connection.
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Record DB stats every 5 seconds until we exit
	defer ocsql.RecordStats(db, 5*time.Second)()

	// Register database health check
	{
		check, err := checkers.NewSQL(&checkers.SQLConfig{Pinger: db})
		if err != nil {
			return fmt.Errorf("failed to create health checker: %v", err.Error())
		}
		err = healthChecker.AddCheck(&health.Config{
			Name:     "postgres",
			Checker:  check,
			Interval: time.Duration(3) * time.Second,
			Fatal:    true,
		})
		if err != nil {
			return fmt.Errorf("failed to add health checker: %v", err.Error())
		}
	}

	// Register http health check
	{
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
	//defer db.Close()
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

		exporter, err := prometheus.NewExporter(cfg.Instrumentation.Prometheus.Config)
		if err != nil {
			logger.Log.Fatal("Prometheus Exporter Error")
		}
		view.RegisterExporter(exporter)
		instrumentationRouter.Handle("/metrics", exporter)
	}

	// Trace everything in development environment or when debugging is enabled
	if cfg.Environment == "development" || cfg.Debug {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	// Configure Jaeger
	if cfg.Instrumentation.Jaeger.Enabled {
		logger.Log.Info("jaeger exporter enabled")

		exporter, err := jaeger.NewExporter(cfg.Instrumentation.Jaeger.Config)
		if err != nil {
			logger.Log.Fatal("Jaeger Exporter Error", zap.String("reason", err.Error()))
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
	)
	if err != nil {
		logger.Log.Error("Failed to register server stats view")
	}

	// Run Instumentation Server
	instrumentationserver := &http.Server{
		Addr:    cfg.Instrumentation.Addr,
		Handler: instrumentationRouter,
	}
	go func() {
		_ = instrumentationserver.ListenAndServe()
	}()

	generator, err := generator.NewTokenGenerator(cfg.JWTPrivateKey)
	if err != nil {
		logger.Log.Fatal("cannot create token generator", zap.String("reason", err.Error()))
	}

	optisamDB := repv1_postgres.NewRepository(db)
	service := v1.NewAuthServiceServer(optisamDB)

	oauth2Server := server.NewServer(token.NewStore(), client.NewStore(), access.NewGenerator(generator, service))

	// server
	fmt.Printf("%s - grpc port,%s - http port", cfg.GRPCPort, cfg.HTTPPort)
	return rest.RunServer(ctx, service, oauth2Server, cfg.HTTPPort)
}
