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

	gconn "optisam-backend/common/optisam/grpc"
	v1 "optisam-backend/dps-service/pkg/service/v1"
	apiworker "optisam-backend/dps-service/pkg/worker/api_worker"
	constants "optisam-backend/dps-service/pkg/worker/constants"
	deferworker "optisam-backend/dps-service/pkg/worker/defer_worker"
	fileworker "optisam-backend/dps-service/pkg/worker/file_worker"
	"os"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"github.com/gobuffalo/packr/v2"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"

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

	// if Dir not exist then creates the archive directory
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
	if error := logger.Init(cfg.Log.LogLevel, cfg.Log.LogTimeFormat); error != nil {
		return fmt.Errorf("failed to initialize logger: %v", error)
	}

	err = cfg.Validate()
	if err != nil {
		logger.Log.Error(err.Error())
		os.Exit(3)
	}

	ctx := context.Background()
	config.SetConfig(*cfg)
	if cfg.MaxAPIWorker == 0 {
		cfg.MaxAPIWorker = 25 // Default api worker count
		logger.Log.Info("max api worker set default : 25 ")
	}
	if cfg.MaxFileWorker == 0 {
		cfg.MaxFileWorker = 5 // Default api worker count
		logger.Log.Info("max MaxFileWorker set default :  5 ")
	}
	if cfg.MaxDeferWorker == 0 {
		cfg.MaxDeferWorker = 10 // Default api worker count
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
	dbObj := repo.GetDpsRepository()

	// Verify connection.
	if error := db.Ping(); error != nil {
		return fmt.Errorf("failed to verify connection to PostgreSQL: %v", error.Error())
	}
	fmt.Printf("Postgres connection verified to %+v \n\n", cfg.Postgres.Host)

	// GRPC Connections
	grpcClientMap, err := gconn.GetGRPCConnections(ctx, cfg.GrpcServers)
	if err != nil {
		logger.Log.Fatal("Failed to initialize GRPC client")
	}
	for _, conn := range grpcClientMap {
		defer conn.Close()
	}

	// Worker Queue
	Queue, err = workerqueue.NewQueue(ctx, constants.DPSQUEUE, db, cfg.WorkerQueue)
	if err != nil {
		return fmt.Errorf("failed to create worker queue: %v", err)
	}

	for i := 0; i < cfg.MaxFileWorker; i++ {
		w := fileworker.NewWorker(constants.FILEWORKER, Queue, db)
		Queue.RegisterWorker(ctx, w)
	}

	for i := 0; i < cfg.MaxAPIWorker; i++ {
		w := apiworker.NewWorker(constants.APIWORKER, Queue, db, grpcClientMap, cfg.GrpcServers.Timeout, cfg.FilesLocation, cfg.ArchiveLocation)
		Queue.RegisterWorker(ctx, w)
	}

	for i := 0; i < cfg.MaxDeferWorker; i++ {
		w := deferworker.NewWorker(constants.DEFERWORKER, Queue, db, grpcClientMap)
		Queue.RegisterWorker(ctx, w)
	}

	logger.Log.Error("total Worker started", zap.Any("fileWorker", cfg.MaxFileWorker), zap.Any("apiWorker", cfg.MaxAPIWorker), zap.Any("deferWorker", cfg.MaxDeferWorker))
	// All worker will wait till all registration will completed , to avoid concurrent map read write errors
	Queue.IsWorkerRegCompleted = true

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

	v1API := v1.NewDpsServiceServer(dbObj, Queue, grpcClientMap)
	// get the verify key to validate jwt
	verifyKey, err := iam.GetVerifyKey(cfg.IAM)
	if err != nil {
		logger.Log.Fatal("Failed to get verify key")
	}

	// This is one time
	cron.ConfigInit(cfg.Cron)

	// cron Job
	cronJob.Init(*Queue, fmt.Sprintf("http://%s/api/v1/token", cfg.HTTPServers.Address["auth"]), cfg.FilesLocation, cfg.ArchiveLocation, cfg.RawdataLocation, v1API, verifyKey, cfg.IAM.APIKey, dbObj, cfg.WaitLimitCount)

	// Below command will trigger the cron job as soon as the service starts
	cronJob.Job()
	cron.AddCronJob(cronJob.Job)

	// get Authorization Policy
	authZPolicies, err := iam.NewOPA(ctx, cfg.IAM.RegoPath)
	if err != nil {
		logger.Log.Fatal("Failed to Load RBAC policies", zap.Error(err))
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()

	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, verifyKey)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, verifyKey, authZPolicies, cfg.IAM.APIKey)
}
