package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/config"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/iam"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/jaeger"
	kafkaConnector "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/kafka"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/prometheus"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/redis"

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

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort          string
	Activationtimeout int
	Forgotpasstimeout int

	// Private key path dor jwt token generation.
	JWTPrivateKey string

	// Database connection information
	Database postgres.DBConfig

	// Log configuration
	Log logger.Config

	// Instrumentation configuration
	Instrumentation InstrumentationConfig
	Redis           *redis.Config
	Emailtemplate   Emailtemplate
	GrpcServers     grpc.Config
	HTTPServers     HttpConfg
	Application     config.Application
	IAM             iam.Config
	Kafka           kafkaConnector.KafkaConfig
}
type HttpConfg struct {
	Address map[string]string
}
type Emailtemplate struct {
	Passwordresetpath        string
	Activationpath           string
	Redirecbaseurl           string
	Redirectappactivation    string
	Redirectappurlforgotpass string
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

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}

	if err := c.Instrumentation.Validate(); err != nil {
		return err
	}

	if err := c.Database.Validate(); err != nil {
		return err
	}
	if err := c.Kafka.Validate(); err != nil {
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
	p.Init("auth-service", pflag.ExitOnError)
	pflag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", "auth-service")
		pflag.PrintDefaults()
	}
	_ = v.BindPFlags(p)

	// v.SetEnvPrefix(EnvPrefix)
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// v.AutomaticEnv()

	// Application constants
	v.Set("serviceName", "auth-service")

	// Global configuration
	v.SetDefault("environment", "production")
	v.SetDefault("debug", false)
	v.SetDefault("shutdownTimeout", 15*time.Second)

	// Instrumentation configuration
	p.String("instrumentation.addr", ":8092", "Instrumentation HTTP server address")
	v.SetDefault("instrumentation.addr", ":8092")

	// Log configuration
	v.SetDefault("log.LogLevel", -1)
	v.SetDefault("log.LogTimeFormat", "2006-01-02T15:04:05.999999999Z07:00")

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

	// Database configuration
	_ = v.BindEnv("database.host")
	v.SetDefault("database.port", 5432)
	_ = v.BindEnv("database.admin.name")
	_ = v.BindEnv("database.user.name")
	_ = v.BindEnv("database.admin.pass", "DB_PASSWORD")
	_ = v.BindEnv("database.user.pass", "DBUSR_PASSWORD")
	_ = v.BindEnv("database.migration.version", "MIG_VERSION")
	_ = v.BindEnv("database.migration.direction", "MIG_DIR")
	//env mapping for redis
	//_ = v.BindEnv("redis.redishost", "REDIS_HOST")
	_ = v.BindEnv("redis.redispassword", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("redis.username", "REDIS_USERNAME")
	_ = v.BindEnv("redis.sentinelhost", "REDIS_SENTINELHOST")
	_ = v.BindEnv("redis.sentinelport", "REDIS_SENTINELPORT")
	_ = v.BindEnv("redis.sentinelmastername", "REDIS_SENTINELMASTERNAME")

	_ = v.BindEnv("kafka.bootstrapservers", "KAFKA_BOOTSTRAPSERVER")
	_ = v.BindEnv("kafka.securityprotocol", "KAFKA_SECURITYPROTOCOL")
	_ = v.BindEnv("kafka.sslkeylocation", "KAFKA_SSLKEYLOCATION")
	_ = v.BindEnv("kafka.sslcertificatelocation", "KAFKA_SSLCERTIFICATELOCATION")
	_ = v.BindEnv("kafka.sslcalocation", "KAFKA_SSLCALOCATION")

	_ = v.BindEnv("application.usernamesuperadmin", "APP_SUPER_ADMIN_USERNAME")
	_ = v.BindEnv("application.passwordsuperadmin", "APP_SUPER_ADMIN_PASSWORD")
}
