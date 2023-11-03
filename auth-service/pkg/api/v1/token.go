package v1

import (
	"database/sql"
	"time"
)

// LoginRequest represents all the fields required for login.
type TokenRequest struct {
	Username  string
	Token     string
	TokenType string
}
type ChangePasswordRequest struct {
	Username             string `json:"user"`
	Token                string `json:"token"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
	TokenType            string `json:"tokenType"`
	Action               string `json:"action"`
}
type ForgotPasswordRequest struct {
	Username string `json:"user"`
}

type AccountInfo struct {
	FirstLogin      bool
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
