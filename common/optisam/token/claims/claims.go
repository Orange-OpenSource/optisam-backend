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

func ReturnRole(role string) (Role, bool) {
	const noRole Role = ""
	switch role {
	case string(RoleSuperAdmin):
		return RoleSuperAdmin, true
	case string(RoleAdmin):
		return RoleAdmin, true
	case string(RoleUser):
		return RoleUser, true
	default:
		return noRole, false
	}
}
