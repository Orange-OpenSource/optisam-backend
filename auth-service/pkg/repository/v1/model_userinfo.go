// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

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

// UserInfo gives information about user
type UserInfo struct {
	UserID       string
	Role         Role
	Locale       string
	Password     string
	FailedLogins uint8
}
