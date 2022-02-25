package v1

var adminRPCMap = make(map[string]struct{})

// AdminRightsRequired returns true for the functions that require admin rights
func AdminRightsRequired(fullMethod string) bool {
	_, ok := adminRPCMap[fullMethod]
	return ok
}
