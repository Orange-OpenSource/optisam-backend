// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
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
