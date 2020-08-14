// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	//gendb "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
)

var repoObj *DpsRepository

//DpsRepository is struct for service to repo
type DpsRepository struct {
	db      *sql.DB
	Queries *gendb.Queries
}

//GetDpsRepository give repo object
func GetDpsRepository() (obj DpsRepository, err error) {
	if repoObj == nil {
		//ERROR
	}
	obj = *repoObj
	return
}

//SetDpsRepository creates new Repository
func SetDpsRepository(db *sql.DB) {
	if repoObj == nil {
		repoObj = &DpsRepository{
			db:      db,
			Queries: gendb.New(db)}
	}
}
