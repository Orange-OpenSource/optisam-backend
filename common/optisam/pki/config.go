package pki

import (
	"github.com/pkg/errors"
)

// Config holds information necessary for PKI.
type Config struct {
	PublicKeyPath string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.PublicKeyPath == "" {
		return errors.New("Public Key Path is required")
	}

	return nil
}
