// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"database/sql"
	"time"
)

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

func (role Role) RoleToRoleString() string {
	switch role {
	case RoleAdmin:
		return "Admin"
	case RoleUser:
		return "User"
	case RoleSuperAdmin:
		return "SuperAdmin"
	default:
		return "undefined"
	}
}

// AccountInfo ...
type AccountInfo struct {
	UserId          string
	FirstName       string
	LastName        string
	Locale          string
	Password        string
	Role            Role
	ProfilePic      []byte
	LastLogin       sql.NullTime
	ContFailedLogin int16
	CreatedOn       time.Time
	FirstLogin      bool
	Group           []int64
	GroupName       []string
}

// UpdateAccount ...
type UpdateAccount struct {
	FirstName  string
	LastName   string
	Locale     string
	ProfilePic []byte
}

type UpdateUserAccount struct {
	Role Role
}
type UserQueryParams struct {
}
