package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDropScopeTX(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() error
		check   func() bool
		r       *AccountRepository
		wantErr bool
	}{
		{
			name: "SuccessCase",
			setup: func() error {
				q := "insert into scopes (scope_code,scope_name,created_by)values('UNT','unittest','admin@test.com') ;"
				if _, err := db.Exec(q); err != nil {
					return err
				}
				q = "insert into groups (id,name, fully_qualified_name,scopes, parent_id,created_by) Values (1000,'UITG','ROOT.UITG','{UIT}',1,'admin@test.com');"
				if _, err := db.Exec(q); err != nil {
					return err
				}
				q = "insert into users (username,first_name, last_name, role,locale, password) Values ('test@test.com','f','l','SuperAdmin','en','p');"
				if _, err := db.Exec(q); err != nil {
					return err
				}
				q = "insert into group_ownership values(1000,'test@test.com');"
				if _, err := db.Exec(q); err != nil {
					return err
				}
				return nil
			},
			check: func() bool {
				q1 := "select scope_code from scopes where scope_code = 'UIT' ;"
				row := db.QueryRow(q1)
				count := 0
				if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("Scan failed for scopes", zap.Error(err))
					return true
				} else if count > 0 {
					return true
				}

				q1 = "select id from groups where 'UIT' = ANY(scopes::TEXT[]) ;"
				row = db.QueryRow(q1)
				if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("Scan failed for group", zap.Error(err))
					return true
				} else if count > 0 {
					return true
				}

				q1 = "select count(*) from group_ownership where group_id = 10;"
				row = db.QueryRow(q1)
				count = 0
				if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("Scan failed for group_ownership", zap.Error(err))
					return true
				} else if count > 0 {
					return true
				}
				return false
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r = NewAccountRepository(db, rc)
			if err := tt.setup(); err != nil {
				t.Errorf("Setup is failed TestDeleteScopeResourceTX  %s", err.Error())
				return
			}
			if err := tt.r.DropScopeTX(context.Background(), "UIT"); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.DeleteScopeResourceTX() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.check() {
				t.Errorf("Failed , data should be deleted")
			}
		})
	}
}

