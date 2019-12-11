// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// LoginRequest represents all the fields required for login.
type LoginRequest struct {
	Username string
	Password string
}

// LoginResponse is the response required for LoginRequest
type LoginResponse struct {
	UserID string
	Entity string
	Locale string
}
