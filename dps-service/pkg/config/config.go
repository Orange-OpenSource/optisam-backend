// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package config

import (
	"optisam-backend/common/optisam/cron"
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

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfg *Config

func SetConfig(conf Config) {
	if cfg == nil {
		cfg = &conf
	}
}

func GetConfig() (conf Config) {
	if cfg != nil {
		conf = *cfg
	}
	return
}

// Config is configuration for Server
type Config struct {
	// Meaningful values are recommended (eg. production, development, staging, release/123, etc)
	Environment string

	// Turns on some debug functionality
	Debug bool

	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	ApiKey string

	// HTTP/REST gateway start parameters section
	HTTPPort string

	//GRPC Server Configuration
	GrpcServers grpc.Config

	//Handles cron config
	Cron cron.Config

	//For interservice http calls(non grpc server)["ip:port"]
	HttpServers httpConfg

	FilesLocation string

	ArchiveLocation string

	Postgres *postgres.Config

	// Log configuration
	Log logger.Config

	//WorkerQueue holds queue config
	WorkerQueue workerqueue.QueueConfig

	//MaxFileWorker
	MaxFileWorker int

	//MaxApiWorker
	MaxApiWorker int

	//MaxDeferWorker
	MaxDeferWorker int

	// Instrumentation configuration
	Instrumentation InstrumentationConfig

	AppParams AppParameters

	//IAM Configuration
	IAM iam.Config
}

type httpConfg struct {
	Address map[string]string
}

type Server struct {
	Name string
	Host string
	Port string
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

	if err := c.Postgres.Validate(); err != nil {
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
	p.Init("dps-service", pflag.ExitOnError)
	pflag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", "license-service")
		pflag.PrintDefaults()
	}
	_ = v.BindPFlags(p)

	// v.SetEnvPrefix(EnvPrefix)
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// v.AutomaticEnv()

	// Application constants
	v.Set("serviceName", "dpsservice")

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
	_ = v.BindEnv("postgres.pass", "DB_PASSWORD")

	// PKI configuraiton
	v.SetDefault("pki.publickeypath", ".")

}
