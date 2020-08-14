// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

var adminRpcMap = make(map[string]struct{})

//AdminRightsRequiredFunc returns true for the functions that require admin rights
func AdminRightsRequired(fullMethod string) bool {
	_, ok := adminRpcMap[fullMethod]
	return ok
}
