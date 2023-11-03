package helper

import (
	"encoding/base64"
	"math/rand"

	"golang.org/x/crypto/bcrypt"
)

func CreateToken() string {
	// Generate a random byte slice with 32 bytes of data
	tokenBytes, err := bcrypt.GenerateFromPassword([]byte(CreateRandomString()), 11)

	_, err = rand.Read(tokenBytes)
	if err != nil {
		panic(err)
	}

	// Encode the byte slice as a base64 string
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token
}

type EmailParams struct {
	FirstName   string
	Email       string
	RedirectUrl string
	TokenType   string
	Token       string
}
