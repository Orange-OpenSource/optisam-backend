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

func Test_metricServiceServer_CreateMetricStaticStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricSS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricSS
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
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

				mockRepo.EXPECT().CreateMetricStaticStandard(ctx, &repo.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
				}, "Scope1").Return(&repo.MetricSS{
					ID:             "Met_SS1ID",
					Name:           "Met_SS1",
					ReferenceValue: 1,
				}, nil).Times(1)
			},
			output: &v1.MetricSS{
				ID:             "Met_SS1ID",
				Name:           "Met_SS1",
				ReferenceValue: 1,
			},
		},
		{name: "FAILURE - CreateMetricStaticStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricStaticStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricStaticStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
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
						Name: "Met_SS1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricStaticStandard - cannot create metric ss",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
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
				mockRepo.EXPECT().CreateMetricStaticStandard(ctx, &repo.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
				}, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricStaticStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricStaticStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricstaticStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricSS(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricSS
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
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSS(ctx, "Met_SS1", "Scope1").Return(&repo.MetricSS{
					ID:             "123",
					Name:           "Met_SS1",
					ReferenceValue: int32(10),
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricSS(ctx, &repo.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricSS - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricSS - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope3"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricSS - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSS(ctx, "Met_SS1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricSS - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSS(ctx, "Met_SS1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},

		{name: "FAILURE - UpdateMetricSS - cannot update metric ss",
			input: args{
				ctx: ctx,
				req: &v1.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSS(ctx, "Met_SS1", "Scope1").Return(&repo.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 5,
				}, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricSS(ctx, &repo.MetricSS{
					Name:           "Met_SS1",
					ReferenceValue: 1,
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
			got, err := s.UpdateMetricStaticStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.UpdateMetricStaticStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricStaticStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
