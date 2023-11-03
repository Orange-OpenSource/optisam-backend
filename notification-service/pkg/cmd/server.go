package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	// "sample-service/pkg/middleware/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/protocol/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/protocol/rest"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/buildinfo"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/healthcheck"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/iam"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/jaeger"
	kafkaConnect "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/prometheus"

	"github.com/InVisionApp/go-health/checkers"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"

	//"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	gconn "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/grpc"
	v1Kafka "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/kafka/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1/postgres"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/service/v1"

	// pq driver

	"github.com/InVisionApp/go-health"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	gomail "gopkg.in/mail.v2"
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
	p, err := kafkaConnect.BuildProducer(cfg.Kafka, map[string]string{})
	if err != nil {
		logger.Log.Sugar().Debug("failed to open producer: %v", err)
		return fmt.Errorf("failed to open producer: %v", err)
	}
	defer p.Close()

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic: %s\n",
						*ev.TopicPartition.Topic)
				}
			}
		}
	}()
	// Run Instumentation Server
	instrumentationServer := &http.Server{
		Addr:    cfg.Instrumentation.Addr,
		Handler: instrumentationRouter,
	}
	go func() {
		_ = instrumentationServer.ListenAndServe()
	}()
	// Create database connection.
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
	rep := repo.NewNotificationRepository(db)
	//v1API := v1.NewNotificationServiceServer(&cfg, p, rep)

	// Settings for SMTP server
	d := gomail.NewDialer(cfg.SMTP.Host, int(cfg.SMTP.Port), cfg.SMTP.From, cfg.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	logger.Log.Info("SMTP connection created")

	v1API := v1.NewNotificationServiceServer(&cfg, p, rep, d)
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
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()

	err = v1Kafka.NotificationConsumer(v1API)
	if err != nil {
		logger.Log.Sugar().Debug("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, verifyKey)
	}()
	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, verifyKey, authZPolicies, cfg.IAM.APIKey)
}
