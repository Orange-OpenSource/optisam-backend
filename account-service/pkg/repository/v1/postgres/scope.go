// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/logger"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	insertScope       = `INSERT INTO scopes (scope_code,scope_name,created_by) VALUES ($1,$2,$3)`
	updateScopeInRoot = `UPDATE groups SET scopes = array_append(scopes, $1) WHERE id = 1`
	getScope          = `SELECT scope_code,scope_name,created_by,created_on from scopes WHERE scope_code = $1`
	getGroupNames     = `Select ARRAY_AGG(name) from groups where $1 = Any (scopes);`
)

// CreateScope implements Account Service CreateScope function
func (r *AccountRepository) CreateScope(ctx context.Context, scopeName, scopeCode, userID string) (retErr error) {

	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil && txn == nil {
			logger.Log.Error(" CreateScope - failed to initiate txn", zap.String("reason", retErr.Error()))
			retErr = fmt.Errorf("CreateScope : Unable to initiate transaction")
			return
		} else if retErr != nil && txn != nil {
			logger.Log.Error("CreateScope - Failed to complete transaction", zap.String("Reason", retErr.Error()))
			if err := txn.Rollback(); err != nil {
				logger.Log.Error(" CreateScope - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf(" CreateScope - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(); err != nil {
			logger.Log.Error(" CreateScope - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf(" CreateScope - cannot commit txn")
		}
	}()

	_, err = txn.ExecContext(ctx, insertScope, scopeCode, scopeName, userID)
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

	var scopesDetails []*v1.Scope
	for _, scopeCode := range scopeCodes {
		var scopeDetails v1.Scope
		// Find scope details
		err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scopeDetails.ScopeCode, &scopeDetails.ScopeName, &scopeDetails.CreatedBy, &scopeDetails.CreatedOn)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
			return nil, fmt.Errorf("ListScopes - cannot fetch scopes")
		}

		//Fetch group array
		err = r.db.QueryRowContext(ctx, getGroupNames, scopeCode).Scan(pq.Array(&scopeDetails.GroupNames))
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch groups", zap.String("Reason", err.Error()))
				return nil, fmt.Errorf("ListScopes - cannot fetch scopes")
			}
		}

		scopesDetails = append(scopesDetails, &scopeDetails)
	}

	return scopesDetails, nil

}

// ScopeByCode implements Account Service ScopeByCode function
func (r *AccountRepository) ScopeByCode(ctx context.Context, scopeCode string) (*v1.Scope, error) {
	var scope v1.Scope
	err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scope.ScopeCode, &scope.ScopeName, &scope.CreatedBy, &scope.CreatedOn)
	switch {
	case err == sql.ErrNoRows:
		return nil, v1.ErrNoData
	case err != nil:
		logger.Log.Error("Repo/Postgres - ScopeByCode - Cannot fetch scope", zap.String("Reason", err.Error()))
		return nil, fmt.Errorf("ScopeByCode - Cannot fetch scope")
	default:
		return &scope, nil
	}

}
