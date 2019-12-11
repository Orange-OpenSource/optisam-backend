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
	"math/rand"
	v1 "optisam-backend/account-service/pkg/repository/v1"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_UpdateAccount(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		req    *v1.UpdateAccount
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, error)
		verify  func(a *AccountRepository) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx:    context.Background(),
				userID: "user8@test.com",
				req: &v1.UpdateAccount{
					Locale: "fr",
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, passwords, locales); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
			verify: func(a *AccountRepository) error {
				ai, err := a.AccountInfo(context.Background(), "user8@test.com")
				if err != nil {
					return err
				}
				if ai.Locale != "fr" {
					return fmt.Errorf("unexpected locale - expected: fr, got: %v", ai.Locale)
				}
				return nil
			},
		},
		{name: "failure user does not exist",
			r: NewAccountRepository(db),
			args: args{
				ctx:    context.Background(),
				userID: "user9@test.com",
				req: &v1.UpdateAccount{
					Locale: "fr",
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, passwords, locales); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.UpdateAccount(tt.args.ctx, tt.args.userID, tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UpdateAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func createUsers(db *sql.DB, usernames, passwords, locales []string) error {
	query := "INSERT INTO users(username,first_name,last_name,role,password,locale) VALUES "
	args := []interface{}{}
	for i := range usernames {
		// TODO" change this
		query += fmt.Sprintf("($%v,'fn','ln','Admin',$%v,$%v)", 3*i+1, 3*i+2, 3*i+3)
		args = append(args, usernames[i], passwords[i], locales[i])
		if i != len(usernames)-1 {
			query += ","
		}
	}
	_, err := db.Exec(query, args...)
	return err
}

func deleteAllUsers(db *sql.DB, users []string) error {
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

func TestAccountRepository_CreateAccount(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	grps := []*v1.Group{
		&v1.Group{
			Name:               "SUPERROOT",
			FullyQualifiedName: "SUPERROOT",
		},
		&v1.Group{
			Name:               "A",
			FullyQualifiedName: "SUPERROOT.A",
			NumberOfUsers:      3,
		},
		&v1.Group{
			Name:               "B",
			FullyQualifiedName: "SUPERROOT.A.B",
			NumberOfGroups:     2,
		},
		&v1.Group{
			Name:               "C",
			FullyQualifiedName: "SUPERROOT.A.C",
			NumberOfUsers:      1,
		},
		&v1.Group{
			Name:               "D",
			FullyQualifiedName: "SUPERROOT.A.B.D",
			NumberOfUsers:      1,
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
	rootID := int64(0)
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, *v1.AccountInfo, error)
		verify  func(a *AccountRepository) error
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, *v1.AccountInfo, error) {
				acc := &v1.AccountInfo{
					UserId:    "u1",
					FirstName: "FIRST",
					LastName:  "LAST",
					Locale:    "en",
					Role:      v1.RoleUser,
				}
				return func() error {
					return deleteAllUsers(db, []string{"u1"})
				}, acc, nil
			},
			verify: func(a *AccountRepository) error {
				ai, err := a.AccountInfo(context.Background(), "u1")
				if err != nil {
					return err
				}
				if ai.Locale != "en" {
					return fmt.Errorf("unexpected locale - expected: en, got: %v", ai.Locale)
				}
				if ai.FirstName != "FIRST" {
					return fmt.Errorf("unexpected firstname - expected: FIRST, got: %v", ai.FirstName)
				}
				if ai.LastName != "LAST" {
					return fmt.Errorf("unexpected lastname - expected: LAST, got: %v", ai.LastName)
				}
				if ai.Role != v1.RoleUser {
					return fmt.Errorf("unexpected role - expected: user, got: %v", ai.Role)
				}
				return nil
			},
		},
		{name: "success - with grps specified",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, *v1.AccountInfo, error) {
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, err
				}
				acc := &v1.AccountInfo{
					UserId:    "u2",
					FirstName: "FIRST",
					LastName:  "LAST",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group:     []int64{grps[3].ID, grps[4].ID, grps[5].ID},
				}
				return func() error {
					err := deleteAllUsers(db, []string{"u2"})
					if err != nil {
						return err
					}
					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, acc, nil
			},
			verify: func(a *AccountRepository) error {
				ai, err := a.AccountInfo(context.Background(), "u2")
				if err != nil {
					return err
				}
				tl, groups, err := a.UserOwnedGroups(context.Background(), "u2", &v1.GroupQueryParams{})
				if err != nil {
					return err
				}

				if ai.Locale != "fr" {
					return fmt.Errorf("unexpected locale - expected: fr, got: %v", ai.Locale)
				}
				if ai.FirstName != "FIRST" {
					return fmt.Errorf("unexpected firstname - expected: FIRST, got: %v", ai.FirstName)
				}
				if ai.LastName != "LAST" {
					return fmt.Errorf("unexpected lastname - expected: LAST, got: %v", ai.LastName)
				}
				if ai.Role != v1.RoleAdmin {
					return fmt.Errorf("unexpected role - expected: admin, got: %v", ai.Role)
				}
				if tl != 5 {
					return fmt.Errorf("unexpected records - expected: 5 , got: %v", tl)
				}
				compareGroupsAll(t, "UserOwnedGroups", groups, grps)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, acc, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.CreateAccount(tt.args.ctx, acc); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

// func createGroups(db *sql.DB, names, fullyQNames, parentID []string) error {
// 	query := "INSERT INTO users(name,fully_qualified_name,parent_id) VALUES "
// 	args := []interface{}{}
// 	for i := range names {
// 		query += fmt.Sprintf("($%v,$%v,$%v)", 3*i+1, 3*i+2, 3*i+3)
// 		args = append(args, names[i], fullyQNames[i], parentID[i])
// 		if i != len(names)-1 {
// 			query += ","
// 		}
// 	}
// 	_, err := db.Exec(query, args...)
// 	return err
// }

func TestAccountRepository_UserExistsByID(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, string, error)
		want    bool
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, string, error) {
				repo := NewAccountRepository(db)
				q := `
				INSERT INTO users(username,first_name,last_name,password) VALUES ('root@orange.com','root','user','password')
				`
				_, err := repo.db.ExecContext(context.Background(), q)
				if err != nil {
					return nil, "", err
				}
				return func() error {
					return deleteAllUsers(db, []string{"root@orange.com"})
				}, "root@orange.com", nil
			},
			want: true,
		},
		{name: "success-group not present",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, string, error) {
				userID := "shweta@orange.com" //non-existent
				return func() error {
					return nil
				}, userID, nil
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, userID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.UserExistsByID(tt.args.ctx, userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UserExistsByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AccountRepository.UserExistsByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountRepository_UsersAll(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.AccountInfo, error)
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, []*v1.AccountInfo, error) {
				grps := []*v1.Group{
					&v1.Group{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					&v1.Group{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfUsers:      2,
						NumberOfGroups:     1,
					},
					&v1.Group{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     1,
					},
					&v1.Group{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.B.C",
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
						NumberOfUsers:      2,
					},
				}
				repo := NewAccountRepository(db)
				rootID := int64(0)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 2, 2}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, err
				}
				acc1 := &v1.AccountInfo{
					UserId:    "u1",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc1); err != nil {
					return nil, nil, err
				}
				acc2 := &v1.AccountInfo{
					UserId:    "u2",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID, grps[4].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc2); err != nil {
					return nil, nil, err
				}
				acc3 := &v1.AccountInfo{
					UserId:    "u3",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[4].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc3); err != nil {
					return nil, nil, err
				}
				return func() error {
						err := deleteAllUsers(db, []string{"u1", "u2", "u3"})
						if err != nil {
							return err
						}
						gp := make([]int64, len(grps))
						for i := range grps {
							gp[i] = grps[i].ID
						}
						return deleteGroups(db, gp)
					}, []*v1.AccountInfo{
						&v1.AccountInfo{
							UserId:    "admin@test.com",
							FirstName: "super",
							LastName:  "admin",
							Locale:    "en",
							Role:      v1.RoleSuperAdmin,
							Group: []int64{
								1,
							},
						},
						acc1,
						acc2,
						acc3,
					}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, users, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.UsersAll(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UsersAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, users) {
				compareUsersAll(t, "UsersAll", got, users)
			}
		})
	}
}

