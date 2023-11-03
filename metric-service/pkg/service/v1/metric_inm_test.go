package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1/mock"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
)

func Test_metricServiceServer_CreateMetricInstanceNumberStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricINM
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricINM
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
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

				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1,
				}, "Scope1").Return(&repo.MetricINM{
					ID:          "Met_INM1ID",
					Name:        "Met_INM1",
					Coefficient: 1,
				}, nil).Times(1)
			},
			output: &v1.MetricINM{
				ID:               "Met_INM1ID",
				Name:             "Met_INM1",
				NumOfDeployments: 1,
			},
		},
		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricInstanceNumberStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
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
						Name: "Met_INM1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot create metric acs",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
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
				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1,
				}, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricInstanceNumberStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricInstanceNumberStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricInstanceNumberStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricINM(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricINM
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
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "Met_INM1", "Scope1").Return(&repo.MetricINM{
					ID:          "123",
					Name:        "Met_INM1",
					Coefficient: int32(10),
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricINM(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1,
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricINM - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricINM - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope3"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricINM - Default Value True, Metric created by import can't be updated error",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
					Default:          true,
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricINM - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "Met_INM1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricINM - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "Met_INM1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},

		{name: "FAILURE - UpdateMetricINM - cannot update metric inm",
			input: args{
				ctx: ctx,
				req: &v1.MetricINM{
					Name:             "Met_INM1",
					NumOfDeployments: 1,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "Met_INM1", "Scope1").Return(&repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 5,
				}, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricINM(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1,
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
			got, err := s.UpdateMetricInstanceNumberStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.UpdateMetricInstanceNumberStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricInstanceNumberStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
