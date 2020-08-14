// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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
