// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	gendb "optisam-backend/acqrights-service/pkg/repository/v1/postgres/db"
)

//AcqRightsRepository
type AcqRightsRepository struct {
	*gendb.Queries
	db *sql.DB
}

//NewAcqRightsRepository creates new Repository
func NewAcqRightsRepository(db *sql.DB) *AcqRightsRepository {
	return &AcqRightsRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}
