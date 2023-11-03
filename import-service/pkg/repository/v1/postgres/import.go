package postgres

import (
	"context"
	"database/sql"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"

	//v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/product-service/pkg/api/v1"
	"go.uber.org/zap"
)

var repoObj *ImportRepository

// ImportRepository is struct for service to repo
type ImportRepository struct {
	db *sql.DB
	*gendb.Queries
}

// ImportRepository give repo object
func GetImportRepository() (obj *ImportRepository) {
	return repoObj
}

// SetImportRepository creates new Repository
func SetImportRepository(db *sql.DB) {
	if repoObj == nil {
		repoObj = &ImportRepository{
			db:      db,
			Queries: gendb.New(db)}
	}
}

// ImportRepositoryTx ...
type ImportRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

// NewImportRepositoryTx ...
func NewImportRepositoryTx(db *sql.Tx) *ImportRepositoryTx {
	return &ImportRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

func (i *ImportRepository) InsertNominativeUserRequestTx(ctx context.Context, nomUsersReq db.InsertNominativeUserRequestParams, nomUserDetails db.InsertNominativeUserRequestDetailsParams) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	iTx := NewImportRepositoryTx(tx)
	uploadId, err := iTx.InsertNominativeUserRequest(ctx, nomUsersReq)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to InsertNominativeUserRequest", zap.Error(err))
		return err
	}
	nomUserDetails.RequestID = sql.NullInt32{Int32: uploadId, Valid: true}
	err = iTx.InsertNominativeUserRequestDetails(ctx, nomUserDetails)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to InsertNominativeUserRequestDetails", zap.Error(err))
		return err
	}
	return nil
}

func (i *ImportRepository) UpdateNominativeUserRequestAnalysisTx(ctx context.Context, nomUsersReq db.UpdateNominativeUserRequestAnalysisParams, nomUserDetails db.UpdateNominativeUserDetailsRequestAnalysisParams) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	iTx := NewImportRepositoryTx(tx)
	uploadId, err := iTx.UpdateNominativeUserRequestAnalysis(ctx, nomUsersReq)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to UpdateNominativeUserRequestAnalysis", zap.Error(err))
		return err
	}
	nomUserDetails.RequestID = sql.NullInt32{Int32: uploadId, Valid: true}
	err = iTx.UpdateNominativeUserDetailsRequestAnalysis(ctx, nomUserDetails)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to UpdateNominativeUserRequestAnalysis", zap.Error(err))
		return err
	}
	return nil
}
