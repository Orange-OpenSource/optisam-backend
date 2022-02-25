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

// BuildSQL is sql builder for dynamic queries
// FilterClause clause currently support string values
// There can be extra speces if  one of the argument is not provided
// TODO SQL injeciton handling and input sanitization
func BuildSQL(selectClause string, filterClause map[string]FilterValue, groupClause string, sortClause string) string {
	var whereClause string
	for col, val := range filterClause {
		whereClause += col + string(val.equality) + val.value + " AND "
	}
	if wlen := len(whereClause); wlen > 0 {
		// prepend WHERE and drop the last AND
		whereClause = "WHERE " + whereClause[:wlen-len(" AND ")]
	}
	return fmt.Sprintf("%s %s %s %s", selectClause, whereClause, groupClause, sortClause)
}
