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
	repo "optisam-backend/account-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

type role string

const (
	roleAdmin      role = "Admin"
	roleUser       role = "User"
	roleSuperAdmin role = "SuperAdmin"
)

const (
	updateAccountQuery = "UPDATE users SET first_name = $1, last_name = $2, locale = $3, profile_pic = $4 WHERE username = $5"

	updateUserAccountQuery = "UPDATE users SET role = $1  WHERE username = $2"

	selectAccountInfo = `
	SELECT
	username,
	password,
	first_name,
	last_name,
	locale,
	role,
	profile_pic,
	cont_failed_login,
	created_on,
	first_login
	FROM users
	WHERE username = $1`

	changeUserFirstLoginQuery = "UPDATE users SET first_login = FALSE  WHERE username = $1"

	createAccountQuery = " INSERT INTO users (username,password,first_name,last_name,role,locale,first_login) VALUES($1,$2,$3,$4,$5,$6,TRUE)"

	selectAccount = `
	SELECT
	username,
	password,
	first_name,
	last_name,
	locale,
	role
	FROM users
	`
	selectAccountWithGroupInfo = `
	SELECT
	username,
	first_name,
	last_name,
	locale,
	-- all the groups owned by users
	ARRAY(
		SELECT 
		name
		from groups 
		WHERE groups.id IN(
            		SELECT group_id
            		FROM group_ownership 
            		WHERE group_ownership.user_id=username
        	)
	) as groups,
	role
	FROM users
	WHERE username<> $1;
	`
	selectAccountWithQueryParams = `
	SELECT 
	DISTINCT ON(username) username,
	first_name,
	last_name,
	locale,
	-- all the groups owned by users
	ARRAY(
		SELECT 
		name
		from groups 
		WHERE groups.id IN(
            		SELECT group_id
            		FROM group_ownership 
            		WHERE group_ownership.user_id=username
        	)
	) as groups,
	role 
	from users 
	INNER JOIN group_ownership ON users.username  = group_ownership.user_id
	-- child groups of the groups owned by the user
	WHERE group_ownership.group_id IN( 
			SELECT id
			FROM groups 
			WHERE fully_qualified_name <@ (
				SELECT ARRAY(
					SELECT fully_qualified_name 
					FROM groups
					INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
					WHERE group_ownership.user_id = $1
				)
			)
	) AND username <> $1;
	`

	selectAccountForGroup = selectAccount + `
	INNER JOIN group_ownership ON users.username  = group_ownership.user_id
	WHERE group_ownership.group_id = $1
	`
	existsUserbyID = `
	SELECT 
	count(*) AS total_records 
	FROM users 
	WHERE username=$1
	`
	existsGroupForUser = `SELECT 
	count(*) AS total_records 
	FROM groups 
	WHERE fully_qualified_name <@ 
	(SELECT ARRAY(SELECT fully_qualified_name 
	FROM groups
	INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
	WHERE group_ownership.user_id = $1
	)) AND id=$2
	`
	// checkPasswordQuery = `
	// SELECT
	// COUNT(*)
	// FROM users
	// WHERE username= $1
	// AND password = crypt($2,password)
	// `
	changePasswordQuery = "UPDATE users SET password = $2 where username =$1"

	userBelongsToAdminGroup = `
	SELECT
	count(*) as total_records
	FROM users
	INNER JOIN group_ownership ON users.username  = group_ownership.user_id
	WHERE group_ownership.group_id IN( 
		SELECT id
		FROM groups 
		WHERE fully_qualified_name <@ (
			SELECT ARRAY(
				SELECT fully_qualified_name 
				FROM groups
				INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
				WHERE group_ownership.user_id = $1
			)
		)
	) AND username = $2;
	`
)

//AccountRepository for Dgraph
type AccountRepository struct {
	*repo.Queries
	db *sql.DB
}

//NewAccountRepository creates new Repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		Queries: repo.New(db),
		db:      db,
	}
}

