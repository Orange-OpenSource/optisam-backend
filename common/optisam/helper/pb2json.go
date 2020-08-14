// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package helper

import (
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

//SnakeCase is for converting CamelCase string to SnakeCase
func ToSnakeCase(str string) string {
	snakeCase := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snakeCase = matchAllCap.ReplaceAllString(snakeCase, "${1}_${2}")
	return strings.ToLower(snakeCase)
}
