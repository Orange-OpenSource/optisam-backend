// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"database/sql"
	"errors"
	v1 "optisam-backend/auth-service/pkg/api/v1"
	repv1 "optisam-backend/auth-service/pkg/repository/v1"
	"optisam-backend/auth-service/pkg/repository/v1/mock"
	"optisam-backend/common/optisam/token/claims"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_authServiceServer_Login(t *testing.T) {
	type args struct {
		ctx context.Context
		req *v1.LoginRequest
	}

	var mockCtrl *gomock.Controller
	var rep repv1.Repository
	tests := []struct {
		name    string
		s       *AuthServiceServer
		args    args
		want    *v1.LoginResponse
		setup   func()
		wantErr bool
	}{
		{name: "success",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			want: &v1.LoginResponse{
				UserID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","secret").Return(true,nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(nil, "user1@test.com").Return(nil).Times(1)
			},
		},
		{name: "failure user not found",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(nil, sql.ErrNoRows).Times(1)
				// mockDB.EXPECT().ResetLoginCount(ctx,"user1@test.com").
				// Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure getting user info unknown error",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(nil, errors.New("test error")).Times(1)
				// mockDB.EXPECT().ResetLoginCount(ctx,"user1@test.com").
				// Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure user is blocked",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 3,
					}, nil).Times(1)
				// mockDB.EXPECT().ResetLoginCount(ctx,"user1@test.com").
				// Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure user is blocked",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 3,
					}, nil).Times(1)
				// mockDB.EXPECT().ResetLoginCount(ctx,"user1@test.com").
				// Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure - Login - failed to check password",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","secret").Return(false,errors.New("failed to check password")).Times(1)
			},
			wantErr:true,
		},
		{name: "failure - Login - wrong password",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "abc",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","abc").Return(false,nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(nil, "user1@test.com").Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure password is wrong failure in increasing reset count",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "wrong password",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","wrong password").Return(false,nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(ctx, "user1@test.com").
					Return(errors.New("test error")).Times(1)
			},
			wantErr: true,
		},
		{name: "failure - Login - failed logins is equal to 2",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "abc",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 2,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","abc").Return(false,nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(nil, "user1@test.com").Return(nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure successful login but failure in resetting login count",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				var ctx context.Context
				mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().CheckPassword(nil,"user1@test.com","secret").Return(true,nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(ctx, "user1@test.com").
					Return(errors.New("test error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = NewAuthServiceServer(rep)
			got, err := tt.s.Login(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				if !assert.Errorf(t, err, "AuthServiceServer.Login() - expexted error got nil") {
					return
				}
			} else {
				if !assert.Empty(t, err) {
					return
				}
			}

			if !assert.Equal(t, tt.want, got) {
				return
			}
			mockCtrl.Finish()
		})
	}
}

func Test_authServiceServer_UserClaims(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Repository
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		s       *AuthServiceServer
		args    args
		setup   func()
		want    *claims.Claims
		wantErr bool
	}{
		{name: "SUCCESS - User role",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "User",
				}, nil).Times(1)
				mockDB.EXPECT().UserOwnedGroupsDirect(nil, "user1@test.com").Return([]*repv1.Group{
					&repv1.Group{
						ID:     2,
						Scopes: []string{"A", "B"},
					},
				}, nil).Times(1)
			},
			want: &claims.Claims{
				UserID: "user1@test.com",
				Role:   "User",
				Socpes: []string{"A", "B"},
			},
		},
		{name: "SUCCESS - Admin role",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "Admin",
				}, nil).Times(1)
				mockDB.EXPECT().UserOwnedGroupsDirect(nil, "user1@test.com").Return([]*repv1.Group{
					&repv1.Group{
						ID:     2,
						Scopes: []string{"A", "B"},
					},
				}, nil).Times(1)
			},
			want: &claims.Claims{
				UserID: "user1@test.com",
				Role:   "Admin",
				Socpes: []string{"A", "B"},
			},
		},
		{name: "SUCCESS - SuperAdmin role",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "SuperAdmin",
				}, nil).Times(1)
				mockDB.EXPECT().UserOwnedGroupsDirect(nil, "user1@test.com").Return([]*repv1.Group{
					&repv1.Group{
						ID:     2,
						Scopes: []string{"A", "B"},
					},
				}, nil).Times(1)
			},
			want: &claims.Claims{
				UserID: "user1@test.com",
				Role:   "SuperAdmin",
				Socpes: []string{"A", "B"},
			},
		},
		{name: "FAILURE - user not found",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(nil, errors.New("")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - translateRole - unknown userRole from database",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "Abc",
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - UserClaims - cannot fetch user info",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "User",
				}, nil).Times(1)
				mockDB.EXPECT().UserOwnedGroupsDirect(nil, "user1@test.com").Return(nil, errors.New("")).Times(1)
			},
			wantErr: true,
		},
		{name: "SUCCESS - elementExists - true ",
			s: &AuthServiceServer{},
			args: args{
				userID: "user1@test.com",
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").Return(&repv1.UserInfo{
					UserID: "user1@test.com",
					Role:   "User",
				}, nil).Times(1)
				mockDB.EXPECT().UserOwnedGroupsDirect(nil, "user1@test.com").Return([]*repv1.Group{
					&repv1.Group{
						ID:     2,
						Scopes: []string{"A", "B"},
					},
					&repv1.Group{
						ID:     2,
						Scopes: []string{"B", "C"},
					},
				}, nil).Times(1)
			},
			want: &claims.Claims{
				UserID: "user1@test.com",
				Role:   "User",
				Socpes: []string{"A", "B", "C"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = NewAuthServiceServer(rep)
			got, err := tt.s.UserClaims(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServiceServer.UserClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				compareClaims(t, "UserClaims", got, tt.want)
			}
		})
	}
}

func compareClaims(t *testing.T, name string, exp *claims.Claims, act *claims.Claims) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.UserID, act.UserID, "%s UserID are not same", name)
	assert.Equalf(t, exp.Role, act.Role, "%s Role are not same", name)
	for i := range exp.Socpes {
		for j := range act.Socpes {
			assert.Equalf(t, exp.Socpes[i], act.Socpes[j], "%s Scopes are not same", name)
		}
	}
}
