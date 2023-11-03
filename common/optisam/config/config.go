package config

import (
	"errors"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/dgraph"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/docker"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/jaeger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/pki"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/postgres"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/prometheus"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/redis"
)

// Config is configuration for Server
type Config struct {
	// Meaningful values are recommended (eg. production, development, staging, release/123, etc)
	Environment string

	// BadgerDir contains path of badgers
	BadgerDir string

	// Turns on some debug functionality
	Debug bool

	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string

	// Dgraph connection information
	Dgraph *dgraph.Config

	// Postgress connection information
	Postgres *postgres.Config

	// Redis connection information
	Redis *redis.Config

	// Dockers Connection information
	Dockers []docker.Config

	// Log configuration
	Log logger.Config

	// Instrumentation configuration
	Instrumentation InstrumentationConfig

	AppParams AppParameters

	// PKI configuration
	PKI pki.Config

	// Init Wait time tells time taken by docker container to setup, now other things can run
	INITWAITTIME time.Duration
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
type Application struct {
	UserNameAdmin      string
	PasswordAdmin      string
	UserNameSuperAdmin string
	PasswordSuperAdmin string
	UserNameUser       string
	PasswordUser       string
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