//UpdateAccount allows user to update their personal information
func (r *AccountRepository) UpdateAccount(ctx context.Context, userID string, req *v1.UpdateAccount) error {
	result, err := r.db.ExecContext(ctx, updateAccountQuery, req.FirstName, req.LastName, req.Locale, req.ProfilePic, userID)
	if err != nil {
		logger.Log.Error("repo/postgres - UpdateAccount - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("repo/postgres - UpdateAccount - failed to get number of rows affected", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		return fmt.Errorf("repo/postgres - UpdateAccount - expected one row to be affected,actual affected rows: %v", n)
	}

	return nil
}

//UpdateUserAccount allows admin to update the role of user
func (r *AccountRepository) UpdateUserAccount(ctx context.Context, userID string, req *v1.UpdateUserAccount) error {
	roleUser, err := dbRoleToPostGresRole(req.Role)
	if err != nil {
		logger.Log.Error("repo/postgres - UpdateUserAccount - dbRoleToPostGresRole", zap.String("reason", err.Error()))
		return err
	}
	result, err := r.db.ExecContext(ctx, updateUserAccountQuery, roleUser, userID)
	if err != nil {
		logger.Log.Error("repo/postgres - UpdateUserAccount - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("repo/postgres - UpdateUserAccount - failed to get number of rows affected", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		return fmt.Errorf("repo/postgres - UpdateUserAccount - expected one row to be affected,actual affected rows: %v", n)
	}
	return nil
}

// AccountInfo implements v1.Account's AccountInfo function.
func (r *AccountRepository) AccountInfo(ctx context.Context, userID string) (*v1.AccountInfo, error) {
	ai := &v1.AccountInfo{}
	var roleUser role
	err := r.db.QueryRowContext(ctx, selectAccountInfo, userID).
		Scan(&ai.UserId, &ai.Password, &ai.FirstName, &ai.LastName, &ai.Locale, &roleUser, &ai.ProfilePic, &ai.ContFailedLogin, &ai.CreatedOn, &ai.FirstLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, v1.ErrNoData
		}
		return nil, err
	}
	roleUserDb, err := postgresRoleToDbRole(roleUser)
	if err != nil {
		return nil, err
	}
	ai.Role = roleUserDb
	return ai, nil
}

//ChangeUserFirstLogin implements Account ChangeUserFirstLogin function
func (r *AccountRepository) ChangeUserFirstLogin(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, changeUserFirstLoginQuery, userID)
	if err != nil {
		logger.Log.Error("repo/postgres - ChangeUserFirstLogin - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error("repo/postgres - ChangeUserFirstLogin - failed to get number of rows affected", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		return fmt.Errorf("repo/postgres - ChangeUserFirstLogin - expected one row to be affected,actual affected rows: %v", n)
	}

	return nil
}

// CreateAccount implements Account CreateAccount function
func (r *AccountRepository) CreateAccount(ctx context.Context, acc *v1.AccountInfo) (retErr error) {
	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {

	}
	defer func() {
		if retErr != nil {
			if err := txn.Rollback(); err != nil {
				logger.Log.Error(" CreateAccount - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf(" CreateAccount - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(); err != nil {
			logger.Log.Error(" CreateAccount - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf(" CreateAccount - cannot commit txn")
		}
	}()
	roleUser, err := dbRoleToPostGresRole(acc.Role)
	if err != nil {
		return err
	}
	result, err := txn.ExecContext(ctx, createAccountQuery, acc.UserId, acc.Password, acc.FirstName, acc.LastName, roleUser, acc.Locale)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("repo/postgres - CreateAccount - expected one row to be affected,actual affected rows: %v", n)
	}
	if len(acc.Group) == 0 {
		return nil
	}
	args, queryInsertGrpOwnership := queryInsertIntoGroupOwnership(acc.UserId, acc.Group)
	result, err = txn.ExecContext(ctx, queryInsertGrpOwnership, args...)
	if err != nil {
		return err
	}
	n, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if n != int64(len(acc.Group)) {
		return fmt.Errorf("repo/postgres - CreateAccount - expected rows to be affected: %v , actual affected rows: %v", acc.Group, n)
	}

	return nil
}

//UserExistsByID implements Account UserExistsByID function
func (r *AccountRepository) UserExistsByID(ctx context.Context, userID string) (bool, error) {
	totalRecords := 0
	err := r.db.QueryRowContext(ctx, existsUserbyID, userID).Scan(&totalRecords)
	if err != nil {
		return false, err
	}
	return totalRecords != 0, nil
}

//UsersAll implements Account UsersAll function
func (r *AccountRepository) UsersAll(ctx context.Context, userID string) ([]*v1.AccountInfo, error) {
	rows, err := r.db.QueryContext(ctx, selectAccountWithGroupInfo, userID)
	if err != nil {
		logger.Log.Error("repo/postgres - UsersAll - failed to execute query", zap.String("reason", err.Error()))
		return nil, err
	}
	users, err := scanUserRowsWithGroupInfo(rows)
	if err != nil {
		logger.Log.Error("repo/postgres - UsersAll - failed to scan rows", zap.String("reason", err.Error()))
		return nil, err
	}
	return users, nil
}

//UsersWithUserSearchParams implements Account UsersAll function
func (r *AccountRepository) UsersWithUserSearchParams(ctx context.Context, userID string, params *v1.UserQueryParams) ([]*v1.AccountInfo, error) {
	rows, err := r.db.QueryContext(ctx, selectAccountWithQueryParams, userID)
	if err != nil {
		logger.Log.Error("repo/postgres - UsersWithUserSearchParams - failed to execute query", zap.String("reason", err.Error()))
		return nil, err
	}
	users, err := scanUserRowsWithGroupInfo(rows)
	if err != nil {
		logger.Log.Error("repo/postgres - UsersWithUserSearchParams - failed to scan rows", zap.String("reason", err.Error()))
		return nil, err
	}
	return users, nil
}

// UserOwnsGroupByID implements UserOwnsGroupByID GroupUsers function
func (r *AccountRepository) UserOwnsGroupByID(ctx context.Context, userID string, groupID int64) (bool, error) {
	totalRecords := 0
	err := r.db.QueryRowContext(ctx, existsGroupForUser, userID, groupID).Scan(&totalRecords)
	if err != nil {
		return false, err
	}
	return totalRecords != 0, nil
}

//GroupUsers implements Account GroupUsers function
func (r *AccountRepository) GroupUsers(ctx context.Context, groupID int64) ([]*v1.AccountInfo, error) {
	rows, err := r.db.QueryContext(ctx, selectAccountForGroup, groupID)
	if err != nil {
		return nil, err
	}
	users, err := scanUserRows(rows)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// // CheckPassword check the password for user
// func (r *AccountRepository) CheckPassword(ctx context.Context, userID, password string) (bool, error) {
// 	record := 0
// 	err := r.db.QueryRowContext(ctx, checkPasswordQuery, userID, password).Scan(&record)
// 	if err != nil {
// 		logger.Log.Error(" CheckPassword - failed to check password", zap.String("reason", err.Error()))
// 		return false, err
// 	}
// 	return record != 0, nil
// }

// ChangePassword ..
func (r *AccountRepository) ChangePassword(ctx context.Context, userID, password string) error {
	result, err := r.db.ExecContext(ctx, changePasswordQuery, userID, password)
	if err != nil {
		logger.Log.Error(" ChangePassword - failed to change password", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("repo/postgres - ChangePassword - expected one row to be affected,actual affected rows: %v", n)
	}
	return nil
}

//UserBelongsToAdminGroup returns true if user belongs to the admin groups
func (r *AccountRepository) UserBelongsToAdminGroup(ctx context.Context, adminUserID, userID string) (bool, error) {
	totalRecords := 0
	err := r.db.QueryRowContext(ctx, userBelongsToAdminGroup, adminUserID, userID).Scan(&totalRecords)
	if err != nil {
		return false, err
	}
	return totalRecords != 0, nil
}

func dbRoleToPostGresRole(roleDb v1.Role) (role, error) {
	switch roleDb {
	case v1.RoleAdmin:
		return roleAdmin, nil
	case v1.RoleSuperAdmin:
		return roleSuperAdmin, nil
	case v1.RoleUser:
		return roleUser, nil
	default:
		return "", fmt.Errorf("undefined role: %v", roleDb)
	}
}

func postgresRoleToDbRole(rolePS role) (v1.Role, error) {
	switch rolePS {
	case roleAdmin:
		return v1.RoleAdmin, nil
	case roleSuperAdmin:
		return v1.RoleSuperAdmin, nil
	case roleUser:
		return v1.RoleUser, nil
	default:
		return 0, fmt.Errorf("undefined role: %v", rolePS)
	}
}

func scanUserRows(rows *sql.Rows) ([]*v1.AccountInfo, error) {
	var users []*v1.AccountInfo
	for rows.Next() {
		user := &v1.AccountInfo{}
		var userRole role
		if err := rows.Scan(&user.UserId, &user.Password, &user.FirstName, &user.LastName,
			&user.Locale, &userRole); err != nil {
			return nil, err
		}
		roleDb, err := postgresRoleToDbRole(userRole)
		if err != nil {
			return nil, err
		}
		user.Role = roleDb
		users = append(users, user)
	}
	return users, nil
}

func scanUserRowsWithGroupInfo(rows *sql.Rows) ([]*v1.AccountInfo, error) {
	var users []*v1.AccountInfo
	for rows.Next() {
		user := &v1.AccountInfo{}
		var userRole role
		if err := rows.Scan(&user.UserId, &user.FirstName, &user.LastName,
			&user.Locale, pq.Array(&user.GroupName), &userRole); err != nil {
			return nil, err
		}
		roleDb, err := postgresRoleToDbRole(userRole)
		if err != nil {
			return nil, err
		}
		user.Role = roleDb
		users = append(users, user)
	}
	return users, nil
}
