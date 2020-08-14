// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	v1 "optisam-backend/auth-service/pkg/repository/v1"

	"github.com/lib/pq"
)

const (
	selectDirectGroupsForUser = `SELECT 
	id,
	scopes
	FROM groups
	INNER JOIN group_ownership ON groups.id  = group_ownership.group_id
	WHERE group_ownership.user_id = $1
	`
)

// UserOwnedGroupsDirect implements Database UserOwnedGroupsDirect function.
func (d *Default) UserOwnedGroupsDirect(ctx context.Context, userID string) ([]*v1.Group, error) {
	rows, err := d.db.QueryContext(ctx, selectDirectGroupsForUser, userID)
	if err != nil {
		return nil, err
	}
	var groups []*v1.Group
	for rows.Next() {
		group := &v1.Group{}
		if err := rows.Scan(&group.ID, pq.Array(&group.Scopes)); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}
