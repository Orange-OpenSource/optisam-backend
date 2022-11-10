// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
)

type Querier interface {
	DeleteReportsByScope(ctx context.Context, scope string) error
	DownloadReport(ctx context.Context, arg DownloadReportParams) (DownloadReportRow, error)
	GetReport(ctx context.Context, arg GetReportParams) ([]GetReportRow, error)
	GetReportType(ctx context.Context, reportTypeID int32) (ReportType, error)
	GetReportTypes(ctx context.Context) ([]ReportType, error)
	InsertReportData(ctx context.Context, arg InsertReportDataParams) error
	SubmitReport(ctx context.Context, arg SubmitReportParams) (int32, error)
	UpdateReportStatus(ctx context.Context, arg UpdateReportStatusParams) error
}

var _ Querier = (*Queries)(nil)
