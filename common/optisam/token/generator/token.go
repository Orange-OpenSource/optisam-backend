// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package generator

import (
	"crypto/rsa"
	"io/ioutil"
	"optisam-backend/common/optisam/token"
	"optisam-backend/common/optisam/token/claims"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type tokenGenerator struct {
	signKey   *rsa.PrivateKey
	accTokDur time.Duration
	refTokDur time.Duration
}

type customClaims struct {
	OptisamClaims *claims.Claims
	jwt.StandardClaims
}

// NewTokenGenerator returns an implementation of token.Generator
func NewTokenGenerator(privateKeyPath string) (token.Generator, error) {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, err
	}

	return &tokenGenerator{
		signKey:   signKey,
		accTokDur: time.Duration(2 * time.Hour),
	}, nil

	// verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	// if err != nil {
	// 	return err
	// }

	// verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	// if err != nil {
	// 	return err
	// }
}

// GenerateAccessToken implements token.Generator GenerateAccessToken function.
func (t *tokenGenerator) GenerateAccessToken(osClaims *claims.Claims) (string, error) {
	return t.generateToken("Access Token", t.accTokDur, osClaims)
}

// GenerateRefreshToken implements token.Generator GenerateRefreshToken function.
func (t *tokenGenerator) GenerateRefreshToken(osClaims *claims.Claims) (string, error) {
	return t.generateToken("Refresh Token", t.refTokDur, osClaims)
}

func (t *tokenGenerator) generateToken(sub string, expDur time.Duration, osClaims *claims.Claims) (string, error) {
	tNow := time.Now().UTC()

	osClaims.StandardClaims = jwt.StandardClaims{
		ExpiresAt: tNow.Add(expDur).Unix(),
		IssuedAt:  tNow.Unix(),
		Issuer:    "Orange",
		Subject:   sub,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, osClaims)
	tokenStr, err := token.SignedString(t.signKey)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}
