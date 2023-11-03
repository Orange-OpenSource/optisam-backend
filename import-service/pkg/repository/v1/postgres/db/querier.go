// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	GetDgraphCompletedBatches(ctx context.Context, uploadID string) (sql.NullInt32, error)
	InsertNominativeUserRequest(ctx context.Context, arg InsertNominativeUserRequestParams) (int32, error)
	InsertNominativeUserRequestDetails(ctx context.Context, arg InsertNominativeUserRequestDetailsParams) error
	ListNominativeUsersUploadedFiles(ctx context.Context, arg ListNominativeUsersUploadedFilesParams) ([]ListNominativeUsersUploadedFilesRow, error)
	UpdateNominativeUserDetailsRequestAnalysis(ctx context.Context, arg UpdateNominativeUserDetailsRequestAnalysisParams) error
	UpdateNominativeUserRequestAnalysis(ctx context.Context, arg UpdateNominativeUserRequestAnalysisParams) (int32, error)
	UpdateNominativeUserRequestDgraphBatchSuccess(ctx context.Context, uploadID string) (UpdateNominativeUserRequestDgraphBatchSuccessRow, error)
	UpdateNominativeUserRequestDgraphSuccess(ctx context.Context, arg UpdateNominativeUserRequestDgraphSuccessParams) error
	UpdateNominativeUserRequestPostgresSuccess(ctx context.Context, arg UpdateNominativeUserRequestPostgresSuccessParams) (UpdateNominativeUserRequestPostgresSuccessRow, error)
	UpdateNominativeUserRequestSuccess(ctx context.Context, arg UpdateNominativeUserRequestSuccessParams) error
}

var _ Querier = (*Queries)(nil)
