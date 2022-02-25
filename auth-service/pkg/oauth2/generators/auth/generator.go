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
	return "", errors.New("not implemented")
}
