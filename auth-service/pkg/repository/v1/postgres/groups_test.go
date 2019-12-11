// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/auth-service/pkg/repository/v1"
	"strconv"
	"testing"

	"github.com/vijay1811/pq"
	"github.com/stretchr/testify/assert"
)

type group struct {
	id       int64
	name     string
	fqn      string
	scopes   []string
	parentID int64
}

func createGroup(g *group) error {
	repo := NewRepository(db)
	var id int64
	createGroupQuery := `INSERT INTO 
	groups(name, fully_qualified_name, scopes, parent_id, created_by)
	VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err := repo.db.QueryRowContext(context.Background(), createGroupQuery, g.name, g.fqn,
		pq.Array(g.scopes), g.parentID, "admin@test.com").Scan(&id); err != nil {
		return err
	}
	g.id = id
	return nil
}
func createGroupsHierarchyNew(groups []*group, hir []int) error {
	for i, group := range groups {
		if i == 0 {
			continue
		}
		group.parentID = groups[hir[i]].id
		err := createGroup(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestDefault_UserOwnedGroupsDirect(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
	}
	grps := []*group{
		&group{
			name: "SUPERROOT",
			fqn:  "SUPERROOT",
		},
		&group{
			name:   "A",
			fqn:    "SUPERROOT.A",
			scopes: []string{"Orange", "France"},
		},
		&group{
			name:   "B",
			fqn:    "SUPERROOT.A.B",
			scopes: []string{"Orange", "France"},
		},
		&group{
			name:   "C",
			fqn:    "SUPERROOT.A.C",
			scopes: []string{"Asia", "Pacific"},
		},
		&group{
			name:   "D",
			fqn:    "SUPERROOT.A.B.D",
			scopes: []string{"Apple"},
		},
	}
	rootID := int64(0)
	type cleanUpFunc func() error
	tests := []struct {
		name    string
		d       *Default
		args    args
		setup   func() ([]*v1.Group, cleanUpFunc, error)
		wantErr bool
	}{
		{name: "success, check if admin own root or not",
			args: args{
				ctx:    context.Background(),
				userID: "admin@test.com",
			},
			d: NewRepository(db),
			setup: func() ([]*v1.Group, cleanUpFunc, error) {
				return []*v1.Group{
						&v1.Group{
							ID: 1,
							Scopes: []string{
								"Orange",
								"Guinea Conakry",
								"Group",
								"France",
								"Ivory Coast",
							},
						},
					}, func() error {
						return nil
					}, nil
			},
		},
		{name: "success, admin owns multiple groups",
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			d: NewRepository(db),
			setup: func() (groups []*v1.Group, clenaup cleanUpFunc, retErr error) {
				repo := NewRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, err
				}
				grps[0].id = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, hir)
				if err != nil {
					return nil, nil, err
				}
				createAccountQuery := " INSERT INTO users (username,password,locale) VALUES('u1','password','en')"
				_, err = repo.db.ExecContext(context.Background(), createAccountQuery)
				if err != nil {
					return nil, nil, err
				}

				insertUsers := `INSERT INTO group_ownership(group_id,user_id) VALUES ($1,'u1'),($2,'u1'),($3,'u1')`
				_, err = repo.db.ExecContext(context.Background(), insertUsers, grps[2].id, grps[3].id, grps[4].id)
				if err != nil {
					return nil, nil, err
				}

				return []*v1.Group{
						&v1.Group{
							ID: grps[2].id,
							Scopes: []string{
								"Orange",
								"France",
							},
						},
						&v1.Group{
							ID: grps[3].id,
							Scopes: []string{
								"Asia",
								"Pacific",
							},
						},
						&v1.Group{
							ID: grps[4].id,
							Scopes: []string{
								"Apple",
							},
						},
					}, func() error {
						err := deleteAllUser(db, []string{"u1"})
						if err != nil {
							return err
						}
						return deleteGroups(db, []int64{grps[0].id, grps[1].id, grps[2].id, grps[3].id, grps[4].id})
					}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantGrps, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected is setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "no error is expected from cleanup")
			}()
			got, err := tt.d.UserOwnedGroupsDirect(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Default.UserOwnedGroupsDirect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareGroupsAll(t, "Groups", wantGrps, got)
			}
		})
	}
}

func compareGroupsAll(t *testing.T, name string, exp []*v1.Group, act []*v1.Group) {
	// if !assert.Lenf(t, act, len(exp), "expected number of records is: %d", len(exp)) {
	// 	return
	// }
	for i := range exp {
		idx := getGroubByID(exp[i].ID, act)
		if idx != -1 {
			compareGroup(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
		}
	}
}

func getGroubByID(grpID int64, grp []*v1.Group) int {
	for i := range grp {
		if grpID == grp[i].ID {
			return i
		}
	}
	return -1
}

func compareGroup(t *testing.T, name string, exp *v1.Group, act *v1.Group) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != 0 {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.ElementsMatchf(t, exp.Scopes, act.Scopes, "%s.Scopes should be same", name)

}

func deleteGroups(db *sql.DB, groups []int64) (retErr error) {
	query := "DELETE FROM groups WHERE id IN("
	params := make([]interface{}, len(groups))
	for i := range groups {
		params[i] = groups[i]
		query += "$" + strconv.Itoa(i+1)
		if i != len(groups)-1 {
			query += ","

		}
	}

	query += ")"

	result, err := db.Exec(query, params...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != int64(len(groups)) {
		return fmt.Errorf("expected rows to be deleted: %d, actual: %d", rows, len(groups))
	}
	return err
}

func deleteAllUser(db *sql.DB, users []string) error {
	query := "DELETE FROM users WHERE username IN("
	params := make([]interface{}, len(users))
	for i := range users {
		params[i] = users[i]
		query += "$" + strconv.Itoa(i+1)
		if i != len(users)-1 {
			query += ","
		}
	}

	query += ")"

	result, err := db.Exec(query, params...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != int64(len(users)) {
		return fmt.Errorf("rows should be deleted: %v , actual deleted: %v", len(users), rows)
	}
	return err
}
