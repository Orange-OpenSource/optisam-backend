// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import "context"

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/auth-service/pkg/repository/v1 Repository
//TODO fix this reflect files should be removed automatically go:generate rm -r gomock_reflect_*

// Repository interface has all the methods we need to operate on database
type Repository interface {
	// UserInfo return a Users information based on users id
	UserInfo(ctx context.Context, userID string) (*UserInfo, error)

	// IncreaseFailedLoginCount increases the count of failed login attempts
	// We should only call this function when user's password do not match with what
	// have stored in database.
	// Note: Don't call this function if user is already blocked, i.e. unsuccessful
	// attemps are already three(for now attempts limit is three this may change).
	IncreaseFailedLoginCount(ctx context.Context, userID string) error

	// ResetLoginCount reset Failed login counts in case of successful login.
	// Note: Don't call this fucntion if user is already blocked even if he provides
	// correct credentials this time.
	ResetLoginCount(ctx context.Context, userID string) error

	// UserOwnedGroupsDirect return the groups directly owned by user
	UserOwnedGroupsDirect(ctx context.Context, userID string) ([]*Group, error)

	// // CheckPassword check for users password in database
	// CheckPassword(ctx context.Context, userID, password string) (bool, error)
}
