package v1

import "time"

// Scope represents scope details
type Scope struct {
	ScopeCode  string
	ScopeName  string
	CreatedBy  string
	ScopeType  string
	CreatedOn  time.Time
	GroupNames []string
}
