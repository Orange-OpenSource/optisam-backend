package v1

// Group represnts a group
type Group struct {
	ID                 int64
	Name               string
	ParentID           int64
	FullyQualifiedName string
	Scopes             []string
	NumberOfGroups     int32
	NumberOfUsers      int32
	GroupCompliance    bool
}

// GroupUpdate contains updatable fields of group
type GroupUpdate struct {
	Name            string
	Scopes          []string
	GroupCompliance bool
}

// GroupQueryParams returns query params for groups
type GroupQueryParams struct {
	// No fields now as we are not supporting queries
}

//GetComplienceGroups contains groups which are compliences
type GetComplienceGroups struct {
	ID        int32    `json:"id"`
	Name      string   `json:"name"`
	ScopeCode []string `json:"scope_code"`
	ScopeName []string `json:"scope_name"`
}
