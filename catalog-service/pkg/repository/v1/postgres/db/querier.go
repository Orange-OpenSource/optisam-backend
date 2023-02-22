// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	CreateUploadFileLog(ctx context.Context, arg CreateUploadFileLogParams) error
	DeleteEditorCatalog(ctx context.Context, id string) error
	DeleteProductCatalog(ctx context.Context, id string) error
	DeleteVersionCatalog(ctx context.Context, id string) error
	GetEditorCatalog(ctx context.Context, id string) (EditorCatalog, error)
	GetEditorCatalogByName(ctx context.Context, name string) (EditorCatalog, error)
	GetEditorCatalogName(ctx context.Context, id string) (GetEditorCatalogNameRow, error)
	GetProductCatalogByEditorId(ctx context.Context, arg GetProductCatalogByEditorIdParams) (ProductCatalog, error)
	GetProductCatalogByPrductID(ctx context.Context, id string) (ProductCatalog, error)
	GetProductCatalogBySwidTag(ctx context.Context, swidTagProduct sql.NullString) (ProductCatalog, error)
	GetProductsByEditorID(ctx context.Context, editorID string) ([]ProductCatalog, error)
	GetProductsNamesByEditorID(ctx context.Context, editorID string) ([]GetProductsNamesByEditorIDRow, error)
	GetUploadFileLogs(ctx context.Context) ([]UploadFileLog, error)
	GetVersionCatalogByPrductID(ctx context.Context, id string) ([]VersionCatalog, error)
	GetVersionCatalogBySwidTag(ctx context.Context, swidTagVersion sql.NullString) (VersionCatalog, error)
	InsertEditorCatalog(ctx context.Context, arg InsertEditorCatalogParams) error
	InsertProductCatalog(ctx context.Context, arg InsertProductCatalogParams) error
	InsertVersionCatalog(ctx context.Context, arg InsertVersionCatalogParams) error
	UpdateEditorCatalog(ctx context.Context, arg UpdateEditorCatalogParams) error
	UpdateEditorNameForProductCatalog(ctx context.Context, arg UpdateEditorNameForProductCatalogParams) error
	UpdateProductCatalog(ctx context.Context, arg UpdateProductCatalogParams) error
	UpdateProductEditor(ctx context.Context, arg UpdateProductEditorParams) error
	UpdateVersionCatalog(ctx context.Context, arg UpdateVersionCatalogParams) error
	UpdateVersionForEditor(ctx context.Context, arg UpdateVersionForEditorParams) error
	UpdateVersionsSysSwidatagsForEditor(ctx context.Context, id string) error
	UpsertEditorCatalog(ctx context.Context, arg UpsertEditorCatalogParams) (UpsertEditorCatalogRow, error)
	UpsertProductCatalog(ctx context.Context, arg UpsertProductCatalogParams) (string, error)
	UpsertVersionCatalog(ctx context.Context, arg UpsertVersionCatalogParams) (string, error)
}

var _ Querier = (*Queries)(nil)
