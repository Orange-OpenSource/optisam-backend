package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/logger"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	insertScope       = `INSERT INTO scopes (scope_code,scope_name,created_by,scope_type) VALUES ($1,$2,$3,$4)`
	updateScopeInRoot = `UPDATE groups SET scopes = array_append(scopes, $1) WHERE fully_qualified_name = 'ROOT'`
	getScope          = `SELECT scope_code,scope_name,created_by,created_on,scope_type from scopes WHERE scope_code = $1`
	getGroupNames     = `Select ARRAY_AGG(name) from groups where $1 = Any (scopes);`
)

// CreateScope implements Account Service CreateScope function
func (r *AccountRepository) CreateScope(ctx context.Context, scopeName, scopeCode, userID, scopeType string) (retErr error) {

	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil && txn == nil {
			logger.Log.Error(" CreateScope - failed to initiate txn", zap.String("reason", retErr.Error()))
			retErr = fmt.Errorf("createScope : Unable to initiate transaction")
			return
		} else if retErr != nil && txn != nil {
			logger.Log.Error("CreateScope - Failed to complete transaction", zap.String("Reason", retErr.Error()))
			if error := txn.Rollback(); error != nil {
				logger.Log.Error(" CreateScope - failed to discard txn", zap.String("reason", error.Error()))
				retErr = fmt.Errorf(" CreateScope - cannot discard txn")
			}
			return
		}
		if error := txn.Commit(); error != nil {
			logger.Log.Error(" CreateScope - failed to commit txn", zap.String("reason", error.Error()))
			retErr = fmt.Errorf(" CreateScope - cannot commit txn")
		}
	}()

	_, err = txn.ExecContext(ctx, insertScope, scopeCode, scopeName, userID, scopeType)
	if err != nil {
		return err
	}

	_, err = txn.ExecContext(ctx, updateScopeInRoot, scopeCode)

	if err != nil {
		return err
	}

	return nil

}

// ListScopes implements Account Service ListScopes function
func (r *AccountRepository) ListScopes(ctx context.Context, scopeCodes []string) ([]*v1.Scope, error) {
	logger.Log.Info("repo.List Scopes", zap.Any("list scopes postgres called", time.Now()))

	var scopesDetails []*v1.Scope // nolint: prealloc
	logger.Log.Info("repo.List Scopes", zap.Any("number of scopes (number of loop itrations)", len(scopeCodes)))
	for _, scopeCode := range scopeCodes {
		var scopeDetails v1.Scope
		// Find scope details
		logger.Log.Info("repo.List Scopes", zap.Any("before get scope query for scope: "+scopeCode, time.Now()))
		err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scopeDetails.ScopeCode, &scopeDetails.ScopeName, &scopeDetails.CreatedBy, &scopeDetails.CreatedOn, &scopeDetails.ScopeType)
		logger.Log.Info("repo.List Scopes", zap.Any("after get scope query for scope: "+scopeCode, time.Now()))

		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
			return nil, fmt.Errorf("listScopes - cannot fetch scopes")
		}

		// Fetch group array
		logger.Log.Info("repo.List Scopes", zap.Any("before fetch group array query for scope: "+scopeCode, time.Now()))
		err = r.db.QueryRowContext(ctx, getGroupNames, scopeCode).Scan(pq.Array(&scopeDetails.GroupNames))
		logger.Log.Info("repo.List Scopes", zap.Any("after fetch group array query for scope: "+scopeCode, time.Now()))
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch groups", zap.String("Reason", err.Error()))
				return nil, fmt.Errorf("listScopes - cannot fetch scopes")
			}
		}

		scopesDetails = append(scopesDetails, &scopeDetails)
	}
	logger.Log.Info("repo.List Scopes", zap.Any("list scopes postgres end", time.Now()))

	return scopesDetails, nil

}

// ScopeByCode implements Account Service ScopeByCode function
func (r *AccountRepository) ScopeByCode(ctx context.Context, scopeCode string) (*v1.Scope, error) {
	var scope v1.Scope
	err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scope.ScopeCode, &scope.ScopeName, &scope.CreatedBy, &scope.CreatedOn, &scope.ScopeType)
	switch {
	case err == sql.ErrNoRows:
		return nil, v1.ErrNoData
	case err != nil:
		logger.Log.Error("Repo/Postgres - ScopeByCode - Cannot fetch scope", zap.String("Reason", err.Error()))
		return nil, fmt.Errorf("scopeByCode - Cannot fetch scope")
	default:
		return &scope, nil
	}

}
