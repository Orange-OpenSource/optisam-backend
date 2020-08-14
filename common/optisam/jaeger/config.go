// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package jaeger

import "github.com/pkg/errors"

// Config holds information necessary for sending trace to Jaeger.
type Config struct {
	// Service Name
	ServiceName string

	// CollectorEndpoint is the Jaeger HTTP Thrift endpoint.
	// For example, http://localhost:14268.
	CollectorEndpoint string

	// AgentEndpoint instructs exporter to send spans to Jaeger agent at this address.
	// For example, localhost:6831.
	AgentEndpoint string

	// Username to be used if basic auth is required.
	// Optional.
	Username string

	// Password to be used if basic auth is required.
	// Optional.
	Password string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.CollectorEndpoint == "" && c.AgentEndpoint == "" {
		return errors.New("either endpoint or agent endpoint must be configured")
	}

	return nil
}
