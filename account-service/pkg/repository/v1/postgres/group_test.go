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
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_UserOwnedGroups(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		params *v1.GroupQueryParams
	}
	type cleanUpFunc func() error
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		want    int
		setup   func() ([]*v1.Group, cleanUpFunc, error)
		wantErr bool
	}{
		{name: "success, check if admin own root or not",
			args: args{
				ctx:    context.Background(),
				userID: "admin@test.com",
			},
			r:    NewAccountRepository(db),
			want: 1,
			setup: func() ([]*v1.Group, cleanUpFunc, error) {
				return []*v1.Group{
						&v1.Group{
							ID:                 1,
							Name:               "ROOT",
							FullyQualifiedName: "ROOT",
							ParentID:           0,
							NumberOfUsers:      1,
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
				userID: "admin@test.com",
			},
			r:    NewAccountRepository(db),
			want: 5,
			setup: func() (grps []*v1.Group, clenaup cleanUpFunc, retErr error) {
				repo := NewAccountRepository(db)
				g1 := &v1.Group{
					Name:               "G1",
					FullyQualifiedName: "ROOT.G1",
					ParentID:           1,
					NumberOfGroups:     1,
					Scopes: []string{
						"Orange",
						"Guinea Conakry",
						"Group",
						"France",
					},
				}

				grp, err := repo.CreateGroup(context.Background(), "admin@test.com", g1)
				if err != nil {
					return nil, nil, err
				}
				g1 = grp
				defer func() {
					if retErr != nil {
						assert.Empty(t, deleteGroups(db, []int64{g1.ID}), "error is not expected from delete group")
					}
				}()
				g2 := &v1.Group{
					Name:               "G2",
					FullyQualifiedName: "ROOT.G2",
					ParentID:           1,
					NumberOfGroups:     1,
					Scopes: []string{
						"Guinea Conakry",
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g2)
				if err != nil {
					return nil, nil, err
				}
				g2 = grp
				defer func() {
					if retErr != nil {
						assert.Empty(t, deleteGroups(db, []int64{g2.ID}), "error is not expected from delete group")
					}
				}()

				g3 := &v1.Group{
					Name:               "G3",
					FullyQualifiedName: "ROOT.G1.G3",
					ParentID:           g1.ID,
					Scopes: []string{
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g3)
				if err != nil {
					return nil, nil, err
				}
				g3 = grp

				defer func() {
					if retErr != nil {
						assert.Empty(t, deleteGroups(db, []int64{g3.ID}), "error is not expected from delete group")
					}
				}()

				g4 := &v1.Group{
					Name:               "G3",
					FullyQualifiedName: "ROOT.G2.G4",
					ParentID:           g2.ID,
					Scopes: []string{
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g4)
				if err != nil {
					return nil, nil, err
				}
				g4 = grp

				defer func() {
					if retErr != nil {
						assert.Empty(t, deleteGroups(db, []int64{g4.ID}), "error is not expected from delete group")
					}
				}()

				// deleteGroupIDs := make([]int64, len(users))

				return []*v1.Group{
						&v1.Group{
							ID:                 1,
							Name:               "ROOT",
							FullyQualifiedName: "ROOT",
							ParentID:           0,
							NumberOfUsers:      1,
							NumberOfGroups:     2,
							Scopes: []string{
								"Orange",
								"Guinea Conakry",
								"Group",
								"France",
								"Ivory Coast",
							},
						},
						g1, g2, g3, g4,
					}, func() error {
						return deleteGroups(db, []int64{g1.ID, g2.ID, g3.ID, g4.ID})
					}, nil
			},
		},
		{name: "success, non admin owns multiple groups",
			args: args{
				ctx:    context.Background(),
				userID: "user7@test.com",
			},
			r:    NewAccountRepository(db),
			want: 2,
			setup: func() (grps []*v1.Group, clenaup cleanUpFunc, retErr error) {
				repo := NewAccountRepository(db)
				g1 := &v1.Group{
					Name:               "G1",
					FullyQualifiedName: "ROOT.G1",
					ParentID:           1,
					NumberOfGroups:     1,
					NumberOfUsers:      0,
					Scopes: []string{
						"Orange",
						"Guinea Conakry",
						"Group",
						"France",
					},
				}

				grp, err := repo.CreateGroup(context.Background(), "admin@test.com", g1)
				if err != nil {
					return nil, nil, err
				}
				g1 = grp
				defer func() {
					if retErr != nil {
						deleteGroups(db, []int64{g1.ID})
					}
				}()
				g2 := &v1.Group{
					Name:               "G2",
					FullyQualifiedName: "ROOT.G2",
					ParentID:           1,
					NumberOfGroups:     1,
					Scopes: []string{
						"Guinea Conakry",
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g2)
				if err != nil {
					return nil, nil, err
				}
				g2 = grp
				defer func() {
					if retErr != nil {
						deleteGroups(db, []int64{g2.ID})
					}
				}()

				g3 := &v1.Group{
					Name:               "G3",
					FullyQualifiedName: "ROOT.G1.G3",
					ParentID:           g1.ID,
					NumberOfUsers:      1,
					Scopes: []string{
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g3)
				if err != nil {
					return nil, nil, err
				}
				g3 = grp

				defer func() {
					if retErr != nil {
						deleteGroups(db, []int64{g3.ID})
					}
				}()

				g4 := &v1.Group{
					Name:               "G4",
					FullyQualifiedName: "ROOT.G2.G4",
					ParentID:           g2.ID,
					NumberOfUsers:      1,
					Scopes: []string{
						"Group",
						"France",
					},
				}

				grp, err = repo.CreateGroup(context.Background(), "admin@test.com", g4)
				if err != nil {
					return nil, nil, err
				}
				g4 = grp

				defer func() {
					if retErr != nil {
						deleteGroups(db, []int64{g4.ID})
					}
				}()

				repo.CreateAccount(context.Background(), &v1.AccountInfo{
					UserId:    "user7@test.com",
					FirstName: "vishal",
					LastName:  "mishra",
					Locale:    "en",
					Role:      v1.RoleAdmin,
					Group: []int64{
						g3.ID,
						g4.ID,
					},
				})
				return []*v1.Group{
						g3, g4,
					}, func() error {
						if err := deleteAllUsers(db, []string{"user7@test.com"}); err != nil {
							return err
						}
						return deleteGroups(db, []int64{g1.ID, g2.ID, g3.ID, g4.ID})
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
			got, actGroups, err := tt.r.UserOwnedGroups(tt.args.ctx, tt.args.userID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UserOwnedGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !assert.Equal(t, tt.want, got, "totlal records should be equal") {
					return
				}
				compareGroupsAll(t, "Groups", wantGrps, actGroups)
			}
		})
	}
}

func TestAccountRepository_DeleteGroup(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, int64, error)
		verify  func(a *AccountRepository, grpID int64) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, int64, error) {

				repo := NewAccountRepository(db)
				acc := &v1.AccountInfo{
					UserId:    "admintest",
					FirstName: "FIRST",
					LastName:  "LAST",
					Locale:    "en",
					Role:      v1.RoleAdmin,
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, 0, err
				}
				grp, err := repo.CreateGroup(context.Background(), "admintest", &v1.Group{
					Name:               "A",
					FullyQualifiedName: "A.B",
					ParentID:           1,
				})
				if err != nil {
					return nil, 0, err
				}
				// grpID := int64(0)
				// rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('A','A.B') returning id`
				// if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&grpID); err != nil {
				// 	return nil, 0, err
				// }
				accUser := &v1.AccountInfo{
					UserId:    "usertest",
					FirstName: "FIRST",
					LastName:  "LAST",
					Locale:    "en",
					Role:      v1.RoleUser,
					Group: []int64{
						grp.ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), accUser); err != nil {
					return nil, 0, fmt.Errorf("cannot create user")
				}
				return func() error {
					return deleteAllUsers(db, []string{"admintest", "usertest"})
				}, grp.ID, nil
			},
			verify: func(a *AccountRepository, grpID int64) error {
				ok, err := a.GroupExistsByFQN(context.Background(), "A.B")
				if err != nil {
					return err
				}
				if ok {
					return fmt.Errorf("group exists")
				}
				q := `SELECT count(*) FROM group_ownership WHERE group_id=$1`
				n := -1
				err = a.db.QueryRowContext(context.Background(), q, grpID).Scan(&n)
				if err != nil {
					return err
				}
				if n != 0 {
					return fmt.Errorf("rows not deleted from group_ownership")
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.DeleteGroup(tt.args.ctx, grpID); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r, grpID))
			}
		})
	}
}

func TestAccountRepository_UpdateGroup(t *testing.T) {
	type args struct {
		ctx    context.Context
		update *v1.GroupUpdate
	}
	//     A
	//     /\
	//    B  C
	//    /\
	//    D E
	//      /\
	//      F G

	rootID := int64(0)
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.Group, int64, error)
		verify  func(a *AccountRepository, grps []*v1.Group, grpID int64) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
				update: &v1.GroupUpdate{
					Name: "Z",
				},
			},
			setup: func() (func() error, []*v1.Group, int64, error) {
				grps := []*v1.Group{
					&v1.Group{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					&v1.Group{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfGroups:     2,
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     2,
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.C",
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
					},
					&v1.Group{
						Name:               "E",
						FullyQualifiedName: "SUPERROOT.A.B.E",
						NumberOfGroups:     2,
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "F",
						FullyQualifiedName: "SUPERROOT.A.B.E.F",
					},
					&v1.Group{
						Name:               "G",
						FullyQualifiedName: "SUPERROOT.A.B.E.G",
					},
				}
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, 0, err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u2",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u3",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u4",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u5",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[5].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u1", "u2", "u3", "u4", "u5"})
					if err != nil {
						return err
					}

					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps, grps[2].ID, nil //B
			},
			verify: func(a *AccountRepository, grps []*v1.Group, grpID int64) error {
				// grps, err := a.ChildGroupsAll(context.Background(), rootID, &v1.GroupQueryParams{})
				// if err != nil {
				// 	return err
				// }
				changeHir(grps, "SUPERROOT.A.B", "SUPERROOT.A.Z")
				idx := getGroubByID(grpID, grps)
				grps[idx].Name = "Z"
				grps[idx].FullyQualifiedName = "SUPERROOT.A.Z"
				group, err := a.GroupInfo(context.Background(), grpID)
				if err != nil {
					return err
				}
				compareGroup(t, "Group", group, grps[idx])
				subGrps, err := a.ChildGroupsAll(context.Background(), grpID, &v1.GroupQueryParams{})
				if err != nil {
					return err
				}
				compareGroupsAll(t, "GroupsAll", subGrps, grps)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grps, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.UpdateGroup(tt.args.ctx, grpID, tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UpdateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r, grps, grpID))
			}
		})
	}
}

