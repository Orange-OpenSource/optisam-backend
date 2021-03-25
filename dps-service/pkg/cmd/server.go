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
	"optisam-backend/common/optisam/buildinfo"
	"optisam-backend/common/optisam/cron"
	"optisam-backend/common/optisam/healthcheck"
	"optisam-backend/common/optisam/iam"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"optisam-backend/common/optisam/prometheus"
	"optisam-backend/dps-service/pkg/config"
	cronJob "optisam-backend/dps-service/pkg/poller"
	"optisam-backend/dps-service/pkg/protocol/grpc"
	"optisam-backend/dps-service/pkg/protocol/rest"
	repo "optisam-backend/dps-service/pkg/repository/v1/postgres"

	//query "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	gconn "optisam-backend/common/optisam/grpc"
	v1 "optisam-backend/dps-service/pkg/service/v1"
	apiworker "optisam-backend/dps-service/pkg/worker/api_worker"
	constants "optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/defer_worker"
	fileworker "optisam-backend/dps-service/pkg/worker/file_worker"
	"os"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"github.com/gobuffalo/packr/v2"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"

	//worker library
	"optisam-backend/common/optisam/workerqueue"

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
	Queue      *workerqueue.Queue
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

	var cfg *config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal configuration: %v", err)
	}

	//if Dir not exist then creates the archive directory
	if _, err = os.Stat(cfg.ArchiveLocation); os.IsNotExist(err) {
		err = os.Mkdir(cfg.ArchiveLocation, os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create archive directory at [", cfg.ArchiveLocation, "]")
		}
		log.Println("Archive directory created : ", cfg.ArchiveLocation)
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

	if cfg.MaxApiWorker == 0 {
		cfg.MaxApiWorker = 25 //Default api worker count
		logger.Log.Info("max api worker set default : 25 ")
	}
	if cfg.MaxFileWorker == 0 {
		cfg.MaxFileWorker = 5 //Default api worker count
		logger.Log.Info("max MaxFileWorker set default :  5 ")
	}
	if cfg.MaxDeferWorker == 0 {
		cfg.MaxDeferWorker = 10 //Default api worker count
		logger.Log.Info("max MaxDeferWorker set default : 10 ")
	}

	db, err := postgres.NewConnection(*cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Run Migration
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("migrations", "./../../pkg/repository/v1/postgres/schema"),
	}
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	repo.SetDpsRepository(db)
	dbObj, err := repo.GetDpsRepository()
	if err != nil {
		log.Println("Failed to get db client", err)
		return err
	}

	// Verify connection.
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to verify connection to PostgreSQL: %v", err.Error())
	}
	fmt.Printf("Postgres connection verified to %+v \n\n", cfg.Postgres)

	//GRPC Connections
	grpcClientMap, err := gconn.GetGRPCConnections(ctx, cfg.GrpcServers)
	if err != nil {
		logger.Log.Fatal("Failed to initialize GRPC client")
	}
	//log.Printf(" config %+v  grpcConn %+v", cfg, grpcClientMap)
	for _, conn := range grpcClientMap {
		defer conn.Close()
	}

	//Worker Queue
	Queue, err = workerqueue.NewQueue(ctx, constants.DPSQUEUE, db, cfg.WorkerQueue)
	if err != nil {
		return fmt.Errorf("failed to create worker queue: %v", err)
	}

	for i := 0; i < cfg.MaxFileWorker; i++ {
		w := fileworker.NewWorker(constants.FILEWORKER, Queue, db)
		Queue.RegisterWorker(ctx, w)
	}

	for i := 0; i < cfg.MaxApiWorker; i++ {
		w := apiworker.NewWorker(constants.APIWORKER, Queue, db, grpcClientMap, cfg.GrpcServers.Timeout)
		Queue.RegisterWorker(ctx, w)
	}

	for i := 0; i < cfg.MaxDeferWorker; i++ {
		w := defer_worker.NewWorker(constants.DEFERWORKER, Queue, db)
		Queue.RegisterWorker(ctx, w)
	}

	//All worker will wait till all registration will completed , to avoid concurrent map read write errors
	Queue.IsWorkerRegCompleted = true

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
	if cfg.Environment == "development" || cfg.Environment == "INTEGRATION" || cfg.Debug {
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

	v1API := v1.NewDpsServiceServer(dbObj.Queries, *Queue, grpcClientMap)
	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(cfg.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	//This is one time
	cron.CronConfigInit(cfg.Cron)

	// cron Job
	cronJob.Init(*Queue, fmt.Sprintf("http://%s/api/v1/token", cfg.HttpServers.Address["auth"]), cfg.FilesLocation, v1API, verifyKey)

	// Below command will trigger the cron job as soon as the service starts
	cronJob.Job()
	cron.AddCronJob(cronJob.Job)

	// get Authorization Policy
	authZPolicies, err := iam.NewOPA(ctx, cfg.IAM.RegoPath)
	if err != nil {
		logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	}

	config.SetConfig(*cfg)

	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, verifyKey)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, verifyKey, authZPolicies, cfg.IAM.APIKey)
}
