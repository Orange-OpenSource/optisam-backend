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
