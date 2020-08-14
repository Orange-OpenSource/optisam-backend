// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

// Editor represent product editor
type Editor struct {
	ID   string
	Name string
}

// EditorQueryParams params required to query editors
type EditorQueryParams struct {
	// Fields will be added in future
}
