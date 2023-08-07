package postgres

import (
	"fmt"

	"github.com/pkg/errors"
)

// Config holds information necessary for connecting to a database.
type Config struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}
type DBConfig struct {
	Host      string
	Port      int
	Admin     DBuserInfo
	User      DBuserInfo
	Migration Migration
}

type Migration struct {
	Version       string
	Direction     string
	MigrationPath string
}
type DBuserInfo struct {
	User string
	Pass string
	Name string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("database host is required")
	}

	if c.Port == 0 {
		return errors.New("database port is required")
	}

	if c.User == "" {
		return errors.New("database user is required")
	}

	if c.Name == "" {
		return errors.New("database name is required")
	}

	return nil
}

func (c DBConfig) Validate() error {
	if c.Host == "" {
		return errors.New("database host is required")
	}

	if c.Port == 0 {
		return errors.New("database port is required")
	}

	if c.Admin.User == "" {
		return errors.New("database admin user is required")
	}
	if c.User.User == "" {
		return errors.New("database user is required")
	}

	if c.Admin.Name == "" {
		return errors.New("database name is required")
	}
	if c.User.Name == "" {
		return errors.New("database name is required")
	}

	return nil
}

// DSN returns a Postgres driver compatible data source name.
func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Pass,
		c.Name,
	)
}
