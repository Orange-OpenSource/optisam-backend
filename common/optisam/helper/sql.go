// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package helper

import "fmt"

type equalityType string

const (
	ExactEqual equalityType = "="
	LikeEqual  equalityType = "LIKE"
)

type FilterValue struct {
	equality equalityType
	value    string
}

//BuildSQL is sql builder for dynamic queries
//FilterClause clause currenlty support string values
//There can be extra speces if  one of the argument is not provided
//TODO SQL injeciton handling and input sanitization
func BuildSQL(selectClause string, FilterClause map[string]FilterValue, groupClause string, sortClause string) string {
	var whereClause string
	for col, val := range FilterClause {
		whereClause += col + string(val.equality) + val.value + " AND "
	}
	if wlen := len(whereClause); wlen > 0 {
		// prepend WHERE and drop the last AND
		whereClause = "WHERE " + whereClause[:wlen-len(" AND ")]
	}
	return fmt.Sprintf("%s %s %s %s", selectClause, whereClause, groupClause, sortClause)
}
