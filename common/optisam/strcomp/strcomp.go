package strcomp

import (
	"strconv"
	"strings"
)

// CompareStrings compares two strings case insensitively
func CompareStrings(str1, str2 string) bool {
	if strings.ToLower(str1) != strings.ToLower(str2) {
		return false
	}
	return true
}

func StringToNum(strVar string) int {
	intVar, err := strconv.Atoi(strVar)
	if err == nil {
		return intVar
	}
	return 0
}
