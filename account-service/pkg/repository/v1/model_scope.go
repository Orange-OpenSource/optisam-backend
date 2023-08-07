package v1

import (
	"database/sql"
	"time"
)

// Scope represents scope details
type Scope struct {
	ScopeCode  string
	ScopeName  string
	CreatedBy  string
	ScopeType  string
	CreatedOn  time.Time
	GroupNames []string
	Expenses   sql.NullFloat64
}
