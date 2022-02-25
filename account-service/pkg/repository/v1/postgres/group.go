package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/logger"
	"strings"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	countGroupUsers       = `SELECT COUNT(*) FROM group_ownership WHERE group_ownership.group_id=groups.id`
	countGroupDirectChild = `SELECT COUNT(*) FROM groups e where groups.id=e.parent_id`

	selectGroup = `SELECT 
	id,
	name,
	fully_qualified_name,
	parent_id,
	scopes,
	(` + countGroupUsers + `) AS total_users,
	(` + countGroupDirectChild + `) AS total_groups  
	FROM groups `

	selectGroupForUser = `SELECT 
	id,
	name,
	fully_qualified_name,
	parent_id,
	scopes,
	count(*) OVER() AS total_records,
	(` + countGroupUsers + `) AS total_users,
	(` + countGroupDirectChild + `) AS total_groups  
	FROM groups
	WHERE fully_qualified_name <@ 
	(SELECT ARRAY(SELECT fully_qualified_name 
	FROM groups
	INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
	WHERE group_ownership.user_id = $1
	))`

	createGroup = `
	INSERT INTO 
	groups(name, fully_qualified_name, scopes, parent_id, created_by)
	VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	selectChildGroupsAllForGroup = selectGroup + ` WHERE fully_qualified_name <@ 
	(SELECT fully_qualified_name 
	FROM groups
	WHERE groups.id = $1
	) AND groups.id != $1`

	selectChildGroupsDirectForGroup = selectGroup + `WHERE parent_id = $1`

	selectDirectGroupsForUser = selectGroup + `INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
	WHERE group_ownership.user_id = $1`

	updateSubGroupsFullyQualifiedName = `UPDATE groups SET fully_qualified_name = $1
	|| subpath(fully_qualified_name, nlevel($2)) 
	WHERE fully_qualified_name ~ $3
	`
	updateGroupName = `UPDATE groups SET fully_qualified_name = $3, name=$2
	WHERE id = $1
    `

	selectGroupWithID = selectGroup + `WHERE id= $1 `

	selectGroupByFQN = `
	SELECT 
	id
	FROM groups 
	WHERE fully_qualified_name=$1
	`

	parentIDOfGroup = `
	SELECT
	parent_id
	FROM groups
	WHERE id=$1
	`

	deleteGroup = `DELETE FROM groups
	WHERE id=$1`

	selectRootGroup = selectGroup + `WHERE parent_id IS NULL`
)

// UserOwnedGroups implements Account UserOwnedGroups function
func (r *AccountRepository) UserOwnedGroups(ctx context.Context, userID string, params *v1.GroupQueryParams) (int, []*v1.Group, error) {
	rows, err := r.db.QueryContext(ctx, selectGroupForUser, userID)
	if err != nil {
		return 0, nil, err
	}
	totalRecords := 0
	var groups []*v1.Group
	for rows.Next() {
		group := &v1.Group{}
		parentID := sql.NullInt64{}
		if error := rows.Scan(&group.ID, &group.Name, &group.FullyQualifiedName, &parentID,
			pq.Array(&group.Scopes), &totalRecords, &group.NumberOfUsers, &group.NumberOfGroups); error != nil {
			return 0, nil, error
		}
		group.ParentID = parentID.Int64
		groups = append(groups, group)
	}
	return totalRecords, groups, err
}

// CreateGroup implements Account CreateGroup function
func (r *AccountRepository) CreateGroup(ctx context.Context, userID string, group *v1.Group) (*v1.Group, error) {
	var id int64
	if err := r.db.QueryRowContext(ctx, createGroup, group.Name, group.FullyQualifiedName,
		pq.Array(group.Scopes), group.ParentID, userID).Scan(&id); err != nil {
		return nil, err
	}
	group.ID = id

	return group, nil
}

// GroupInfo implements Account GroupInfo function
func (r *AccountRepository) GroupInfo(ctx context.Context, groupID int64) (*v1.Group, error) {
	grp := &v1.Group{}
	parentID := sql.NullInt64{}
	if err := r.db.QueryRowContext(ctx, selectGroupWithID, groupID).Scan(&grp.ID, &grp.Name, &grp.FullyQualifiedName, &parentID, pq.Array(&grp.Scopes), &grp.NumberOfUsers, &grp.NumberOfGroups); err != nil {
		return nil, err
	}
	grp.ParentID = parentID.Int64
	return grp, nil
}

// ChildGroupsDirect implements Account ChildGroupsDirect function
func (r *AccountRepository) ChildGroupsDirect(ctx context.Context, groupID int64, params *v1.GroupQueryParams) ([]*v1.Group, error) {
	rows, err := r.db.QueryContext(ctx, selectChildGroupsDirectForGroup, groupID)
	if err != nil {
		return nil, err
	}
	groups, err := scanGroupRows(rows)
	if err != nil {
		return nil, err
	}
	return groups, err
}

// ChildGroupsAll implements Account ChildGroupsAll function
func (r *AccountRepository) ChildGroupsAll(ctx context.Context, groupID int64, params *v1.GroupQueryParams) ([]*v1.Group, error) {
	rows, err := r.db.QueryContext(ctx, selectChildGroupsAllForGroup, groupID)
	if err != nil {
		return nil, err
	}
	groups, err := scanGroupRows(rows)
	if err != nil {
		return nil, err
	}
	return groups, err
}

// UserOwnedGroupsDirect implements Account UserOwnedGroupsDirect function
func (r *AccountRepository) UserOwnedGroupsDirect(ctx context.Context, userID string, params *v1.GroupQueryParams) ([]*v1.Group, error) {
	rows, err := r.db.QueryContext(ctx, selectDirectGroupsForUser, userID)
	if err != nil {
		return nil, err
	}
	groups, err := scanGroupRows(rows)
	if err != nil {
		return nil, err
	}
	return groups, err
}

// DeleteGroup implements Account DeleteGroup function
func (r *AccountRepository) DeleteGroup(ctx context.Context, groupID int64) (retErr error) {

	result, err := r.db.ExecContext(ctx, deleteGroup, groupID)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("repo/postgres - DeleteGroup - expected one row to be affected,actual affected rows: %v", n)
	}
	return nil
}

// UpdateGroup implements Account UpdateGroup function
func (r *AccountRepository) UpdateGroup(ctx context.Context, groupID int64, update *v1.GroupUpdate) (retErr error) {
	grpInfo, err := r.GroupInfo(ctx, groupID)
	if err != nil {
		logger.Log.Error(" UpdateGroup - failed to get group info", zap.String("reason", err.Error()))
		return err
	}
	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if error := txn.Rollback(); error != nil {
				logger.Log.Error(" UpdateGroup - failed to discard txn", zap.String("reason", error.Error()))
				retErr = fmt.Errorf(" UpdateGroup - cannot discard txn")
			}
			return
		}
		if error := txn.Commit(); error != nil {
			logger.Log.Error(" UpdateGroup - failed to commit txn", zap.String("reason", error.Error()))
			retErr = fmt.Errorf(" UpdateGroup - cannot commit txn")
		}
	}()
	newFullyQualifiedName := strings.TrimSuffix(grpInfo.FullyQualifiedName, grpInfo.Name) + update.Name
	result, err := r.db.ExecContext(ctx, updateGroupName, groupID, update.Name, newFullyQualifiedName)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("repo/postgres - UpdateGroup - expected one row to be affected,actual affected rows: %v", n)
	}
	regGroup := grpInfo.FullyQualifiedName + `.*{1,}`
	result, err = r.db.ExecContext(ctx, updateSubGroupsFullyQualifiedName, newFullyQualifiedName, grpInfo.FullyQualifiedName, regGroup)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil

}

// GroupExistsByFQN implements Account GroupExistsByFQN function
func (r *AccountRepository) GroupExistsByFQN(ctx context.Context, fullyQN string) (bool, error) {
	id := int64(0)
	err := r.db.QueryRowContext(ctx, selectGroupByFQN, fullyQN).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}
	return true, nil
}

// AddGroupUsers implements Account AddGroupUsers function
func (r *AccountRepository) AddGroupUsers(ctx context.Context, groupID int64, userIDs []string) error {

	args, queryInsertGrpOwnership := queryInsertUsersIntoGroupOwnership(groupID, userIDs)
	result, err := r.db.ExecContext(ctx, queryInsertGrpOwnership, args...)
	if err != nil {
		logger.Log.Error(" AddGroupUsers - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(" AddGroupUsers - failed to return affected rows in query", zap.String("reason", err.Error()))
		return err
	}

	if n != int64(len(userIDs)) {
		return fmt.Errorf("repo/postgres - AddGroupUsers - expected rows to be affected: %v , actual affected rows: %v", len(userIDs), n)
	}
	return nil
}

// DeleteGroupUsers implements Account DeleteGroupUsers function
func (r *AccountRepository) DeleteGroupUsers(ctx context.Context, groupID int64, userIDs []string) error {

	args, queryInsertGrpOwnership := queryDeleteUsersIntoGroupOwnership(groupID, userIDs)
	result, err := r.db.ExecContext(ctx, queryInsertGrpOwnership, args...)
	if err != nil {
		logger.Log.Error(" DeleteGroupUsers - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Error(" DeleteGroupUsers - failed to return affected rows in query", zap.String("reason", err.Error()))
		return err
	}

	if n != int64(len(userIDs)) {
		return fmt.Errorf("repo/postgres - DeleteGroupUsers - expected rows to be affected: %v , actual affected rows: %v", len(userIDs), n)
	}
	return nil
}

// IsGroupRoot implements Account IsGroupRoot function
func (r *AccountRepository) IsGroupRoot(ctx context.Context, groupID int64) (bool, error) {
	var parentID *int64
	err := r.db.QueryRowContext(ctx, parentIDOfGroup, groupID).Scan(&parentID)
	if err != nil {
		return false, err
	}
	return parentID == nil, nil
}

// GetRootGroup implements Account GetRootGroup function
func (r *AccountRepository) GetRootGroup(ctx context.Context) (*v1.Group, error) {
	grp := &v1.Group{}
	parentID := sql.NullInt64{}
	if err := r.db.QueryRowContext(ctx, selectRootGroup).Scan(&grp.ID, &grp.Name, &grp.FullyQualifiedName, &parentID, pq.Array(&grp.Scopes), &grp.NumberOfUsers, &grp.NumberOfGroups); err != nil {
		return nil, err
	}
	grp.ParentID = parentID.Int64
	return grp, nil
}

func queryDeleteUsersIntoGroupOwnership(groupID int64, users []string) ([]interface{}, string) {

	query := "DELETE FROM group_ownership WHERE group_id=$1 AND user_id IN ("
	args := []interface{}{
		groupID,
	}

	for i := range users {
		query += fmt.Sprintf("$%v", i+2)
		args = append(args, users[i])
		if i != len(users)-1 {
			query += ","
		}
	}
	query += ")"
	return args, query
}

func queryInsertIntoGroupOwnership(userID string, groups []int64) ([]interface{}, string) {

	query := "INSERT INTO group_ownership(user_id,group_id) VALUES "
	args := []interface{}{
		userID,
	}

	for i := range groups {
		query += fmt.Sprintf("($1,$%v)", i+2)
		args = append(args, groups[i])
		if i != len(groups)-1 {
			query += ","
		}
	}

	return args, query
}

func scanGroupRows(rows *sql.Rows) ([]*v1.Group, error) {
	var groups []*v1.Group
	for rows.Next() {
		group := &v1.Group{}
		parentID := sql.NullInt64{}
		if err := rows.Scan(&group.ID, &group.Name, &group.FullyQualifiedName, &parentID,
			pq.Array(&group.Scopes), &group.NumberOfUsers, &group.NumberOfGroups); err != nil {
			return nil, err
		}
		group.ParentID = parentID.Int64
		groups = append(groups, group)
	}
	return groups, nil
}

func queryInsertUsersIntoGroupOwnership(groupID int64, users []string) ([]interface{}, string) {

	query := "INSERT INTO group_ownership(group_id,user_id) VALUES "
	args := []interface{}{
		groupID,
	}

	for i := range users {
		query += fmt.Sprintf("($1,$%v)", i+2)
		args = append(args, users[i])
		if i != len(users)-1 {
			query += ","
		}
	}

	return args, query
}
