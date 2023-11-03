package v1

import (
	"context"
	"database/sql"

	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1 Notification

// DBTX to satisfy SQL DB and TX interface
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Notification interface
type Notification interface {
	gendb.Querier
	// Need to add these for transaction support
}
