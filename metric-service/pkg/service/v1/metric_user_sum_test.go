package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1"
	accmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1/mock"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
)

func Test_metricServiceServer_CreateMetricUserSumStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	var acc accv1.AccountServiceClient
	type args struct {
		ctx context.Context
		req *v1.MetricUSS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricUSS
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().CreateMetricUSS(ctx, &repo.MetricUSS{
					Name: "Met_USS1",
				}, "Scope1").Return(&repo.MetricUSS{
					ID:   "Met_USS1ID",
					Name: "Met_USS1",
				}, nil).Times(1)
			},
			output: &v1.MetricUSS{
				ID:   "Met_USS1ID",
				Name: "Met_USS1",
			},
		},
		{name: "FAILURE - CreateMetricUserSumStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - account/GetScope -  can not fetch scope type",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(nil, errors.New("internal"))
			},
			outErr: true,
		},
		{name: "FAILURE - Scope type specific",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_SPECIFIC.String(),
				}, nil)
			},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricUserSumStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricUserSumStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "Met_USS1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricUserSumStandard - cannot create metric uss",
			input: args{
				ctx: ctx,
				req: &v1.MetricUSS{
					Name:   "Met_USS1",
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().CreateMetricUSS(ctx, &repo.MetricUSS{
					Name: "Met_USS1",
				}, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &metricServiceServer{
				metricRepo: rep,
				account:    acc,
			}
			got, err := s.CreateMetricUserSumStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricUserSumStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricUserSumStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}
