package token

import "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

//go:generate mockgen -destination=mock/mock_generator.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token Generator

// Generator has functionality to generate access token
type Generator interface {
	// GenerateAccessToken generates a access token
	GenerateAccessToken(osClaims *claims.Claims) (string, error)
	// GenerateRefreshToken generates a refresh token
	GenerateRefreshToken(osClaims *claims.Claims) (string, error)
}
