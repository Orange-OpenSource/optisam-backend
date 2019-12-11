// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// Role ...
type Role uint8

const (
	// RoleAdmin is Admin Role
	RoleAdmin Role = 1
	// RoleUser is User Role
	RoleUser Role = 2
	// RoleSuperAdmin is Super admin Role
	RoleSuperAdmin Role = 3
)

// AccountInfo ...
type AccountInfo struct {
	UserId    string
	FirstName string
	LastName  string
	Locale    string
	Role      Role
	Group     []int64
}

// UpdateAccount ...
type UpdateAccount struct {
	Locale string
}
