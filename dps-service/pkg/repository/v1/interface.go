package v1

import (
	"context"
	"database/sql"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock optisam-backend/dps-service/pkg/repository/v1 Dps
//go:generate mockgen -destination=queuemock/mock.go -package=mock optisam-backend/common/optisam/workerqueue  Workerqueue

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Dps interface {
	gendb.Querier

	// StoreCoreFactorReferences store the corefactor in dB in batch format
	StoreCoreFactorReferences(context.Context, map[string]map[string]string) error
}
