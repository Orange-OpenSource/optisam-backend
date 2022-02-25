package strcomp

import (
	"strings"
)

// CompareStrings compares two strings case insensitively
func CompareStrings(str1, str2 string) bool {
	if strings.ToLower(str1) != strings.ToLower(str2) {
		return false
	}
	return true
}
