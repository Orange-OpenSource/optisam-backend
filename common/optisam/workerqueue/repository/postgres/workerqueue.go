// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	gendb "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
)

type Repository struct {
	*gendb.Queries
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Queries: gendb.New(db),
		db:      db,
	}
}
