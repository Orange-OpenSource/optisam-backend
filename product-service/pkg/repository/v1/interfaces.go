package v1

import (
	"context"
	"database/sql"
	v1 "optisam-backend/product-service/pkg/api/v1"
	gendb "optisam-backend/product-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock optisam-backend/product-service/pkg/repository/v1 Product
//go:generate mockgen -destination=queuemock/mock.go -package=mock optisam-backend/common/optisam/workerqueue  Workerqueue

// DBTX to satisfy SQL DB and TX interface
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Product interface
type Product interface {
	gendb.Querier
	// Need to add these for transaction support
	// UpsertProductTx handles upsert product request
	UpsertProductTx(ctx context.Context, req *v1.UpsertProductRequest, user string) error

	// DropProductDataTx handles drop product data
	DropProductDataTx(ctx context.Context, scope string, deletionType v1.DropProductDataRequestDeletionTypes) error
}