// func printGroups(prefix string, groups []*v1.Group) {
// 	for i, group := range groups {
// 		fmt.Printf("%v.groups[%d]%+v\n", prefix, i, group)
// 	}
// }

func TestAccountRepository_ChildGroupsDirect(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *v1.GroupQueryParams
	}
	//group hierarchy

	rootID := int64(0)
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.Group, int64, error)
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, []*v1.Group, int64, error) {
				grps := []*v1.Group{
					&v1.Group{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					&v1.Group{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfGroups:     2,
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     2,
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.C",
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
					},
					&v1.Group{
						Name:               "E",
						FullyQualifiedName: "SUPERROOT.A.B.E",
						NumberOfGroups:     2,
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "F",
						FullyQualifiedName: "SUPERROOT.A.B.E.F",
					},
					&v1.Group{
						Name:               "G",
						FullyQualifiedName: "SUPERROOT.A.B.E.G",
					},
				}
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, 0, err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u2",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u3",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u4",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				acc = &v1.AccountInfo{
					UserId: "u5",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[5].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, 0, err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u1", "u2", "u3", "u4", "u5"})
					if err != nil {
						return err
					}
					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps, grps[2].ID, nil //B
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grps, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.ChildGroupsDirect(tt.args.ctx, grpID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.ChildGroupsDirect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareGroupsAll(t, "GroupsAll", got, grps)
			}
		})
	}
}

