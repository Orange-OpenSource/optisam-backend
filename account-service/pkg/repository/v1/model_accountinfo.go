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
	FirstLogin      bool
	Role            Role
	ContFailedLogin int16
	UserID          string
	FirstName       string
	LastName        string
	Locale          string
	Password        string
	ProfilePic      []byte
	LastLogin       sql.NullTime
	CreatedOn       time.Time
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
