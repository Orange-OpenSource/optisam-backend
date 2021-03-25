// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repv1 "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/account-service/pkg/repository/v1/mock"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

func Test_accountServiceServer_CreateScope(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
	})
	type args struct {
		ctx context.Context
		req *v1.CreateScopeRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.CreateScopeResponse
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, repv1.ErrNoData)
				mockRepo.EXPECT().CreateScope(ctx, "France", "OFR", "admin@test.com").Times(1).Return(nil)
			},
			want: &v1.CreateScopeResponse{},
		},
		{
			name: "Failure - Scope already exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(&repv1.Scope{
					ScopeCode: "OFR",
					ScopeName: "France",
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "Failure - Can not fetch scopes",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "Failure - Unable to create scope",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, repv1.ErrNoData)
				mockRepo.EXPECT().CreateScope(ctx, "France", "OFR", "admin@test.com").Times(1).Return(errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "Failure - user is admin (not superadmin)",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "Failure - user is user (not superadmin)",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "user@test.com",
					Role:   "User",
				}),
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "Failure - claims does not exists",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewAccountServiceServer(rep)
			got, err := s.CreateScope(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.CreateScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.CreateScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_ListScopes(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
		Socpes: []string{"O1", "O2"},
	})
	type args struct {
		ctx context.Context
		req *v1.ListScopesRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListScopesResponse
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx: ctx,
				req: &v1.ListScopesRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{
					&repv1.Scope{
						ScopeCode:  "O1",
						ScopeName:  "India",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(10, 0),
						GroupNames: []string{"ROOT"},
					},
					&repv1.Scope{
						ScopeCode:  "O2",
						ScopeName:  "France",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(11, 0),
						GroupNames: []string{"ROOT", "INDIA"},
					},
				}, nil)
			},
			want: &v1.ListScopesResponse{
				Scopes: []*v1.Scope{
					&v1.Scope{
						ScopeCode:  "O1",
						ScopeName:  "India",
						CreatedBy:  "admin@test.com",
						CreatedOn:  &tspb.Timestamp{Seconds: 10},
						GroupNames: []string{"ROOT"},
					},
					&v1.Scope{
						ScopeCode:  "O2",
						ScopeName:  "France",
						CreatedBy:  "admin@test.com",
						CreatedOn:  &tspb.Timestamp{Seconds: 11},
						GroupNames: []string{"ROOT", "INDIA"},
					},
				},
			},
		},
		{
			name: "Success - Scope is not associated with any group",
			args: args{
				ctx: ctx,
				req: &v1.ListScopesRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{
					&repv1.Scope{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
					&repv1.Scope{
						ScopeCode: "O2",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(11, 0),
					},
				}, nil)
			},
			want: &v1.ListScopesResponse{
				Scopes: []*v1.Scope{
					&v1.Scope{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: &tspb.Timestamp{Seconds: 10},
					},
					&v1.Scope{
						ScopeCode: "O2",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: &tspb.Timestamp{Seconds: 11},
					},
				},
			},
		},
		{
			name: "Success - Empty list of scopes",
			args: args{
				ctx: ctx,
				req: &v1.ListScopesRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{}, nil)
			},
			want: &v1.ListScopesResponse{},
		},
		{
			name: "Failure - claims does not exists",
			args: args{
				ctx: context.Background(),
				req: &v1.ListScopesRequest{},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "Failure - Unable to List scopes",
			args: args{
				ctx: ctx,
				req: &v1.ListScopesRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "Success - Claims scopes are nill",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.ListScopesRequest{},
			},
			setup:   func() {},
			wantErr: false,
			want:    &v1.ListScopesResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewAccountServiceServer(rep)
			got, err := s.ListScopes(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ListScopes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.ListScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}
