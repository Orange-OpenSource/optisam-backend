package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/generators/access"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/server"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/stores/client"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/stores/token"
	repv1_postgres "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1/postgres"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/redis"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/generator"

	redisClient "github.com/go-redis/redis/v8"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"go.uber.org/zap"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/protocol/rest"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/service/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/buildinfo"
	gconn "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/healthcheck"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/jaeger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	postgres "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/prometheus"

	kafkaConnector "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"

	"contrib.go.opencensus.io/integrations/ocsql"

	// pq driver
	"github.com/confluentinc/confluent-kafka-go/kafka"
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
// nolint: funlen, gocyclo
func RunServer() error {
	config.Configure(viper.GetViper(), pflag.CommandLine)

	pflag.Parse()
	isLocal := false
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
	p, err := kafkaConnector.BuildProducer(cfg.Kafka)
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
					fmt.Printf("Produced event to topic: %s \n",
						*ev.TopicPartition.Topic)
				}
			}
		}
	}()

	// Create database connection.
	db, err := postgres.NewConnection(postgres.Config{
		Host: cfg.Database.Host,
		Port: cfg.Database.Port,
		Name: cfg.Database.User.Name,
		User: cfg.Database.User.User,
		Pass: cfg.Database.User.Pass,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	// Verify connection.
	if err = db.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		return fmt.Errorf("failed to verify connection to PostgreSQL: %v", zap.Error(err))
	}
	logger.Log.Info("Postgres connection verified to", zap.Any("", cfg.Database.Host))

	// Record DB stats every 5 seconds until we exit
	defer ocsql.RecordStats(db, 5*time.Second)()

	// Register database health check
	{
		check, error := checkers.NewSQL(&checkers.SQLConfig{Pinger: db})
		if error != nil {
			return fmt.Errorf("failed to create health checker: %v", error.Error())
		}
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
	// defer db.Close()
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
			logger.Log.Fatal("Jaeger Exporter Error", zap.String("reason", error.Error()))
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

	optisamDB := repv1_postgres.NewRepository(db, redisC)

	service := v1.NewAuthServiceServer(optisamDB, cfg, grpcClientMap, p)

	oauth2Server := server.NewServer(token.NewStore(), client.NewStore(), access.NewGenerator(generator, service))

	// server
	logger.Log.Sugar().Infow("%s - grpc port,%s - http port", cfg.GRPCPort, cfg.HTTPPort)
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Debug("Recovered in RunServer", r)
		}
	}()
	return rest.RunServer(ctx, service, oauth2Server, cfg.HTTPPort, cfg)
}
