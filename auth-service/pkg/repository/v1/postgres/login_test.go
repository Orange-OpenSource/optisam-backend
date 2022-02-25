package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	v1 "optisam-backend/auth-service/pkg/repository/v1"
	"reflect"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Default_UserInfo(t *testing.T) {
	//	var db *sql.DB
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		d       *Default
		args    args
		want    *v1.UserInfo
		setup   func() (func() error, error)
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx:    context.Background(),
				userID: "user1@test.com",
			},
			want: &v1.UserInfo{
				UserID:       "user1@test.com",
				Password:     "supersecret1",
				Locale:       "en",
				Role:         "Admin",
				FailedLogins: 0,
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
		},
		{name: "failure",
			args: args{
				ctx:    context.Background(),
				userID: "user3@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
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
			if !assert.Empty(t, err) {
				return
			}
			defer func() {
				require.Empty(t, cleanup())
			}()
			tt.d = NewRepository(db)
			got, err := tt.d.UserInfo(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Default.UserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Default.UserInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Default_IncreaseFailedLoginCount(t *testing.T) {
	//	var db *sql.DB
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		d       *Default
		args    args
		setup   func() (func() error, error)
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx:    context.Background(),
				userID: "user1@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
		},
		{name: "failure",
			args: args{
				ctx:    context.Background(),
				userID: "user3@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
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
			if !assert.Empty(t, err) {
				return
			}
			defer func() {
				require.Empty(t, cleanup())
			}()
			tt.d = NewRepository(db)
			err = tt.d.IncreaseFailedLoginCount(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Default.IncreaseFailedLoginCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_Default_ResetLoginCount(t *testing.T) {
	//	var db *sql.DB
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		d       *Default
		args    args
		setup   func() (func() error, error)
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx:    context.Background(),
				userID: "user1@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
					return nil, err
				}
				return func() error {
					return deleteAllUsers(db, usernames)
				}, nil
			},
		},
		{name: "failure",
			args: args{
				ctx:    context.Background(),
				userID: "user3@test.com",
			},
			setup: func() (func() error, error) {
				usernames := []string{"user1@test.com", "user2@test.com"}
				passwords := []string{"supersecret1", "supersecret2"}
				if err := createUsers(db, usernames, passwords); err != nil {
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
			if !assert.Empty(t, err) {
				return
			}
			defer func() {
				require.Empty(t, cleanup())
			}()
			tt.d = NewRepository(db)
			err = tt.d.ResetLoginCount(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Default.ResetLoginCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func createUsers(db *sql.DB, usernames, passwords []string) error {
	query := "INSERT INTO users(username,password,first_name,last_name,role,locale) VALUES "
	args := []interface{}{}
	for i := range usernames {
		query += fmt.Sprintf("($%v,$%v,'super','sdmin','Admin','en')", 2*i+1, 2*i+2)
		args = append(args, usernames[i], passwords[i])
		if i != len(usernames)-1 {
			query += ","
		}
	}

	_, err := db.Exec(query, args...)
	return err
}

func deleteAllUsers(db *sql.DB, users []string) error {
	result, err := db.Exec("DELETE FROM users WHERE username = ANY ($1)", pq.Array(users))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != int64(len(users)) {
		return errors.New("2 rows should be deleted")
	}
	return err
}

// func TestDefault_CheckPassword(t *testing.T) {
// 	type args struct {
// 		ctx      context.Context
// 		userID   string
// 		password string
// 	}
// 	tests := []struct {
// 		name    string
// 		d       *Default
// 		args    args
// 		setup   func() (func() error, error)
// 		want    bool
// 		wantErr bool
// 	}{
// 		{name: "success",
// 			args: args{
// 				ctx:      context.Background(),
// 				userID:   "user1@test.com",
// 				password: "secret",
// 			},
// 			setup: func() (func() error, error) {
// 				if err := createUsers(db, []string{"user1@test.com"}, []string{"secret"}); err != nil {
// 					return nil, err
// 				}
// 				return func() error {
// 					return deleteAllUsers(db, []string{"user1@test.com"})
// 				}, nil
// 			},
// 			want: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cleanup, err := tt.setup()
// 			if !assert.Empty(t, err) {
// 				return
// 			}
// 			defer func() {
// 				require.Empty(t, cleanup())
// 			}()
// 			tt.d = NewRepository(db)
// 			got, err := tt.d.CheckPassword(tt.args.ctx, tt.args.userID, tt.args.password)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Default.CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("Default.CheckPassword() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
