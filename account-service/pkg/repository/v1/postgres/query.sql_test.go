// Code generated by sqlc. DO NOT EDIT.

// source: query.sql

package postgres

import (
	"context"
	"fmt"
	v1 "optisam-backend/account-service/pkg/repository/v1"
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
		r       *AccountRepository
		args    args
		setup   func() (func() error, string, error)
		verify  func(a *AccountRepository, userID string) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewAccountRepository(db),
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
				if err := createUsers(db, usernames, firstnames, lastnames, roles, passwords, locales); err != nil {
					return nil, "", err
				}
				return func() error {
					err := deleteAllUsers(db, []string{"user1@test.com"})
					if err != nil {
						return err
					}
					return nil
				}, "admin1@test.com", nil
			},
			verify: func(a *AccountRepository, userID string) error {
				_, err := a.AccountInfo(context.Background(), userID)
				if err != nil {
					if err == v1.ErrNoData {
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
				assert.Empty(t, tt.verify(tt.r, userID))
			}
		})
	}
}

func TestQueries_InsertUserAudit(t *testing.T) {
	type args struct {
		ctx context.Context
		arg InsertUserAuditParams
	}
	tests := []struct {
		name    string
		q       *Queries
		args    args
		cleanup func(userID string) error
		verify  func(userID string) error
		wantErr bool
	}{
		{name: "SUCCESS",
			q: New(db),
			args: args{
				ctx: context.Background(),
				arg: InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            "Admin",
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				},
			},
			cleanup: func(userID string) error {
				query := "DELETE FROM users_audit WHERE username=$1"
				_, err := db.ExecContext(context.Background(), query, userID)
				return err
			},
			verify: func(userID string) error {
				query := "SELECT username from users_audit WHERE username=$1"
				userAudit := UsersAudit{}
				if err := db.QueryRowContext(context.Background(), query, userID).Scan(&userAudit.Username); err != nil {
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