func TestAccountRepository_UserOwnedGroupsDirect(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *v1.GroupQueryParams
	}

	rootID := int64(0)
	//group hierarchy
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.Group, string, error)
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, []*v1.Group, string, error) {
				grps := []*v1.Group{
					&v1.Group{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					&v1.Group{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfGroups:     2,
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     2,
						NumberOfUsers:      2,
					},
					&v1.Group{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.C",
						NumberOfUsers:      3,
					},
					&v1.Group{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
					},
					&v1.Group{
						Name:               "E",
						FullyQualifiedName: "SUPERROOT.A.B.E",
						NumberOfGroups:     2,
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "F",
						FullyQualifiedName: "SUPERROOT.A.B.E.F",
					},
					&v1.Group{
						Name:               "G",
						FullyQualifiedName: "SUPERROOT.A.B.E.G",
					},
				}
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, "", err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, "", err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				acc = &v1.AccountInfo{
					UserId: "u2",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				acc = &v1.AccountInfo{
					UserId: "u3",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				acc = &v1.AccountInfo{
					UserId: "u4",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				acc = &v1.AccountInfo{
					UserId: "u5",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[5].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				acc = &v1.AccountInfo{
					UserId: "u6",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, nil, "", err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u1", "u2", "u3", "u4", "u5", "u6"})
					if err != nil {
						return err
					}
					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps, "u6", nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grps, userID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.UserOwnedGroupsDirect(tt.args.ctx, userID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UserOwnedGroupsDirect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareGroupsAll(t, "GroupsAll", got, grps)
			}
		})
	}
}

// func TestAccountRepository_GroupExistsByID(t *testing.T) {
// 	type args struct {
// 		ctx     context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *AccountRepository
// 		args    args
// 		setup   func() (func() error, int64, error)
// 		want    bool
// 		wantErr bool
// 	}{
// 		{name: "success",
// 			r: NewAccountRepository(db),
// 			args: args{
// 				ctx: context.Background(),
// 			},
// 			setup: func() (func() error, int64, error) {
// 				repo := NewAccountRepository(db)
// 				q := `
// 				INSERT INTO groups(name, fully_qualified_name) VALUES ('A','ROOT.A') RETURNING id
// 				`
// 				grpID := int64(0)
// 				if err := repo.db.QueryRowContext(context.Background(), q).Scan(&grpID); err != nil {
// 					return nil, 0, err
// 				}
// 				return func() error {
// 					return deleteGroups(db, []int64{grpID})
// 				}, grpID, nil
// 			},
// 			want: true,
// 		},
// 		{name: "success-group not present",
// 			r: NewAccountRepository(db),
// 			args: args{
// 				ctx: context.Background(),
// 			},
// 			setup: func() (func() error, int64, error) {
// 				grpID := int64(999) //non-existent
// 				return func() error {
// 					return nil
// 				}, grpID, nil
// 			},
// 			want: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cleanup, grpID, err := tt.setup()
// 			if !assert.Empty(t, err, "no error is expected in setup") {
// 				return
// 			}
// 			defer func() {
// 				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
// 					return
// 				}
// 			}()
// 			got, err := tt.r.GroupExistsByID(tt.args.ctx, grpID)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("AccountRepository.GroupExistsByID() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("AccountRepository.GroupExistsByID() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestAccountRepository_AddGroupUsers(t *testing.T) {
	type args struct {
		ctx     context.Context
		userIDs []string
	}

	type groupWithUsers struct {
		groupID int64
		users   []*v1.AccountInfo
	}
	grps := []*v1.Group{
		&v1.Group{
			Name:               "SUPERROOT",
			FullyQualifiedName: "SUPERROOT",
		},
		&v1.Group{
			Name:               "A",
			FullyQualifiedName: "SUPERROOT.A",
		},
		&v1.Group{
			Name:               "B",
			FullyQualifiedName: "SUPERROOT.B",
		},
		&v1.Group{
			Name:               "C",
			FullyQualifiedName: "SUPERROOT.C",
		},
		&v1.Group{
			Name:               "D",
			FullyQualifiedName: "SUPERROOT.B.D",
		},
		&v1.Group{
			Name:               "E",
			FullyQualifiedName: "SUPERROOT.B.E",
			NumberOfGroups:     2,
		},
		&v1.Group{
			Name:               "F",
			FullyQualifiedName: "SUPERROOT.C.F",
		},
	}

	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*groupWithUsers, int64, error)
		verify  func(a *AccountRepository, gwu []*groupWithUsers) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx:     context.Background(),
				userIDs: []string{"u2"},
			},
			setup: func() (func() error, []*groupWithUsers, int64, error) {
				repo := NewAccountRepository(db)
				grpID := int64(0)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&grpID); err != nil {
					return nil, nil, 0, err
				}
				grps[0].ID = grpID
				hir := []int{-1, 0, 0, 0, 2, 2, 3}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, 0, err
				}
				accounts := []*v1.AccountInfo{
					&v1.AccountInfo{
						UserId:    "u1",
						FirstName: "F1",
						LastName:  "L1",
						Locale:    "en",
						Role:      v1.RoleAdmin,
						Group:     []int64{grps[1].ID, grps[4].ID, grps[6].ID},
					},
					&v1.AccountInfo{
						UserId:    "u2",
						FirstName: "F2",
						LastName:  "L2",
						Locale:    "en",
						Role:      v1.RoleAdmin,
						Group:     []int64{grps[4].ID, grps[6].ID},
					},
					&v1.AccountInfo{
						UserId:    "u3",
						FirstName: "F3",
						LastName:  "L3",
						Locale:    "en",
						Role:      v1.RoleAdmin,
						Group:     []int64{grps[0].ID},
					},
				}
				for _, acc := range accounts {
					if err := repo.CreateAccount(context.Background(), acc); err != nil {
						return nil, nil, 0, err
					}
				}
				gwu := []*groupWithUsers{
					&groupWithUsers{
						groupID: grps[0].ID,
						users: []*v1.AccountInfo{
							accounts[1],
							accounts[2],
						},
					},
					&groupWithUsers{
						groupID: grps[4].ID,
						users: []*v1.AccountInfo{
							accounts[0],
						},
					},
					&groupWithUsers{
						groupID: grps[6].ID,
						users: []*v1.AccountInfo{
							accounts[0],
						},
					},
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u1", "u2", "u3"})
					if err != nil {
						return err
					}
					groupIDs := make([]int64, len(grps))
					for i := range grps {
						groupIDs[i] = grps[i].ID
					}
					return deleteGroups(db, groupIDs)
				}, gwu, grpID, nil
			},
			verify: func(a *AccountRepository, gwu []*groupWithUsers) error {
				repo := NewAccountRepository(db)
				for i := range gwu {
					acc, err := repo.GroupUsers(context.Background(), gwu[i].groupID)
					if err != nil {
						return err
					}
					compareUsersAll(t, "GroupUsers", gwu[i].users, acc)
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, gwu, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.AddGroupUsers(tt.args.ctx, grpID, tt.args.userIDs); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.AddGroupUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r, gwu))
			}
		})
	}
}