func TestAccountRepository_GroupUsers(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.AccountInfo, int64, error)
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, []*v1.AccountInfo, int64, error) {
				grps := []*v1.Group{
					&v1.Group{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					&v1.Group{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfUsers:      2,
						NumberOfGroups:     1,
					},
					&v1.Group{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     1,
					},
					&v1.Group{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.B.C",
						NumberOfUsers:      1,
					},
					&v1.Group{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
						NumberOfUsers:      2,
					},
				}
				repo := NewAccountRepository(db)
				rootID := int64(0)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 2, 2}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, nil, 0, err
				}
				acc1 := &v1.AccountInfo{
					UserId:    "u1",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc1); err != nil {
					return nil, nil, 0, err
				}
				acc2 := &v1.AccountInfo{
					UserId:    "u2",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID, grps[4].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc2); err != nil {
					return nil, nil, 0, err
				}
				acc3 := &v1.AccountInfo{
					UserId:    "u3",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[4].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc3); err != nil {
					return nil, nil, 0, err
				}
				return func() error {
						err := deleteAllUsers(db, []string{"u1", "u2", "u3"})
						if err != nil {
							return err
						}
						gp := make([]int64, len(grps))
						for i := range grps {
							gp[i] = grps[i].ID
						}
						return deleteGroups(db, gp)
					}, []*v1.AccountInfo{
						acc1,
						acc2,
					}, grps[1].ID, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, users, groupId, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.GroupUsers(tt.args.ctx, groupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.GroupUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, users) {
				compareUsersAll(t, "UsersAll", got, users)
			}
		})
	}
}

