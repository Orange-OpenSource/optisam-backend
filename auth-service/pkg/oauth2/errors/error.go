// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package errors

import (
	"errors"
	"net/http"

	oauth2Errors "gopkg.in/oauth2.v3/errors"
)

const (
	// ErrInvalidCredentails when username is wrong or userdoes not exist
	ErrInvalidCredentails = 1

	// ErrCodeAccountBlocked is when account is blocked
	ErrCodeAccountBlocked = 2

	// ErrCodeLoginBlockedAccount user is trying to login into an already blocked account
	ErrCodeLoginBlockedAccount = 3
)

var (
	// ErrInvalidCredentials when username is wrong or user does not exist
	ErrInvalidCredentials = &Error{
		Response: &oauth2Errors.Response{
			Error:       errors.New("unauthorised"),
			ErrorCode:   ErrInvalidCredentails,
			Description: "Invalid username or password",
			StatusCode:  http.StatusUnauthorized,
		},
	}

	// ErrAccountBlocked when too many invalid attempts
	ErrAccountBlocked = &Error{
		Response: &oauth2Errors.Response{
			Error:       errors.New("blocked"),
			ErrorCode:   ErrCodeAccountBlocked,
			Description: "Your account has been locked because of too many invalid login attempts, please contact the administrator.",
			StatusCode:  http.StatusUnauthorized,
		},
	}

	// ErrLoginBlockedAccount when too many invalid attempts
	ErrLoginBlockedAccount = &Error{
		Response: &oauth2Errors.Response{
			Error:       errors.New("login_blocked"),
			ErrorCode:   ErrCodeLoginBlockedAccount,
			Description: "Your account is currently blocked, please contact the administrator.",
			StatusCode:  http.StatusUnauthorized,
		},
	}
)

// Error for optisam oauth2 custiom errors
type Error struct {
	Response *oauth2Errors.Response
}

// Error implements error interface Error function.
func (err *Error) Error() string {
	return err.Response.Error.Error()
}
