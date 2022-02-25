package helper

import (
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// SnakeCase is for converting CamelCase string to SnakeCase
func ToSnakeCase(str string) string {
	snakeCase := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snakeCase = matchAllCap.ReplaceAllString(snakeCase, "${1}_${2}")
	return strings.ToLower(snakeCase)
}
