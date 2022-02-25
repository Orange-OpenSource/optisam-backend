package postgres

import (
	"database/sql"

	// pq driver
	_ "github.com/lib/pq"
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

	// Max connections for DB to run multiple queries parellel/concurrent
	db.SetMaxOpenConns(1000)
	// max idle connections in pool time
	db.SetMaxIdleConns(0)
	return db, errors.WithStack(err)
}
