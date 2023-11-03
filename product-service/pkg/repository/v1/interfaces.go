package v1

import (
	"context"
	"database/sql"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1 Product
//go:generate mockgen -destination=queuemock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue  Workerqueue

// DBTX to satisfy SQL DB and TX interface
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

//for dgraph connection

// //-destination=dgmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1 Product
// type DgraphConn interface {
// 	ListMetrices(ctx context.Context, scopes string) error
// }

// Product interface
type Product interface {
	gendb.Querier
	// Need to add these for transaction support
	// UpsertProductTx handles upsert product request
	UpsertProductTx(ctx context.Context, req *v1.UpsertProductRequest, user string) error

	// DropProductDataTx handles drop product data
	DropProductDataTx(ctx context.Context, scope string, deletionType v1.DropProductDataRequestDeletionTypes) error

	// UpsertNominativeUserTx upserts nominative user data
	UpsertNominativeUsersTx(ctx context.Context, req *v1.UpserNominativeUserRequest) error

	// UpsertConcurrentUserTx upserts nominative user data
	UpsertConcurrentUserTx(ctx context.Context, req *v1.ProductConcurrentUserRequest, createdBy string) error
}
