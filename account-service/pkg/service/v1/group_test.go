package v1

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/account-service/pkg/repository/v1/mock"
	"reflect"
	"testing"

	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_accountServiceServer_ListGroups(t *testing.T) {

	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account

	type args struct {
		ctx context.Context
		req *v1.ListGroupsRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		want    *v1.ListGroupsResponse
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListGroupsRequest{},
			},
			want: &v1.ListGroupsResponse{
				NumOfRecords: 2,
				Groups: []*v1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
						ParentId:           1,
					},
					{
						ID:                 3,
						Name:               "OFS",
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
						ParentId:           1,
					},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, []*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},

		{name: "FAILURE",
			args: args{
				ctx: ctx,
				req: &v1.ListGroupsRequest{},
			},
			want: &v1.ListGroupsResponse{
				NumOfRecords: 2,
				Groups: []*v1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
						ParentId:           1,
					},
					{
						ID:                 3,
						Name:               "OFS",
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
						ParentId:           1,
					},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, nil, fmt.Errorf("Test error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListGroupsRequest{},
			},
			want: &v1.ListGroupsResponse{
				NumOfRecords: 2,
				Groups: []*v1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
						ParentId:           1,
					},
					{
						ID:                 3,
						Name:               "OFS",
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
						ParentId:           1,
					},
				},
			},
			mock: func() {

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.ListGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ListGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareGroups(t, "ListGroupsResponse", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_CreateGroup(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account

	type args struct {
		ctx context.Context
		req *v1.Group
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		want    *v1.Group
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			want: &v1.Group{
				ID:                 4,
				Name:               "OLS",
				FullyQualifiedName: "Orange.OBS.OLS",
				Scopes:             []string{"B", "C", "D"},
				ParentId:           2,
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
				mockRepo.EXPECT().CreateGroup(ctx, "admin@superuser.com", &repo.Group{
					Name:               "OLS",
					ParentID:           2,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"B", "C", "D"},
				}).Return(&repo.Group{
					ID:                 4,
					Name:               "OLS",
					ParentID:           2,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"B", "C", "D"},
				}, nil).Times(1)
			},
			wantErr: false,
		},

		{name: "FAILURE - parent does not exist",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 5,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},

		{name: "FAILURE - fully qualified name already exists ",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "ONS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 1,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(3, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
					{
						ID:                 1,
						Name:               "Orange",
						ParentID:           0,
						FullyQualifiedName: "Orange",
						Scopes:             []string{"A", "B", "C", "D", "E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - failed to fetch parent group ",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "ONS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(3, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
					{
						ID:                 1,
						Name:               "Orange",
						ParentID:           0,
						FullyQualifiedName: "Orange",
						Scopes:             []string{"A", "B", "C", "D", "E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(nil, errors.New("failed to return group")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent and child can not have same name ",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OBS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(3, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
					{
						ID:                 1,
						Name:               "Orange",
						ParentID:           0,
						FullyQualifiedName: "Orange",
						Scopes:             []string{"A", "B", "C", "D", "E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - fully qualified name already exists - case are different ",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "ons",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 1,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(3, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
					{
						ID:                 1,
						Name:               "Orange",
						ParentID:           0,
						FullyQualifiedName: "Orange",
						Scopes:             []string{"A", "B", "C", "D", "E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
			},
			wantErr: true,
		},

		{name: "FAILURE - scope is not subset of parent scope",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"E", "C", "D"},
					ParentId: 2,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
			},
			wantErr: true,
		},

		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock:    func() {},
			wantErr: true,
		},

		{name: "FAILURE - create group not successful",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(2, []*repo.Group{
					{
						ID:                 2,
						Name:               "OBS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS",
						Scopes:             []string{"A", "B", "C", "D"},
					},
					{
						ID:                 3,
						Name:               "ONS",
						ParentID:           1,
						FullyQualifiedName: "Orange.ONS",
						Scopes:             []string{"E", "F"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					Name: "OBS",
				}, nil).Times(1)
				mockRepo.EXPECT().CreateGroup(ctx, "admin@superuser.com", &repo.Group{
					Name:               "OLS",
					ParentID:           2,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"B", "C", "D"},
				}).Return(nil, fmt.Errorf("Test error")).Times(1)
			},
			wantErr: true,
		},

		{name: "FAILURE - failed to fetch user owner groups",
			args: args{
				ctx: ctx,
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				var queryParams *repo.GroupQueryParams
				mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", queryParams).Return(0, nil, fmt.Errorf("Test error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - permission denied - user do not have access to create name",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
				}),
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - permission denied - unknown role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Abc",
				}),
				req: &v1.Group{
					Name:     "OLS",
					Scopes:   []string{"B", "C", "D"},
					ParentId: 2,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.CreateGroup(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.CreateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareGroup(t, "Created Group", got, tt.want)
			}
			if tt.mock == nil {
				defer mockCtrl.Finish()
			}
		})
	}
}

func Test_accountServiceServer_ListUserGroups(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account

	type args struct {
		ctx context.Context
		req *v1.ListGroupsRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListGroupsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListGroupsRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return([]*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
			},
			want: &v1.ListGroupsResponse{
				Groups: []*v1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
						ParentId:           1,
					},
					{
						ID:                 3,
						Name:               "OFS",
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
						ParentId:           1,
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - cannot retrive claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListGroupsRequest{},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot get Groups",
			args: args{
				ctx: ctx,
				req: &v1.ListGroupsRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return(nil, errors.New("failed to get Groups")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.ListUserGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ListUserGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				groupsCompare(t, "ListGroupsResponse", got, tt.want)
			}

		})
	}

}

func Test_accountServiceServer_ListChildGroups(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account

	type args struct {
		ctx context.Context
		req *v1.ListChildGroupsRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListGroupsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{FullyQualifiedName: "Orange.OBS.OLS.OFS"}, nil).Times(1)
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return([]*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().ChildGroupsDirect(ctx, int64(1), nil).Return([]*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
			},
			want: &v1.ListGroupsResponse{
				Groups: []*v1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
						ParentId:           1,
					},
					{
						ID:                 3,
						Name:               "OFS",
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
						ParentId:           1,
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - cannot retrive claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - failed to get group",
			args: args{
				ctx: ctx,
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(nil, errors.New("failed to get Group")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - failed to get groups",
			args: args{
				ctx: ctx,
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{FullyQualifiedName: "Orange.OBS.OLS.OFS"}, nil).Times(1)
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return(nil, errors.New("failed to get groups")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - permission denied",
			args: args{
				ctx: context.Background(),
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{FullyQualifiedName: "Orange.OBS.OLS.OFS"}, nil).Times(1)
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return([]*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "a",
						Scopes:             []string{"A", "B"},
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - failed to get child groups",
			args: args{
				ctx: ctx,
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{FullyQualifiedName: "Orange.OBS.OLS.OFS"}, nil).Times(1)
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return([]*repo.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().ChildGroupsDirect(ctx, int64(1), nil).Return(nil, errors.New("failed to get groups")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - user doesnot have access for the group",
			args: args{
				ctx: ctx,
				req: &v1.ListChildGroupsRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(1)).Return(&repo.Group{FullyQualifiedName: "Orange.OBS.OLS.OFS"}, nil).Times(1)
				mockRepo.EXPECT().UserOwnedGroupsDirect(ctx, "admin@superuser.com", nil).Return([]*repo.Group{
					{
						FullyQualifiedName: "A",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.ListChildGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ListChildGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				groupsCompare(t, "ListGroupsResponse", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_UpdateGroup(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account
	type args struct {
		ctx context.Context
		req *v1.UpdateGroupRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.Group
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupExistsByFQN(ctx, "Orange.OBS.OFS").Return(false, nil).Times(1)
				mockRepo.EXPECT().UpdateGroup(ctx, int64(2), &repo.GroupUpdate{
					Name: "OFS",
				}).Return(nil).Times(1)
			},
			want: &v1.Group{
				ID:                 2,
				Name:               "OFS",
				ParentId:           1,
				FullyQualifiedName: "Orange.OBS.OFS",
				Scopes:             []string{"A", "B"},
			},
		},
		{name: "FAILURE - UpdateGroup - cannot retrive claims",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - invalid request - GroupId can not be 0",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: 0,
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - invalid request - Group can not be nil",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: 2,
					Group:   nil,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - invalid request - Group name can not be empty",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: 2,
					Group: &v1.UpdateGroup{
						Name: "",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - permission denied - user do not have access to update name",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
				}),
				req: &v1.UpdateGroupRequest{
					GroupId: 2,
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - failed to get group",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(nil, errors.New("failed to get groups")).Times(1)
			},
			wantErr: true,
		},
		{name: "SUCCESS - UpdateGroup - no change in name",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OFS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OFS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
			},
			want: &v1.Group{
				ID:                 2,
				Name:               "OFS",
				ParentId:           1,
				FullyQualifiedName: "Orange.OBS.OFS",
				Scopes:             []string{"A", "B"},
			},
		},
		{name: "FAILURE - UpdateGroup - failed to check GroupExistsByFQN",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupExistsByFQN(ctx, "Orange.OBS.OFS").Return(false, errors.New("")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - group name is not available",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupExistsByFQN(ctx, "Orange.OBS.OFS").Return(true, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - failed to update group",
			args: args{
				ctx: ctx,
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupExistsByFQN(ctx, "Orange.OBS.OFS").Return(false, nil).Times(1)
				mockRepo.EXPECT().UpdateGroup(ctx, int64(2), &repo.GroupUpdate{
					Name: "OFS",
				}).Return(errors.New("Test error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateGroup - unknown role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "ABC",
				}),
				req: &v1.UpdateGroupRequest{
					GroupId: 2,
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "SUCCESS - UpdateGroup - Admin userRole",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateGroupRequest{
					GroupId: int64(2),
					Group: &v1.UpdateGroup{
						Name: "OFS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}), int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupExistsByFQN(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}), "Orange.OBS.OFS").Return(false, nil).Times(1)
				mockRepo.EXPECT().UpdateGroup(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}), int64(2), &repo.GroupUpdate{
					Name: "OFS",
				}).Return(nil).Times(1)
			},
			want: &v1.Group{
				ID:                 2,
				Name:               "OFS",
				ParentId:           1,
				FullyQualifiedName: "Orange.OBS.OFS",
				Scopes:             []string{"A", "B"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.UpdateGroup(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.UpdateGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				compareGroup(t, "UpdatedGroup", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_DeleteGroup(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account
	type args struct {
		ctx context.Context
		req *v1.DeleteGroupRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.DeleteGroupResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
					NumberOfGroups:     0,
					NumberOfUsers:      0,
				}, nil).Times(1)
				mockRepo.EXPECT().DeleteGroup(ctx, int64(2)).Return(nil).Times(1)
			},
			want: &v1.DeleteGroupResponse{
				Success: true,
			},
		},
		{name: "FAILURE - DeleteGroup - cannot retrive claims",
			args: args{
				ctx: context.Background(),
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - invalid request - GroupId can not be 0",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: 0,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - permission denied - user do not have access to update name",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
				}),
				req: &v1.DeleteGroupRequest{
					GroupId: 2,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - failed to get group",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: int64(2),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(nil, errors.New("failed to get groups")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - permission denied - group contains users",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: int64(2),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
					NumberOfGroups:     0,
					NumberOfUsers:      1,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - permission denied - group contains child groups",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: int64(2),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
					NumberOfGroups:     1,
					NumberOfUsers:      0,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - failed to delete group",
			args: args{
				ctx: ctx,
				req: &v1.DeleteGroupRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(ctx, int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OLS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OLS",
					Scopes:             []string{"A", "B"},
					NumberOfGroups:     0,
					NumberOfUsers:      0,
				}, nil).Times(1)
				mockRepo.EXPECT().DeleteGroup(ctx, int64(2)).Return(errors.New("")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteGroup - unknown user role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "ABC",
				}),
				req: &v1.DeleteGroupRequest{
					GroupId: 2,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "SUCCESS - Admin UserRole",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}),
				req: &v1.DeleteGroupRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GroupInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}), int64(2)).Return(&repo.Group{
					ID:                 2,
					Name:               "OFS",
					ParentID:           1,
					FullyQualifiedName: "Orange.OBS.OFS",
					Scopes:             []string{"A", "B"},
					NumberOfGroups:     0,
					NumberOfUsers:      0,
				}, nil).Times(1)
				mockRepo.EXPECT().DeleteGroup(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "Admin",
				}), int64(2)).Return(nil).Times(1)
			},
			want: &v1.DeleteGroupResponse{
				Success: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.DeleteGroup(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.DeleteGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.DeleteGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareGroup(t *testing.T, name string, exp *v1.Group, act *v1.Group) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.ID, act.ID, "%d.ID are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.FullyQualifiedName, act.FullyQualifiedName, "%s.FullyQualified Name are not same", name)
	assert.ElementsMatchf(t, exp.Scopes, act.Scopes, "%s.Scopes are not same", name)
	assert.Equalf(t, exp.ParentId, act.ParentId, "%d.ParentId are not same", name)
}

func compareGroups(t *testing.T, name string, exp *v1.ListGroupsResponse, act *v1.ListGroupsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.NumOfRecords, act.NumOfRecords, "%s.Records are not same", name)
	compareAllGroups(t, name+".Group", exp.Groups, act.Groups)
}

func compareAllGroups(t *testing.T, name string, exp []*v1.Group, act []*v1.Group) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareGroup(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func groupsCompare(t *testing.T, name string, exp *v1.ListGroupsResponse, act *v1.ListGroupsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	compareAllGroups(t, name+".Group", exp.Groups, act.Groups)
}

func Test_accountServiceServer_ListComplienceGroups(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   claims.RoleSuperAdmin,
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repo.Account

	type args struct {
		ctx context.Context
		req *v1.ListGroupsRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListComplienceGroupsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListGroupsRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetComplienceGroups(ctx).Return([]repo.GetComplienceGroups{
					{
						ID:        2,
						Name:      "ROOT.Orange",
						ScopeCode: []string{"A", "B"},
						ScopeName: []string{"A", "B"},
					},
					{
						ID:        3,
						Name:      "ROOT.API",
						ScopeCode: []string{"C", "D"},
						ScopeName: []string{"C", "D"},
					},
				}, nil).Times(1)
			},
			want: &v1.ListComplienceGroupsResponse{
				ComplienceGroups: []*v1.ComplienceGroup{
					{
						GroupId:   2,
						Name:      "ROOT.Orange",
						ScopeCode: []string{"A", "B"},
						ScopeName: []string{"A", "B"},
					},
					{
						GroupId:   3,
						Name:      "ROOT.API",
						ScopeCode: []string{"C", "D"},
						ScopeName: []string{"C", "D"},
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - cannot retrive claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListGroupsRequest{},
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.ListComplienceGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ListUserGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.ListComplienceGroups() = %v, want %v", got, tt.want)
			}

		})
	}
}
