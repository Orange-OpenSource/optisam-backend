// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
)

// SimulationServiceRepo implements all the methods defined by repository of this interface
type SimulationServiceRepo struct {
	db *sql.DB
	*db.Queries
}

// NewSimulationServiceRepository returns an implementation of repository
func NewSimulationServiceRepository(d *sql.DB) *SimulationServiceRepo {
	return &SimulationServiceRepo{
		db:      d,
		Queries: db.New(d),
	}

}