func TestAccountRepository_DeleteGroupUsers(t *testing.T) {
	type args struct {
		ctx     context.Context
		groupID int64
		userIDs []string
	}
	accounts := []*v1.AccountInfo{
		&v1.AccountInfo{
			UserId:    "u1",
			FirstName: "F1",
			LastName:  "L1",
			Locale:    "en",
			Role:      v1.RoleAdmin,
		},
		&v1.AccountInfo{
			UserId:    "u2",
			FirstName: "F2",
			LastName:  "L2",
			Locale:    "en",
			Role:      v1.RoleAdmin,
		},
		&v1.AccountInfo{
			UserId:    "u3",
			FirstName: "F3",
			LastName:  "L3",
			Locale:    "en",
			Role:      v1.RoleAdmin,
		},
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, int64, error)
		verify  func(a *AccountRepository, grpID int64) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx:     context.Background(),
				userIDs: []string{"u1", "u3"},
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db)
				grpID := int64(0)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('TEST','TEST') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&grpID); err != nil {
					return nil, 0, err
				}
				for _, acc := range accounts {
					if err := repo.CreateAccount(context.Background(), acc); err != nil {
						return nil, 0, err
					}
				}

				err := repo.AddGroupUsers(context.Background(), grpID, []string{"u1", "u2", "u3"})
				if err != nil {
					return nil, 0, err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u2"})
					if err != nil {
						return err
					}
					return deleteGroups(db, []int64{grpID})
				}, grpID, nil //B
			},
			verify: func(a *AccountRepository, grpID int64) error {
				repo := NewAccountRepository(db)
				acc, err := repo.GroupUsers(context.Background(), grpID)
				if err != nil {
					return err
				}
				//printUsers("got", acc)
				//printUsers("actual", accounts[1:2])
				compareUsersAll(t, "GroupUsers", accounts[1:2], acc)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.DeleteGroupUsers(tt.args.ctx, grpID, tt.args.userIDs); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.DeleteGroupUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r, grpID))
			}
		})
	}
}

