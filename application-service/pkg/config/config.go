package config

import (
	"optisam-backend/common/optisam/cron"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/grpc"
	"optisam-backend/common/optisam/iam"
	"optisam-backend/common/optisam/jaeger"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"optisam-backend/common/optisam/workerqueue"

	"os"
	"time"

	"optisam-backend/common/optisam/prometheus"

	"errors"
	"fmt"

	"optisam-backend/common/optisam/config"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config is configuration for Server
type Config struct {
	// Meaningful values are recommended (eg. production, development, staging, release/123, etc)
	Environment string

	// Turns on some debug functionality
	Debug bool

	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// for interservice rpc calls
	GrpcServers grpc.Config

	// Handles cron config
	Cron cron.Config

	// For interservice http calls(non grpc server)["ip:port"]
	HTTPServers httpConfg

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string

	// Dgraph connection information
	Dgraph *dgraph.Config

	// Database connection information
	Database postgres.Config

	// MaxAPIWorker
	MaxAPIWorker int

	// Log configuration
	Log logger.Config

	// Instrumentation configuration
	Instrumentation InstrumentationConfig

	// WorkerQueue holds queue config
	WorkerQueue workerqueue.QueueConfig

	// MaxRiskWorker
	MaxRiskWorker int

	AppParams AppParameters

	// IAM Configuration
	IAM iam.Config
	//Application cred.
	Application config.Application
}

type httpConfg struct {
	Address map[string]string
}

// InstrumentationConfig represents the instrumentation related configuration.
type InstrumentationConfig struct {
	// Instrumentation HTTP server address
	Addr string

	// Prometheus configuration
	Prometheus struct {
		Enabled           bool
		prometheus.Config `mapstructure:",squash"`
	}

	// Jaeger configuration
	Jaeger struct {
		Enabled       bool
		jaeger.Config `mapstructure:",squash"`
	}
}

type AppParameters struct {
	PageSize  int
	PageNum   int
	SortBy    string
	SortOrder string
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}

	if err := c.Instrumentation.Validate(); err != nil {
		return err
	}

	if err := c.Dgraph.Validate(); err != nil {
		return err
	}

	if err := c.IAM.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates the configuration.
func (c InstrumentationConfig) Validate() error {
	if c.Jaeger.Enabled {
		if err := c.Jaeger.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Configure configures some defaults in the Viper instance.
func Configure(v *viper.Viper, p *pflag.FlagSet) {
	v.AllowEmptyEnv(true)
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("./")
	// v.AddConfigPath(fmt.Sprintf("$%s_CONFIG_DIR/", strings.ToUpper(EnvPrefix)))
	p.Init("application-service", pflag.ExitOnError)
	pflag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", "application-service")
		pflag.PrintDefaults()
	}
	_ = v.BindPFlags(p)

	// v.SetEnvPrefix(EnvPrefix)
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// v.AutomaticEnv()

	// Application constants
	v.Set("serviceName", "application-service")

	// Global configuration
	v.SetDefault("environment", "production")
	v.SetDefault("debug", false)
	v.SetDefault("shutdownTimeout", 15*time.Second)

	// Log configuration
	v.SetDefault("log.LogLevel", -1)
	v.SetDefault("log.LogTimeFormat", "2006-01-02T15:04:05.999999999Z07:00")

	// Instrumentation configuration
	p.String("instrumentation.addr", ":8092", "Instrumentation HTTP server address")
	v.SetDefault("instrumentation.addr", ":8092")

	v.SetDefault("instrumentation.prometheus.enabled", false)
	v.SetDefault("instrumentation.jaeger.enabled", false)
	v.SetDefault("instrumentation.jaeger.endpoint", "http://localhost:14268")
	v.SetDefault("instrumentation.jaeger.agentEndpoint", "localhost:6831")
	v.RegisterAlias("instrumentation.jaeger.serviceName", "serviceName")
	_ = v.BindEnv("instrumentation.jaeger.username")
	_ = v.BindEnv("instrumentation.jaeger.password")

	// Server configuration
	p.String("grpcport", ":8090", "App HTTP server address")
	v.SetDefault("httpport", ":8091")

	// Dgraph configuration
	_ = v.BindEnv("dgraph.host")

	// Database Password configuration
	_ = v.BindEnv("database.pass", "DB_PASSWORD")
	_ = v.BindEnv("application.usernameadmin", "APP_ADMIN_USERNAME")
	_ = v.BindEnv("application.passwordadmin", "APP_ADMIN_PASSWORD")
	_ = v.BindEnv("application.usernamesuperadmin", "APP_SUPER_ADMIN_USERNAME")
	_ = v.BindEnv("application.passwordsuperadmin", "APP_SUPER_ADMIN_PASSWORD")
	_ = v.BindEnv("application.usernameuser", "APP_USER_USERNAME")
	_ = v.BindEnv("application.passworduser", "APP_USER_PASSWORD")

}