func TestAccountRepository_UserOwnsGroupByID(t *testing.T) {
	type args struct {
		ctx     context.Context
		userID  string
		groupID string
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
			FullyQualifiedName: "SUPERROOT.A.B",
		},
		&v1.Group{
			Name:               "C",
			FullyQualifiedName: "SUPERROOT.A.C",
		},
		&v1.Group{
			Name:               "D",
			FullyQualifiedName: "SUPERROOT.A.B.D",
		},
		&v1.Group{
			Name:               "E",
			FullyQualifiedName: "SUPERROOT.A.B.E",
			NumberOfGroups:     2,
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
	rootID := int64(0)
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, int64, error)
		want    bool
		wantErr bool
	}{
		{name: "success - owns direct group",
			r: NewAccountRepository(db),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, 0, err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, //add u1 in B group
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, 0, err
				}

				return func() error {
					err := deleteAllUsers(db, []string{"u1"})
					if err != nil {
						return err
					}

					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps[2].ID, nil //B
			},
			want: true,
		},
		{name: "success - own subgroup",
			r: NewAccountRepository(db),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, 0, err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, //add u1 in B group
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, 0, err
				}

				return func() error {
					err := deleteAllUsers(db, []string{"u1"})
					if err != nil {
						return err
					}

					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps[6].ID, nil //F
			},
			want: true,
		},
		{name: "success - doesnt own ",
			r: NewAccountRepository(db),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, 0, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2, 2, 5, 5}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, 0, err
				}
				acc := &v1.AccountInfo{
					UserId: "u1",
					Role:   v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, //add u1 in B group
					},
				}
				if err := repo.CreateAccount(context.Background(), acc); err != nil {
					return nil, 0, err
				}

				return func() error {
					err := deleteAllUsers(db, []string{"u1"})
					if err != nil {
						return err
					}

					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, grps[3].ID, nil //C
			},
			want: false,
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
			got, err := tt.r.UserOwnsGroupByID(tt.args.ctx, tt.args.userID, grpID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UserOwnsGroupByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AccountRepository.UserOwnsGroupByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccountRepository_ChangePassword(t *testing.T) {
	type args struct {
		ctx      context.Context
		userID   string
		password string
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, error)
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewAccountRepository(db),
			args: args{
				ctx:      context.Background(),
				userID:   "m@m.com",
				password: "abc",
			},
			setup: func() (func() error, error) {
				repo := NewAccountRepository(db)
				if err := repo.CreateAccount(context.Background(), &v1.AccountInfo{
					UserId:    "m@m.com",
					FirstName: "fn",
					LastName:  "ln",
					Role:      v1.RoleAdmin,
					Locale:    "en",
				}); err != nil {
					return nil, err
				}

				return func() error {
					return deleteAllUsers(db, []string{"m@m.com"})
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expected from setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "error is not expected from cleanup") {
					return
				}
			}()
			if err := tt.r.ChangePassword(tt.args.ctx, tt.args.userID, tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			correct, err := tt.r.CheckPassword(tt.args.ctx, tt.args.userID, tt.args.password)
			if !assert.Empty(t, err, "error is not expected from CheckPassword") {
				return
			}
			if !assert.Equal(t, true, correct, "password did not match") {
				return
			}
			correct, err = tt.r.CheckPassword(tt.args.ctx, tt.args.userID, randomString(10))
			if !assert.Empty(t, err, "error is not expected from CheckPassword") {
				return
			}
			if !assert.Equal(t, false, correct, "password did not match") {
				return
			}
		})
	}
}

func compareUsersAll(t *testing.T, name string, exp []*v1.AccountInfo, act []*v1.AccountInfo) {
	for i := range exp {
		idx := getUserByID(exp[i].UserId, act)
		if !assert.NotEqualf(t, -1, idx, "%s.User with UserId: %s not found in users", exp[i].UserId) {
			return
		}
		compareUser(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
	}
}

func getUserByID(userID string, user []*v1.AccountInfo) int {
	for i := range user {
		if userID == user[i].UserId {
			return i
		}
	}
	return -1
}

func compareUser(t *testing.T, name string, exp *v1.AccountInfo, act *v1.AccountInfo) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.UserId != "" {
		assert.Equalf(t, exp.UserId, act.UserId, "%s.UserId should be same", name)
	}
	assert.Equalf(t, exp.FirstName, act.FirstName, "%s.FirstName should be same", name)
	assert.Equalf(t, exp.LastName, act.LastName, "%s.LastName should be same", name)
	assert.Equalf(t, exp.Locale, act.Locale, "%s.Locale should be same", name)
	assert.Equalf(t, exp.Role, act.Role, "%s.Role should be same", name)
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
