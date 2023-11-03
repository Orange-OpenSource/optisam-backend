package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	insertScope         = `INSERT INTO scopes (scope_code,scope_name,created_by,scope_type) VALUES ($1,$2,$3,$4)`
	updateScopeInRoot   = `UPDATE groups SET scopes = array_append(scopes, $1) WHERE fully_qualified_name = 'ROOT'`
	getScope            = `SELECT scope_code,scope_name,created_by,created_on,scope_type from scopes WHERE scope_code = $1`
	getGroupNames       = `Select ARRAY_AGG(name) from groups where $1 = Any (scopes);`
	upsertScopeExpenses = `
							INSERT INTO scopes_expenditure (scope_code,expenses,expenses_year,created_on,created_by)
							VALUES ($1,$2,$3,$4,$5)
							ON CONFLICT (scope_code,expenses_year)
							DO
							UPDATE SET expenses=$2,updated_on=$6, updated_by=$7;
						`
	getScopeExpenses = `SELECT expenses from scopes_expenditure WHERE scope_code = $1 and expenses_year =$2`
	getScopes        = `SELECT s.scope_code,scope_name,s.created_by,s.created_on,
						scope_type, ARRAY_AGG(g.name),se.expenses
						from scopes s
						left outer join scopes_expenditure se  
						on s.scope_code=se.scope_code and se.expenses_year = $1
						full outer Join groups g on s.scope_code = any(g.scopes)
						WHERE s.scope_code =ANY($2) 
						GROUP BY s.scope_code,se.expenses`
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
	var scopesDetails []*v1.Scope // nolint: prealloc
	rows, err := r.db.Query(getScopes, time.Now().Year()-1, pq.Array(scopeCodes))
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Error("Repo/Postgres - ListScopes - no scopes exists", zap.String("Reason", err.Error()))
			return scopesDetails, nil
		}
		logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
		return nil, fmt.Errorf("listScopes - cannot fetch scopes")
	}
	defer rows.Close()
	for rows.Next() {
		sD := v1.Scope{}
		err := rows.Scan(&sD.ScopeCode, &sD.ScopeName, &sD.CreatedBy, &sD.CreatedOn, &sD.ScopeType, pq.Array(&sD.GroupNames), &sD.Expenses)
		if err != nil {
			logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
			return nil, fmt.Errorf("listScopes - cannot fetch scopes")
		}
		scopesDetails = append(scopesDetails, &sD)
	}
	return scopesDetails, nil

	// logger.Log.Info("repo.List Scopes", zap.Any("list scopes postgres called", time.Now()))

	// var scopesDetails []*v1.Scope // nolint: prealloc
	// logger.Log.Info("repo.List Scopes", zap.Any("number of scopes (number of loop itrations)", len(scopeCodes)))
	// for _, scopeCode := range scopeCodes {
	// 	var scopeDetails v1.Scope
	// 	// Find scope details
	// 	logger.Log.Info("repo.List Scopes", zap.Any("before get scope query for scope: "+scopeCode, time.Now()))
	// 	err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scopeDetails.ScopeCode, &scopeDetails.ScopeName, &scopeDetails.CreatedBy, &scopeDetails.CreatedOn, &scopeDetails.ScopeType)
	// 	logger.Log.Info("repo.List Scopes", zap.Any("after get scope query for scope: "+scopeCode, time.Now()))

	// 	if err != nil {
	// 		logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
	// 		return nil, fmt.Errorf("listScopes - cannot fetch scopes")
	// 	}

	// 	// Fetch group array
	// 	logger.Log.Info("repo.List Scopes", zap.Any("before fetch group array query for scope: "+scopeCode, time.Now()))
	// 	err = r.db.QueryRowContext(ctx, getGroupNames, scopeCode).Scan(pq.Array(&scopeDetails.GroupNames))
	// 	logger.Log.Info("repo.List Scopes", zap.Any("after fetch group array query for scope: "+scopeCode, time.Now()))
	// 	if err != nil {
	// 		if err != sql.ErrNoRows {
	// 			logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch groups", zap.String("Reason", err.Error()))
	// 			return nil, fmt.Errorf("listScopes - cannot fetch scopes")
	// 		}
	// 	}

	// 	scopesDetails = append(scopesDetails, &scopeDetails)
	// }
	// logger.Log.Info("repo.List Scopes", zap.Any("list scopes postgres end", time.Now()))

	// return scopesDetails, nil
	// for _, scopeCode := range scopeCodes {
	// 	var scopeDetails v1.Scope
	// 	// Find scope details
	// 	err := r.db.QueryRowContext(ctx, getScope, scopeCode).Scan(&scopeDetails.ScopeCode, &scopeDetails.ScopeName, &scopeDetails.CreatedBy, &scopeDetails.CreatedOn, &scopeDetails.ScopeType)
	// 	if err != nil {
	// 		if err == sql.ErrNoRows {
	// 			continue
	// 		}
	// 		logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch scopes", zap.String("Reason", err.Error()))
	// 		return nil, fmt.Errorf("listScopes - cannot fetch scopes")
	// 	}

	// 	// Fetch group array
	// 	err = r.db.QueryRowContext(ctx, getGroupNames, scopeCode).Scan(pq.Array(&scopeDetails.GroupNames))
	// 	if err != nil {
	// 		if err != sql.ErrNoRows {
	// 			logger.Log.Error("Repo/Postgres - ListScopes - Cannot fetch groups", zap.String("Reason", err.Error()))
	// 			return nil, fmt.Errorf("listScopes - cannot fetch scopes")
	// 		}
	// 	}

	// 	scopesDetails = append(scopesDetails, &scopeDetails)
	// }
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

// UpsertScopeExpenses allows user to update their scope expenses
func (r *AccountRepository) UpsertScopeExpenses(ctx context.Context, scope_code, created_by, updated_by string, expenses float64, expenses_year int) error {
	result, err := r.db.ExecContext(ctx, upsertScopeExpenses, scope_code, expenses, expenses_year, time.Now(), created_by, time.Now(), updated_by)
	if err != nil {
		logger.Log.Error("repo/postgres - UpsertScopeExpenses - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("repo/postgres - UpsertScopeExpenses - failed to get number of rows affected", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		return fmt.Errorf("repo/postgres - UpsertScopeExpenses - expected one row to be affected,actual affected rows: %v", n)
	}

	return nil
}

// ScopeExpensesByScopeCode implements Account Service ScopeExpensesByScopeCode function
func (r *AccountRepository) ScopeExpensesByScopeCode(ctx context.Context, scopeCode string) (expenses float64, err error) {

	err = r.db.QueryRowContext(ctx, getScopeExpenses, scopeCode, time.Now().Year()-1).Scan(&expenses)
	switch {
	case err == sql.ErrNoRows:
		return 0, v1.ErrNoData
	case err != nil:
		logger.Log.Error("Repo/Postgres - ScopeExpensesByScopeCode - Cannot fetch scope", zap.String("Reason", err.Error()))
		return 0, fmt.Errorf("ScopeExpensesByScopeCode - Cannot fetch scope")
	default:
		return expenses, nil
	}
}
func (r *AccountRepository) GenerateRandomPassword() ([]byte, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(helper.CreateRandomString()), 11)
	return b, e
}
func (r *AccountRepository) CreateToken() string {
	tkn := helper.CreateToken()
	return tkn
}
