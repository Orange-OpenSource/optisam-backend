package v1

import (
	"context"
	"database/sql"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres"
	pcdb "optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/catalog-service/pkg/repository/v1 ProductCatalog

// ProductCatalog interface
type ProductCatalog interface {
	pcdb.Querier
	// Need to add these for transaction support
	// UpsertProductTx handles upsert product request
	InsertProductTx(ctx context.Context, req *v1.Product) (res *v1.Product, err error)
	UpdateProductTx(ctx context.Context, req *v1.Product) (err error)
	UpdateEditorTx(ctx context.Context, req *v1.Editor) (err error)
	InsertRecordsTx(ctx context.Context, req *v1.UploadRecords) (message string, err error)
	GetScope(ctx context.Context, s []string) (scope []*postgres.Scope, err error)
	GetAllScope(ctx context.Context) (scope []*postgres.Scope, err error)
}

func NullString(str string) sql.NullString {
	if str == "" {
		return sql.NullString{
			String: str,
			Valid:  false,
		}
	}
	return sql.NullString{
		String: str,
		Valid:  true,
	}
}
