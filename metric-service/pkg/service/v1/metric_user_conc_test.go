package v1

import (
	"context"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"
	"optisam-backend/metric-service/pkg/repository/v1/mock"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_metricServiceServer_CreateMetricUserConcurentStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricUCS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricUCS
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().CreateMetricUserConcurentStandard(ctx, &repo.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
				}, "Scope1").Return(&repo.MetricUCS{
					ID:      "Met_UCS1ID",
					Name:    "Met_UCS1",
					Profile: "P1",
				}, nil).Times(1)
			},
			output: &v1.MetricUCS{
				ID:      "Met_UCS1ID",
				Name:    "Met_UCS1",
				Profile: "P1",
			},
		},
		{name: "FAILURE - CreateMetricUserConcurentStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricUserConcurentStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricUserConcurentStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "Met_UCS1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricUserConcurentStandard - cannot create metric UCS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().CreateMetricUserConcurentStandard(ctx, &repo.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
				}, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricUserConcurentStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricUserConcurentStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricUserConcurentStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricUCS(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricUCS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.UpdateMetricResponse
		outErr bool
	}{
		{
			name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigConcurentUser(ctx, "Met_UCS1", "Scope1").Return(&repo.MetricUCS{
					ID:      "123",
					Name:    "Met_UCS1",
					Profile: "P10",
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricUCS(ctx, &repo.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricUCS - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricUCS - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope3"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricUCS - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigConcurentUser(ctx, "Met_UCS1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricUCS - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigConcurentUser(ctx, "Met_UCS1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},

		{name: "FAILURE - UpdateMetricUCS - cannot update metric inm",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigConcurentUser(ctx, "Met_UCS1", "Scope1").Return(&repo.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P5",
				}, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricUCS(ctx, &repo.MetricUCS{
					Name:    "Met_UCS1",
					Profile: "P1",
				}, "Scope1").Return(errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.UpdateMetricUserConcurentStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.UpdateMetricUserConcurentStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricUserConcurentStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
