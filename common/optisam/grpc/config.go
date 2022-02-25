package grpc

import (
	"time"

	"github.com/pkg/errors"
)

// Config holds information necessary for connecting to a database.
type Config struct {
	APIKey  string
	Address map[string]string
	Timeout time.Duration
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("grpc Access Key is required")
	}
	return nil
}
