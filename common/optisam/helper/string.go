package helper

import (
	"math/rand"
)

const (
	specialChars   = ".@#$&*_,"
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numericChars   = "0123456789"
)

func CreateRandomString() string {
	var password string
	password += string(specialChars[rand.Intn(len(specialChars))])
	password += string(lowercaseChars[rand.Intn(len(lowercaseChars))])
	password += string(uppercaseChars[rand.Intn(len(uppercaseChars))])
	password += string(numericChars[rand.Intn(len(numericChars))])

	for i := 0; i < 8; i++ {
		charSet := []string{specialChars, lowercaseChars, uppercaseChars, numericChars}
		randomChar := charSet[rand.Intn(len(charSet))]
		password += string(randomChar[rand.Intn(len(randomChar))])
	}

	return password
}
