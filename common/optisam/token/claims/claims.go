// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package claims

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// Role definses the role of the user
type Role string

const (
	// RoleSuperAdmin as all the rights possible in system
	RoleSuperAdmin Role = "SuperAdmin"
	// RoleAdmin also has all the rights with few exceptions
	RoleAdmin Role = "Admin"
	// RoleUser has read only rightds in the system
	RoleUser Role = "User"
)

// Claims carried by optisam token
type Claims struct {
	UserID string
	Locale string
	Role   Role
	Socpes []string
	jwt.StandardClaims
}
