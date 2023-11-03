package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/api/v1"
	repv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1/mock"
	equipment "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/thirdparty/equipment-service/pkg/api/v1"
	equipmentMock "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/thirdparty/equipment-service/pkg/api/v1/mock"
	met "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/thirdparty/metric-service/pkg/api/v1/mock"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

func Test_accountServiceServer_CreateScope(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	var equip equipment.EquipmentServiceClient
	var metric met.MetricServiceClient
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
			name: "Success1",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
					ScopeType: v1.ScopeType_GENERIC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, repv1.ErrNoData)
				mockRepo.EXPECT().CreateScope(ctx, "France", "OFR", "admin@test.com", "GENERIC").Times(1).Return(nil)
				mockEquipClient := equipmentMock.NewMockEquipmentServiceClient(mockCtrl)
				equip = mockEquipClient
				mockEquipClient.EXPECT().CreateGenericScopeEquipmentTypes(ctx, &equipment.CreateGenericScopeEquipmentTypesRequest{Scope: "OFR"}).Return(&equipment.CreateGenericScopeEquipmentTypesResponse{Success: true}, nil).Times(1)

				mockMetClient := metmock.NewMockMetricServiceClient(mockCtrl)
				metric = mockMetClient
				mockMetClient.EXPECT().CreateScopeMetric(ctx, &met.CreateScopeMetricRequest{Scope: "OFR"}).Times(1).Return(&met.CreateScopeMetricResponse{Success: true}, nil).AnyTimes()
				mockRepo.EXPECT().ListScopes(ctx, []string{"OFR"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}, nil)
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}).Times(1).Return(nil)
				// metObj.EXPECT().CreateScopeMetric(ctx1, &metric.CreateScopeMetricRequest{Scope: "OFR"}).Times(1).Return(&metric.ListMetricResponse{}, nil)
			},
			want: &v1.CreateScopeResponse{Success: true},
		},
		{
			name: "Success Non generic scope",
			args: args{
				ctx: ctx,
				req: &v1.CreateScopeRequest{
					ScopeCode: "OFR",
					ScopeName: "France",
					ScopeType: v1.ScopeType_SPECIFIC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, repv1.ErrNoData)
				mockRepo.EXPECT().CreateScope(ctx, "France", "OFR", "admin@test.com", "SPECIFIC").Times(1).Return(nil)
				mockMetClient := metmock.NewMockMetricServiceClient(mockCtrl)
				metric = mockMetClient
				mockMetClient.EXPECT().CreateScopeMetric(ctx, &met.CreateScopeMetricRequest{Scope: "OFR"}).Return(&met.CreateScopeMetricResponse{Success: true}, nil).AnyTimes()
				mockRepo.EXPECT().ListScopes(ctx, []string{"OFR"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}, nil)
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}).Times(1).Return(nil)
			},
			want: &v1.CreateScopeResponse{Success: true},
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
					ScopeType: v1.ScopeType_GENERIC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "OFR").Times(1).Return(nil, repv1.ErrNoData)
				mockRepo.EXPECT().CreateScope(ctx, "France", "OFR", "admin@test.com", "GENERIC").Times(1).Return(errors.New("Internal"))
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
			s := &accountServiceServer{
				accountRepo: rep,
				equipment:   equip,
				metric:      metric,
			}
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
				mockRepo.EXPECT().GetScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode:  "O1",
						ScopeName:  "India",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(10, 0),
						GroupNames: []string{"ROOT"},
					},
					{
						ScopeCode:  "O2",
						ScopeName:  "France",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(11, 0),
						GroupNames: []string{"ROOT", "INDIA"},
					},
				}, nil)
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{
					{
						ScopeCode:  "O1",
						ScopeName:  "India",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(10, 0),
						GroupNames: []string{"ROOT"},
					},
					{
						ScopeCode:  "O2",
						ScopeName:  "France",
						CreatedBy:  "admin@test.com",
						CreatedOn:  time.Unix(11, 0),
						GroupNames: []string{"ROOT", "INDIA"},
					},
				}).Return(nil).AnyTimes()
			},
			want: &v1.ListScopesResponse{
				Scopes: []*v1.Scope{
					{
						ScopeCode:  "O1",
						ScopeName:  "India",
						CreatedBy:  "admin@test.com",
						CreatedOn:  &tspb.Timestamp{Seconds: 10},
						GroupNames: []string{"ROOT"},
					},
					{
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
					{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
					{
						ScopeCode: "O2",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(11, 0),
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().GetScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
					{
						ScopeCode: "O2",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(11, 0),
					},
				}, nil)
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{
					{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
					{
						ScopeCode: "O2",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(11, 0),
					},
				}).Return(nil).AnyTimes()
			},
			want: &v1.ListScopesResponse{
				Scopes: []*v1.Scope{
					{
						ScopeCode: "O1",
						ScopeName: "India",
						CreatedBy: "admin@test.com",
						CreatedOn: &tspb.Timestamp{Seconds: 10},
					},
					{
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
				mockRepo.EXPECT().GetScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{}, nil)
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return([]*repv1.Scope{}, nil)
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{}).Times(1).Return(nil)
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
				mockRepo.EXPECT().GetScopes(ctx, []string{"O1", "O2"}).Times(1).Return(nil, errors.New("Internal"))
				mockRepo.EXPECT().ListScopes(ctx, []string{"O1", "O2"}).Times(1).Return(nil, errors.New("Internal"))
				// mockRepo.EXPECT().GetScopes(ctx, []string{"O1", "O2"}).Times(1).Return(nil, errors.New("Internal"))
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
			s := &accountServiceServer{
				accountRepo: rep,
			}
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

func Test_accountServiceServer_GetScope(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
		Socpes: []string{"scope1", "scope2"},
	})
	type args struct {
		ctx context.Context
		req *v1.GetScopeRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.Scope
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetScopeRequest{
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "scope1").Times(1).Return(&repv1.Scope{
					ScopeCode:  "scope1",
					ScopeName:  "India",
					CreatedBy:  "admin@test.com",
					CreatedOn:  time.Unix(10, 0),
					GroupNames: []string{"ROOT"},
				}, nil)
			},
			want: &v1.Scope{
				ScopeCode:  "scope1",
				ScopeName:  "India",
				CreatedBy:  "admin@test.com",
				CreatedOn:  &tspb.Timestamp{Seconds: 10},
				GroupNames: []string{"ROOT"},
			},
		},
		{
			name: "FAILURE - ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.GetScopeRequest{
					Scope: "scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.GetScopeRequest{
					Scope: "scope3",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - scope does not exist",
			args: args{
				ctx: ctx,
				req: &v1.GetScopeRequest{
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "scope1").Times(1).Return(nil, repv1.ErrNoData)
			},
			want:    &v1.Scope{},
			wantErr: true,
		},
		{
			name: "FAILURE - unable to fetch scope info",
			args: args{
				ctx: ctx,
				req: &v1.GetScopeRequest{
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ScopeByCode(ctx, "scope1").Times(1).Return(nil, errors.New("internal"))
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
			got, err := tt.s.GetScope(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.GetScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.GetScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_GetScopeLists(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	var equip equipment.EquipmentServiceClient
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
	})
	type args struct {
		ctx context.Context
		req *v1.GetScopeListRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ScopeListResponse
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx: ctx,
				req: &v1.GetScopeListRequest{
					Scopes: []string{"OFR"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetScopes(ctx, []string{"OFR"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}, nil)
				mockRepo.EXPECT().ListScopes(ctx, []string{"OFR"}).Times(1).Return([]*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().SetScope(ctx, []*repv1.Scope{
					{
						ScopeCode: "OFR",
						ScopeName: "France",
						CreatedBy: "admin@test.com",
						CreatedOn: time.Unix(10, 0),
					},
				}).Return(nil).AnyTimes()
			},
			want: &v1.ScopeListResponse{ScopeNames: []string{"France"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &accountServiceServer{
				accountRepo: rep,
				equipment:   equip,
			}
			got, err := s.GetScopeLists(tt.args.ctx, tt.args.req)
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
