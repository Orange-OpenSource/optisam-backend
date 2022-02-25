package access

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token"
	"optisam-backend/common/optisam/token/claims"

	"go.uber.org/zap"

	"gopkg.in/oauth2.v3"
)

// ClaimsFetcher fetches user's claims
type ClaimsFetcher interface {
	// UserClaims gets users claims with given id
	UserClaims(context context.Context, userID string) (*claims.Claims, error)
}

//go:generate mockgen -destination=mock/mock.go -package=mock gopkg.in/oauth2.v3 AccessGenerate
type generator struct {
	claimsFetcher ClaimsFetcher
	gen           token.Generator
}

// NewGenerator returns a custom implementation of oauth2.AccessGenerate
func NewGenerator(gen token.Generator, claimsFetcher ClaimsFetcher) oauth2.AccessGenerate {
	return &generator{
		claimsFetcher: claimsFetcher,
		gen:           gen,
	}
}

func (g *generator) Token(data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	claims, err := g.claimsFetcher.UserClaims(context.Background(), data.UserID)
	if err != nil {
		logger.Log.Error("oauth2/generators/access - Token", zap.Error(err))
		return "", "", errors.New("cannot fetch claims for user")
	}

	access, err := g.gen.GenerateAccessToken(claims)
	if err != nil {
		return "", "", err
	}

	if !isGenRefresh {
		return access, "", nil
	}

	refresh, err := g.gen.GenerateRefreshToken(claims)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}
