package iam

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
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

func GetAPIKey(config Config) string {
	return config.APIKey
}
