package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	"net/http"
	"net/url"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/buildinfo"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/healthcheck"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/jaeger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/prometheus"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/protocol/rest"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	kafkaConnect "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres"
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
	p, err := kafkaConnect.BuildProducer(cfg.Kafka, map[string]string{
		"message.max.bytes":  "120971520",
		"message.timeout.ms": "1200000",
		"compression.type":   "gzip",
		//"max.block.ms":       "90000",
	})
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
	repo.SetImportRepository(db)
	dbObj := repo.GetImportRepository()
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()
	c, err := kafkaConnect.BuildConsumer(cfg.Kafka, map[string]string{})
	if err != nil {
		logger.Log.Sugar().Debug("failed to open consumer: %v", err)
		return fmt.Errorf("failed to open consumer: %v", err)
	}
	defer c.Close()

	return rest.RunServer(ctx, cfg, dbObj, p, c)
}
