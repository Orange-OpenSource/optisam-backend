package v1

import (
	"context"
	"database/sql"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1 Import

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Import interface {
	gendb.Querier
	InsertNominativeUserRequestTx(ctx context.Context, nomUsersReq db.InsertNominativeUserRequestParams, nomUserDetails db.InsertNominativeUserRequestDetailsParams) error
	UpdateNominativeUserRequestAnalysisTx(ctx context.Context, nomUsersReq db.UpdateNominativeUserRequestAnalysisParams, nomUserDetails db.UpdateNominativeUserDetailsRequestAnalysisParams) error
	// StoreCoreFactorReferences store the corefactor in dB in batch format
	//StoreCoreFactorReferences(context.Context, map[string]map[string]string) error
}
