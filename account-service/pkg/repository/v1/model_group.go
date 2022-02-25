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
}

// GroupUpdate contains updatable fields of group
type GroupUpdate struct {
	Name string
}

// GroupQueryParams returns query params for groups
type GroupQueryParams struct {
	// No fields now as we are not supporting queries
}
