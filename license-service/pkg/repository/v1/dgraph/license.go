// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"github.com/dgraph-io/dgo"
)

//LicenseRepository for Dgraph
type LicenseRepository struct {
	dg *dgo.Dgraph
}

//NewLicenseRepository creates new Repository
func NewLicenseRepository(dg *dgo.Dgraph) *LicenseRepository {
	return &LicenseRepository{
		dg: dg,
	}
}
