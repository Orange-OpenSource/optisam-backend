// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package token

import "optisam-backend/common/optisam/token/claims"

//go:generate mockgen -destination=mock/mock_generator.go -package=mock optisam-backend/common/optisam/token Generator

// Generator has functionality to generate access token
type Generator interface {
	// GenerateAccessToken generates a access token
	GenerateAccessToken(osClaims *claims.Claims) (string, error)
	// GenerateRefreshToken generates a refresh token
	GenerateRefreshToken(osClaims *claims.Claims) (string, error)
}
