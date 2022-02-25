package opa

import (
	"github.com/pkg/errors"
)

// Config holds information necessary for PKI.
type Config struct {
	Rego string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Rego == "" {
		return errors.New("Rego Path is required")
	}

	return nil
}
