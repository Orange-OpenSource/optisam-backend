// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
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