func printUsers(prefix string, users []*v1.AccountInfo) {
	for i, user := range users {
		fmt.Printf("%v.groups[%d]%+v\n", prefix, i, user)
	}
}

func compareGroupsAll(t *testing.T, name string, exp []*v1.Group, act []*v1.Group) {
	// if !assert.Lenf(t, act, len(exp), "expected number of records is: %d", len(exp)) {
	// 	return
	// }
	for i := range exp {
		idx := getGroubByID(exp[i].ID, act)
		if !assert.NotEqualf(t, -1, idx, "group by id: %d not found in expected groups", exp[i].ID) {
			return
		}
		compareGroup(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
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
	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
	assert.Equalf(t, exp.FullyQualifiedName, act.FullyQualifiedName, "%s.FullyQualifiedName should be same", name)
	assert.Equalf(t, exp.ParentID, act.ParentID, "%s.ParentID should be same", name)
	assert.ElementsMatchf(t, exp.Scopes, act.Scopes, "%s.Scopes should be same", name)
	assert.Equalf(t, exp.NumberOfGroups, act.NumberOfGroups, "%s.Number of Groups should be same", name)
	assert.Equalf(t, exp.NumberOfUsers, act.NumberOfUsers, "%s.Number of users should be same", name)

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

// []int{-1,0,1,1,2,2,5,5}
func createGroupsHierarchyNew(groups []*v1.Group, userID string, hir []int) error {
	repo := NewAccountRepository(db)
	for i, group := range groups {
		if i == 0 {
			continue
		}
		group.ParentID = groups[hir[i]].ID
		grp, err := repo.CreateGroup(context.Background(), userID, group)
		if err != nil {
			return err
		}
		group.ID = grp.ID
	}
	return nil
}

func changeHir(groups []*v1.Group, orig, repl string) {
	for _, group := range groups {
		group.FullyQualifiedName = strings.Replace(group.FullyQualifiedName, orig, repl, 1)
	}
}

func TestAccountRepository_IsGroupRoot(t *testing.T) {
	type args struct {
		ctx     context.Context
		groupID int64
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, int64, error)
		want    bool
		wantErr bool
	}{
		{name: "success - root group",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, int64, error) {
				grpID := int64(1)
				return func() error {
					return nil
				}, grpID, nil
			},
			want: true,
		},
		{name: "success-group present but not root",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db)
				q := `
						INSERT INTO groups(name, fully_qualified_name,parent_id) VALUES ('A','ROOT.A',1) RETURNING id
						`
				grpID := int64(0)
				if err := repo.db.QueryRowContext(context.Background(), q).Scan(&grpID); err != nil {
					return nil, 0, err
				}
				return func() error {
					return deleteGroups(db, []int64{grpID})
				}, grpID, nil
			},
			want: false,
		},
		{name: "success-group not present",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, int64, error) {
				grpID := int64(999) //non-existent
				return func() error {
					return nil
				}, grpID, nil
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, grpID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.IsGroupRoot(tt.args.ctx, grpID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.IsGroupRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AccountRepository.IsGroupRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
