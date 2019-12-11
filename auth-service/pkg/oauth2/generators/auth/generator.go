// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package auth

import (
	"errors"
	"gopkg.in/oauth2.v3"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gopkg.in/oauth2.v3 AuthorizeGenerate
type generator struct{}

// NewGenerator returns a custom implementation of oauth2.AuthorizeGenerate.
func NewGenerator() oauth2.AuthorizeGenerate {
	return &generator{}
}

// Token implements oauth2.AuthorizeGenerate Token function.
func (g *generator) Token(data *oauth2.GenerateBasic) (code string, err error) {
	return "", errors.New("Not implemented")
}
