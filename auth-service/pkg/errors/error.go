// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package errors

import "errors"

var (
	errUserNotFound    = errors.New("not_found")
	errInvalidPassword = errors.New("unauthorised")
	errUserBlocked     = errors.New("forbidden")
)
