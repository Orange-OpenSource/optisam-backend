// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"optisam-backend/account-service/pkg/repository/v1/postgres/db"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueries_DeleteUser(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		r       *db.Queries
		args    args
		setup   func() (func() error, string, error)
		verify  func(userID string) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: db.New(sqldb),
			args: args{
				ctx:    context.Background(),
				userID: "admin1@test.com",
			},
			setup: func() (func() error, string, error) {
				usernames := []string{"admin1@test.com", "user1@test.com"}
				firstnames := []string{"adminf1", "userf1"}
				lastnames := []string{"adminl1", "userl1"}
				roles := []string{"Admin", "User"}
				passwords := []string{"admin1", "user1"}
				locales := []string{"en", "fr"}
				if err := createUsers(sqldb, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
					return nil, "", err
				}
				return func() error {
					err := deleteAllUsers(sqldb, []string{"user1@test.com"})
					if err != nil {
						return err
					}
					return nil
				}, "admin1@test.com", nil
			},
			verify: func(userID string) error {
				query := "SELECT * from users WHERE username=$1"
				user := db.User{}
				err := sqldb.QueryRowContext(context.Background(), query, userID).Scan(&user.Username)
				if err != nil {
					if err == sql.ErrNoRows {
						return nil
					}
					return err
				}
				return fmt.Errorf("user not deleted")
			},
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
			if err := tt.r.DeleteUser(tt.args.ctx, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("Queries.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(userID))
			}
		})
	}
}

func TestQueries_InsertUserAudit(t *testing.T) {
	type args struct {
		ctx context.Context
		arg db.InsertUserAuditParams
	}
	tests := []struct {
		name    string
		q       *db.Queries
		args    args
		cleanup func(userID string) error
		verify  func(userID string) error
		wantErr bool
	}{
		{name: "SUCCESS",
			q: db.New(sqldb),
			args: args{
				ctx: context.Background(),
				arg: db.InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            "Admin",
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       db.AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				},
			},
			cleanup: func(userID string) error {
				query := "DELETE FROM users_audit WHERE username=$1"
				_, err := sqldb.ExecContext(context.Background(), query, userID)
				return err
			},
			verify: func(userID string) error {
				query := "SELECT username from users_audit WHERE username=$1"
				userAudit := db.UsersAudit{}
				if err := sqldb.QueryRowContext(context.Background(), query, userID).Scan(&userAudit.Username); err != nil {
					return err
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if !assert.Empty(t, tt.cleanup(tt.args.arg.Username), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.q.InsertUserAudit(tt.args.ctx, tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Queries.InsertUserAudit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.args.arg.Username))
			}
		})
	}
}

func createUsers(db *sql.DB, usernames, firstnames, lastnames, roles, passwords, locales []string) error {
	query := "INSERT INTO users(username,first_name,last_name,role,password,locale) VALUES "
	args := []interface{}{}
	for i := range usernames {
		// TODO" change this
		query += fmt.Sprintf("($%v,$%v,$%v,$%v,$%v,$%v)", 6*i+1, 6*i+2, 6*i+3, 6*i+4, 6*i+5, 6*i+6)
		args = append(args, usernames[i], firstnames[i], lastnames[i], roles[i], passwords[i], locales[i])
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

func compareDBUser(t *testing.T, name string, act, exp db.User) {
	if exp.Username != "" {
		assert.Equalf(t, exp.Username, act.Username, "%s.UserId should be same", name)
	}
	assert.Equalf(t, exp.FirstName, act.FirstName, "%s.FirstName should be same", name)
	assert.Equalf(t, exp.LastName, act.LastName, "%s.LastName should be same", name)
	assert.Equalf(t, exp.Locale, act.Locale, "%s.Locale should be same", name)
	if exp.Password != "" {
		assert.Equalf(t, exp.Password, act.Password, "%s.Password should be same", name)
	}
	assert.Equalf(t, exp.Role, act.Role, "%s.Role should be same", name)
}
