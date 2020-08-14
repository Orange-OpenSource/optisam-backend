// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	gendb "optisam-backend/report-service/pkg/repository/v1/postgres/db"
)

//ReportRepository
type ReportRepository struct {
	*gendb.Queries
	db *sql.DB
}

//NewReportRepository creates new Repository
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}
