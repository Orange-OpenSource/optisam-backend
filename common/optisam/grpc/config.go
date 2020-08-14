// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package grpc

import (
	"time"

	"github.com/pkg/errors"
)

// Config holds information necessary for connecting to a database.
type Config struct {
	ApiKey  string
	Address map[string]string
	Timeout time.Duration
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.ApiKey == "" {
		return errors.New("grpc Access Key is required")
	}
	return nil
}
