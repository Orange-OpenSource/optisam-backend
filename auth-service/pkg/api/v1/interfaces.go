// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/auth-service/pkg/api/v1 AuthService

// AuthService is collection of all the methods required by AuthService
type AuthService interface {
	// Login will return LoginResponse. Error if it is not able to fetch user,
	// user does not exist or if user is blocked after three unsuccessful atemps.
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}
