package helper

import (
	"fmt"
	"unicode"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidatePassword(s string) (bool, error) {
	var number, upper, lower, special bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case specialCharacter(c):
			special = true
		}
	}
	if !number {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one number")
	}
	if !upper {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one upper case letter")
	}
	if !lower {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one lower case letter")
	}
	if !special {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one special character(./@/#/$/&/*/_/,)")
	}
	return true, nil
}

func specialCharacter(c rune) bool {
	s := fmt.Sprintf("%c", c)
	specialList := []string{".", "@", "#", "$", "&", "*", "_", ","}
	for _, a := range specialList {
		if a == s {
			return true
		}
	}
	return false
}
