package v1

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"
	repv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1/mock"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func Test_authServiceServer_Login(t *testing.T) {
	type args struct {
		ctx context.Context
		req *v1.LoginRequest
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), 11)
	if err != nil {
		t.Fatal(err)
	}

	var mockCtrl *gomock.Controller
	var rep repv1.Repository
	var cfg config.Config
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
						Password:     string(hash),
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(nil, "user1@test.com").Return(nil).Times(1)
			},
		},
		{name: "reset login count error",
			s: &AuthServiceServer{},
			args: args{
				req: &v1.LoginRequest{
					Username: "user1@test.com",
					Password: "secret",
				},
			},
			want:    nil,
			wantErr: true,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().UserInfo(nil, "user1@test.com").
					Return(&repv1.UserInfo{
						UserID:       "user1@test.com",
						Password:     string(hash),
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(gomock.Any(), "user1@test.com").Return(errors.New("some db error")).Times(1)
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
						Password:     string(hash),
						FailedLogins: 3,
					}, nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(ctx, "user1@test.com").
					Return(nil).Times(1)
			},
			wantErr: false,
			want: &v1.LoginResponse{
				UserID: "user1@test.com",
			},
		},
		// {name: "failure wrong pass",
		// 	s: &AuthServiceServer{},
		// 	args: args{
		// 		req: &v1.LoginRequest{
		// 			Username: "user1@test.com",
		// 			Password: "secret",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockDB := mock.NewMockRepository(mockCtrl)
		// 		rep = mockDB
		// 		var ctx context.Context
		// 		mockDB.EXPECT().UserInfo(ctx, "user1@test.com").
		// 			Return(&repv1.UserInfo{
		// 				UserID:       "user1@test.com",
		// 				Password:     string("abc"),
		// 				FailedLogins: 3,
		// 			}, nil).Times(1)
		// 		mockDB.EXPECT().IncreaseFailedLoginCount(ctx, "user1@test.com").Return(nil).Times(1)
		// 		mockDB.EXPECT().ResetLoginCount(ctx, "user1@test.com").Return(nil).Times(1).AnyTimes()
		// 	},
		// 	wantErr: true,
		// },
		{name: "failure wrong pass update",
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
						Password:     string("abc"),
						FailedLogins: 3,
					}, nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(ctx, "user1@test.com").Return(errors.New("some db error")).Times(1)
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
						Password:     string(hash),
						FailedLogins: 3,
					}, nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(ctx, "user1@test.com").Return(nil).Times(1).AnyTimes()
			},
			wantErr: false,
			want:    &v1.LoginResponse{UserID: "user1@test.com"},
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
						Password:     string(hash),
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(nil, "user1@test.com").Return(nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(nil, "user1@test.com").Return(nil).Times(1).AnyTimes()
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
						Password:     string(hash),
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(ctx, "user1@test.com").Return(errors.New("test error")).Times(1)
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
						Password:     string(hash),
						FailedLogins: 2,
					}, nil).Times(1)
				mockDB.EXPECT().IncreaseFailedLoginCount(nil, "user1@test.com").Return(errors.New("db error")).Times(1)
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
						Password:     string(hash),
						FailedLogins: 0,
					}, nil).Times(1)
				mockDB.EXPECT().ResetLoginCount(ctx, "user1@test.com").Return(errors.New("some err")).Times(1).AnyTimes()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = NewAuthServiceServer(rep, cfg, nil, nil)
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
					{
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
					{
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
					{
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
					{
						ID:     2,
						Scopes: []string{"A", "B"},
					},
					{
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
			var cfg config.Config
			tt.s = NewAuthServiceServer(rep, cfg, nil, nil)
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