func TestAccountRepository_UpdateAccount(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		req    *v1.UpdateAccount
	}
	profile_pic := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00,
		0x10, 0x00, 0x00, 0x00, 0x0f, 0x04, 0x03, 0x00, 0x00, 0x00, 0x1f, 0x5d, 0x52, 0x1c, 0x00, 0x00, 0x00, 0x0f, 0x50,
		0x4c, 0x54, 0x45, 0x7a, 0xdf, 0xfd, 0xfd, 0xff, 0xfc, 0x39, 0x4d, 0x52, 0x19, 0x16, 0x15, 0xc3, 0x8d, 0x76, 0xc7,
		0x36, 0x2c, 0xf5, 0x00, 0x00, 0x00, 0x40, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x95, 0xc9, 0xd1, 0x0d, 0xc0, 0x20,
		0x0c, 0x03, 0xd1, 0x23, 0x5d, 0xa0, 0x49, 0x17, 0x20, 0x4c, 0xc0, 0x10, 0xec, 0x3f, 0x53, 0x8d, 0xc2, 0x02, 0x9c,
		0xfc, 0xf1, 0x24, 0xe3, 0x31, 0x54, 0x3a, 0xd1, 0x51, 0x96, 0x74, 0x1c, 0xcd, 0x18, 0xed, 0x9b, 0x9a, 0x11, 0x85,
		0x24, 0xea, 0xda, 0xe0, 0x99, 0x14, 0xd6, 0x3a, 0x68, 0x6f, 0x41, 0xdd, 0xe2, 0x07, 0xdb, 0xb5, 0x05, 0xca, 0xdb,
		0xb2, 0x9a, 0xdd, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user8@test.com",
				req: &v1.UpdateAccount{
					FirstName:  "userF8",
					LastName:   "userL8",
					Locale:     "fr",
					ProfilePic: profile_pic,
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				firstnames := []string{"first8", "first7"}
				lastnames := []string{"last8", "last7"}
				roles := []string{"Admin", "User"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsersWithProfilePic(db, usernames, firstnames, lastnames, roles, passwords, locales, [][]byte{
					[]byte("profile_pic1"),
					[]byte("profile_pic2"),
				}); err != nil {
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
				if ai.FirstName != "userF8" {
					return fmt.Errorf("unexpected firstname - expected: userF8, got: %v", ai.FirstName)
				}
				if ai.LastName != "userL8" {
					return fmt.Errorf("unexpected lastname - expected: userL8, got: %v", ai.LastName)
				}
				if ai.Locale != "fr" {
					return fmt.Errorf("unexpected locale - expected: fr, got: %v", ai.Locale)
				}
				assert.Equalf(t, ai.ProfilePic, profile_pic, "ProfilePic should be same")
				return nil
			},
		},
		{name: "failure user does not exist",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user9@test.com",
				req: &v1.UpdateAccount{
					Locale: "fr",
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				firstnames := []string{"first8", "first7"}
				lastnames := []string{"last8", "last7"}
				roles := []string{"Admin", "User"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
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

func TestAccountRepository_UpdateUserAccount(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		req    *v1.UpdateUserAccount
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user8@test.com",
				req: &v1.UpdateUserAccount{
					Role: v1.RoleAdmin,
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				firstnames := []string{"first8", "first7"}
				lastnames := []string{"last8", "last7"}
				roles := []string{"User", "User"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
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
				if ai.Role != v1.RoleAdmin {
					return fmt.Errorf("unexpected Role - expected: Admin, got: %v", ai.Role)
				}
				return nil
			},
		},
		{name: "failure user does not exist",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user9@test.com",
				req: &v1.UpdateUserAccount{
					Role: v1.RoleAdmin,
				},
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				firstnames := []string{"first8", "first7"}
				lastnames := []string{"last8", "last7"}
				roles := []string{"User", "User"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
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
			if err := tt.r.UpdateUserAccount(tt.args.ctx, tt.args.userID, tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UpdateUserAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func TestAccountRepository_ChangeUserFirstLogin(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user8@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user8@test.com", "user7@test.com"}
				firstnames := []string{"first8", "first7"}
				lastnames := []string{"last8", "last7"}
				roles := []string{"User", "User"}
				passwords := []string{"supersecret8", "supersecret7"}
				locales := []string{"en", "fr"}
				if err := createUsers(db, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
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
				if ai.FirstLogin != false {
					return fmt.Errorf("unexpected FirstLogin - expected: false, got: %v", ai.Role)
				}
				return nil
			},
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
			if err := tt.r.ChangeUserFirstLogin(tt.args.ctx, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.ChangeUserFirstLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func TestAccountRepository_AccountInfo(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
	}
	profile_pic1 := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00,
		0x10, 0x00, 0x00, 0x00, 0x0f, 0x04, 0x03, 0x00, 0x00, 0x00, 0x1f, 0x5d, 0x52, 0x1c, 0x00, 0x00, 0x00, 0x0f, 0x50,
		0x4c, 0x54, 0x45, 0x7a, 0xdf, 0xfd, 0xfd, 0xff, 0xfc, 0x39, 0x4d, 0x52, 0x19, 0x16, 0x15, 0xc3, 0x8d, 0x76, 0xc7,
		0x36, 0x2c, 0xf5, 0x00, 0x00, 0x00, 0x40, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x95, 0xc9, 0xd1, 0x0d, 0xc0, 0x20,
		0x0c, 0x03, 0xd1, 0x23, 0x5d, 0xa0, 0x49, 0x17, 0x20, 0x4c, 0xc0, 0x10, 0xec, 0x3f, 0x53, 0x8d, 0xc2, 0x02, 0x9c,
		0xfc, 0xf1, 0x24, 0xe3, 0x31, 0x54, 0x3a, 0xd1, 0x51, 0x96, 0x74, 0x1c, 0xcd, 0x18, 0xed, 0x9b, 0x9a, 0x11, 0x85,
		0x24, 0xea, 0xda, 0xe0, 0x99, 0x14, 0xd6, 0x3a, 0x68, 0x6f, 0x41, 0xdd, 0xe2, 0x07, 0xdb, 0xb5, 0x05, 0xca, 0xdb,
		0xb2, 0x9a, 0xdd, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	profile_pic2 := []byte{}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, error)
		want    *v1.AccountInfo
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "user1@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				firstnames := []string{"first1", "first2"}
				lastnames := []string{"last1", "last2"}
				roles := []string{"Admin", "User"}
				passwords := []string{"supersecret1", "supersecret2"}
				locales := []string{"en", "fr"}
				profile_pics := [][]byte{profile_pic1, profile_pic2}
				if err := createUsersWithProfilePic(db, usernames, firstnames, lastnames, roles, passwords, locales, profile_pics); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
			want: &v1.AccountInfo{
				UserID:     "user1@test.com",
				FirstName:  "first1",
				LastName:   "last1",
				Locale:     "en",
				Role:       v1.RoleAdmin,
				ProfilePic: profile_pic1,
			},
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
			got, err := tt.r.AccountInfo(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.AccountInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUser(t, "AccountInfo", tt.want, got)
			}
		})
	}
}
func createUsers(db *sql.DB, usernames, firstnames, lastnames, roles, passwords, locales []string) error {
	query := "INSERT INTO users(username,first_name,last_name,role,password,locale,first_login) VALUES "
	args := []interface{}{}
	for i := range usernames {
		// TODO" change this
		query += fmt.Sprintf("($%v,$%v,$%v,$%v,$%v,$%v,TRUE)", 6*i+1, 6*i+2, 6*i+3, 6*i+4, 6*i+5, 6*i+6)
		args = append(args, usernames[i], firstnames[i], lastnames[i], roles[i], passwords[i], locales[i])
		if i != len(usernames)-1 {
			query += ","
		}
	}
	_, err := db.Exec(query, args...)
	return err
}
func createUsersWithProfilePic(db *sql.DB, usernames, firstnames, lastnames, roles, passwords, locales []string, profile_pics [][]byte) error {
	query := "INSERT INTO users(username,first_name,last_name,role,password,locale,first_login,profile_pic) VALUES "
	args := []interface{}{}
	for i := range usernames {
		// TODO" change this
		query += fmt.Sprintf("($%v,$%v,$%v,$%v,$%v,$%v,TRUE,$%v)", 7*i+1, 7*i+2, 7*i+3, 7*i+4, 7*i+5, 7*i+6, 7*i+7)
		args = append(args, usernames[i], firstnames[i], lastnames[i], roles[i], passwords[i], locales[i], profile_pics[i])
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
		{
			Name:               "SUPERROOT",
			FullyQualifiedName: "SUPERROOT",
		},
		{
			Name:               "A",
			FullyQualifiedName: "SUPERROOT.A",
			NumberOfUsers:      3,
		},
		{
			Name:               "B",
			FullyQualifiedName: "SUPERROOT.A.B",
			NumberOfGroups:     2,
		},
		{
			Name:               "C",
			FullyQualifiedName: "SUPERROOT.A.C",
			NumberOfUsers:      1,
		},
		{
			Name:               "D",
			FullyQualifiedName: "SUPERROOT.A.B.D",
			NumberOfUsers:      1,
		},
		{
			Name:               "E",
			FullyQualifiedName: "SUPERROOT.A.B.E",
			NumberOfGroups:     2,
			NumberOfUsers:      1,
		},
		{
			Name:               "F",
			FullyQualifiedName: "SUPERROOT.A.B.E.F",
		},
		{
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, *v1.AccountInfo, error) {
				acc := &v1.AccountInfo{
					UserID:     "u1",
					FirstName:  "FIRST",
					LastName:   "LAST",
					Password:   "password",
					Locale:     "en",
					Role:       v1.RoleUser,
					FirstLogin: true,
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
				if ai.FirstLogin != true {
					return fmt.Errorf("unexpected firstlogin - expected: true, got: %v", ai.FirstLogin)
				}
				return nil
			},
		},
		{name: "success - with grps specified",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, *v1.AccountInfo, error) {
				repo := NewAccountRepository(db, rc)
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
					UserID:    "u2",
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
				if ai.FirstLogin != true {
					return fmt.Errorf("unexpected firstlogin - expected: true, got: %v", ai.FirstLogin)
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
			acc.Password = "password"
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, string, error) {
				repo := NewAccountRepository(db, rc)
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
			r: NewAccountRepository(db, rc),
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
		ctx    context.Context
		userID string
	}

	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.AccountInfo, error)
		wantErr bool
	}{
		{name: "success",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "admin@test.com",
			},
			setup: func() (func() error, []*v1.AccountInfo, error) {
				grps := []*v1.Group{
					{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfUsers:      2,
						NumberOfGroups:     1,
					},
					{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     1,
					},
					{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.B.C",
						NumberOfUsers:      1,
					},
					{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
						NumberOfUsers:      2,
					},
				}
				repo := NewAccountRepository(db, rc)
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
					UserID:    "u1",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Password:  "password",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
					GroupName: []string{"A"},
				}
				if err := repo.CreateAccount(context.Background(), acc1); err != nil {
					return nil, nil, err
				}
				acc2 := &v1.AccountInfo{
					UserID:    "u2",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
					GroupName: []string{"A"},
				}
				if err := repo.CreateAccount(context.Background(), acc2); err != nil {
					return nil, nil, err
				}
				acc3 := &v1.AccountInfo{
					UserID:    "u3",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[4].ID,
					},
					GroupName: []string{"D"},
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
			got, err := tt.r.UsersAll(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UsersAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, u := range users {
				u.Password = ""
			}

			compareUsersAll(t, "UsersAll", users, got)

		})
	}
}

func TestAccountRepository_UsersWithUserSearchParams(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
		params *v1.UserQueryParams
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, []*v1.AccountInfo, error)
		want    []*v1.AccountInfo
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
				params: &v1.UserQueryParams{},
			},
			setup: func() (func() error, []*v1.AccountInfo, error) {
				grps := []*v1.Group{
					{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
					},
					{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
					},
					{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.B.C",
					},
					{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
					},
				}
				repo := NewAccountRepository(db, rc)
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
					UserID:    "u1",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc1); err != nil {
					return nil, nil, err
				}
				acc2 := &v1.AccountInfo{
					UserID:    "u2",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
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
					UserID:    "u3",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
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
			got, err := tt.r.UsersWithUserSearchParams(tt.args.ctx, tt.args.userID, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UsersWithUserSearchParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, u := range users {
					u.Password = ""
				}
				compareUsersAll(t, "UsersWithUserSearchParams", users, got)
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, []*v1.AccountInfo, int64, error) {
				grps := []*v1.Group{
					{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
						NumberOfUsers:      2,
						NumberOfGroups:     1,
					},
					{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
						NumberOfGroups:     1,
					},
					{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.B.C",
						NumberOfUsers:      1,
					},
					{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
						NumberOfUsers:      2,
					},
				}
				repo := NewAccountRepository(db, rc)
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
					UserID:    "u1",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
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
					UserID:    "u2",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
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
					UserID:    "u3",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
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
			cleanup, users, groupID, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.GroupUsers(tt.args.ctx, groupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.GroupUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, users) {
				compareUsersAll(t, "GroupUsers", users, got)
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
		{
			Name:               "SUPERROOT",
			FullyQualifiedName: "SUPERROOT",
		},
		{
			Name:               "A",
			FullyQualifiedName: "SUPERROOT.A",
		},
		{
			Name:               "B",
			FullyQualifiedName: "SUPERROOT.A.B",
		},
		{
			Name:               "C",
			FullyQualifiedName: "SUPERROOT.A.C",
		},
		{
			Name:               "D",
			FullyQualifiedName: "SUPERROOT.A.B.D",
		},
		{
			Name:               "E",
			FullyQualifiedName: "SUPERROOT.A.B.E",
			NumberOfGroups:     2,
		},
		{
			Name:               "F",
			FullyQualifiedName: "SUPERROOT.A.B.E.F",
		},
		{
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db, rc)
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
					UserID:   "u1",
					Role:     v1.RoleAdmin,
					Password: "password",
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db, rc)
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
					UserID:   "u1",
					Role:     v1.RoleAdmin,
					Password: "password",
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:    context.Background(),
				userID: "u1",
			},
			setup: func() (func() error, int64, error) {
				repo := NewAccountRepository(db, rc)
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
					UserID:   "u1",
					Role:     v1.RoleAdmin,
					Password: "password",
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
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:      context.Background(),
				userID:   "m@m.com",
				password: "abc",
			},
			setup: func() (func() error, error) {
				repo := NewAccountRepository(db, rc)
				if err := repo.CreateAccount(context.Background(), &v1.AccountInfo{
					UserID:    "m@m.com",
					FirstName: "fn",
					LastName:  "ln",
					Password:  "password",
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

			ai, err := tt.r.AccountInfo(tt.args.ctx, tt.args.userID)
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}

			assert.Equalf(t, tt.args.password, ai.Password, "expected password to be changed")
			// correct, err := tt.r.CheckPassword(tt.args.ctx, tt.args.userID, tt.args.password)
			// if !assert.Empty(t, err, "error is not expected from CheckPassword") {
			// 	return
			// }
			// if !assert.Equal(t, true, correct, "password did not match") {
			// 	return
			// }
			// correct, err = tt.r.CheckPassword(tt.args.ctx, tt.args.userID, randomString(10))
			// if !assert.Empty(t, err, "error is not expected from CheckPassword") {
			// 	return
			// }
			// if !assert.Equal(t, false, correct, "password did not match") {
			// 	return
			// }
		})
	}
}

func TestAccountRepository_UserBelongsToAdminGroup(t *testing.T) {
	type args struct {
		ctx         context.Context
		adminUserID string
		userID      string
	}
	tests := []struct {
		name    string
		r       *AccountRepository
		args    args
		setup   func() (func() error, error)
		want    bool
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewAccountRepository(db, rc),
			args: args{
				ctx:         context.Background(),
				adminUserID: "admin1@test.com",
				userID:      "user1@test.com",
			},
			setup: func() (func() error, error) {
				grps := []*v1.Group{
					{
						Name:               "SUPERROOT",
						FullyQualifiedName: "SUPERROOT",
					},
					{
						Name:               "A",
						FullyQualifiedName: "SUPERROOT.A",
					},
					{
						Name:               "B",
						FullyQualifiedName: "SUPERROOT.A.B",
					},
					{
						Name:               "C",
						FullyQualifiedName: "SUPERROOT.A.C",
					},
					{
						Name:               "D",
						FullyQualifiedName: "SUPERROOT.A.B.D",
					},
				}
				repo := NewAccountRepository(db, rc)
				rootID := int64(0)
				rootQuery := `INSERT INTO groups(name,fully_qualified_name) VALUES ('SUPERROOT','SUPERROOT') returning id`
				if err := repo.db.QueryRowContext(context.Background(), rootQuery).Scan(&rootID); err != nil {
					return nil, err
				}
				grps[0].ID = rootID
				hir := []int{-1, 0, 1, 1, 2}
				err := createGroupsHierarchyNew(grps, "admin@test.com", hir)
				if err != nil {
					return nil, err
				}
				acc1 := &v1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[1].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc1); err != nil {
					return nil, err
				}
				acc2 := &v1.AccountInfo{
					UserID:    "admin2@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[2].ID, grps[3].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc2); err != nil {
					return nil, err
				}
				acc3 := &v1.AccountInfo{
					UserID:    "user1@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Password:  "password",
					Locale:    "fr",
					Role:      v1.RoleAdmin,
					Group: []int64{
						grps[4].ID,
					},
				}
				if err := repo.CreateAccount(context.Background(), acc3); err != nil {
					return nil, err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"admin1@test.com", "admin2@test.com", "user1@test.com"})
					if err != nil {
						return err
					}

					gp := make([]int64, len(grps))
					for i := range grps {
						gp[i] = grps[i].ID
					}
					return deleteGroups(db, gp)
				}, nil
			},
			want: true,
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
			got, err := tt.r.UserBelongsToAdminGroup(tt.args.ctx, tt.args.adminUserID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountRepository.UserBelongsToAdminGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AccountRepository.UserBelongsToAdminGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareUsersAll(t *testing.T, name string, exp []*v1.AccountInfo, act []*v1.AccountInfo) {
	for i := range exp {
		idx := getUserByID(exp[i].UserID, act)
		if !assert.NotEqualf(t, -1, idx, "%s.User with UserId: %s not found in users", exp[i].UserID) {
			return
		}
		compareUser(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
	}
}

func getUserByID(userID string, user []*v1.AccountInfo) int {
	for i := range user {
		if userID == user[i].UserID {
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

	if exp.UserID != "" {
		assert.Equalf(t, exp.UserID, act.UserID, "%s.UserId should be same", name)
	}
	assert.Equalf(t, exp.FirstName, act.FirstName, "%s.FirstName should be same", name)
	assert.Equalf(t, exp.LastName, act.LastName, "%s.LastName should be same", name)
	assert.Equalf(t, exp.Locale, act.Locale, "%s.Locale should be same", name)
	assert.Equalf(t, exp.Role, act.Role, "%s.Role should be same", name)
	assert.Equalf(t, exp.ProfilePic, act.ProfilePic, "%s.ProfilePic should be same", name)
	if exp.Password != "" {
		assert.Equalf(t, exp.Password, act.Password, "%s.Password should be same", name)
	}
	if exp.ContFailedLogin != int16(0) {
		assert.Equalf(t, exp.ContFailedLogin, act.ContFailedLogin, "%s.ContFailedLogin should be same", name)
	}
	if exp.ProfilePic != nil {
		assert.Equalf(t, exp.ProfilePic, act.ProfilePic, "%s.ProfilePic should be same", name)
	}
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
