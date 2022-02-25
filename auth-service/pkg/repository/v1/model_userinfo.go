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
