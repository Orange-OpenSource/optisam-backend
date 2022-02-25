package iam

import (
	"github.com/pkg/errors"
)

// Config holds information necessary for PKI.
type Config struct {
	PublicKeyPath string
	RegoPath      string
	APIKey        string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.PublicKeyPath == "" {
		return errors.New("Public Key Path is required")
	}

	if c.RegoPath == "" {
		return errors.New("Rego file path is required")
	}
	return nil
}
