// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package postgres

import (
	"database/sql"
	//Postgres pq driver
	_ "github.com/vijay1811/pq"
	"github.com/opencensus-integrations/ocsql"
	"github.com/pkg/errors"
)

// NewConnection returns a new database connection for the application.
func NewConnection(config Config) (*sql.DB, error) {
	driverName, err := ocsql.Register(
		"postgres",
		ocsql.WithAllTraceOptions(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register ocsql driver")
	}

	db, err := sql.Open(driverName, config.DSN())

	return db, errors.WithStack(err)
}
