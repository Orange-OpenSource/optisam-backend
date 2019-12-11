// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package pki

import (
	"crypto/rsa"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"io/ioutil"
)

// GetVerifyKey is for getting the public key
func GetVerifyKey(config Config) (*rsa.PublicKey, error) {
	// Get the verify key
	verifyBytes, err := ioutil.ReadFile(config.PublicKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read publickey")
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register read publickey")
	}
	return verifyKey, nil
}
