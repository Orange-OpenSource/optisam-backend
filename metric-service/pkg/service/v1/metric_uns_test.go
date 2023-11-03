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

func Test_metricServiceServer_CreateMetricUserNominativeStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricUNS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricUNS
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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

				mockRepo.EXPECT().CreateMetricUserNominativeStandard(ctx, &repo.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
				}, "Scope1").Return(&repo.MetricUNS{
					ID:      "Met_UNS1ID",
					Name:    "Met_UNS1",
					Profile: "P1",
				}, nil).Times(1)
			},
			output: &v1.MetricUNS{
				ID:      "Met_UNS1ID",
				Name:    "Met_UNS1",
				Profile: "P1",
			},
		},
		{name: "FAILURE - CreateMetricUserNominativeStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricUserNominativeStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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
		{name: "FAILURE - CreateMetricUserNominativeStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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
						Name: "Met_UNS1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricUserNominativeStandard - cannot create metric UNS",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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
				mockRepo.EXPECT().CreateMetricUserNominativeStandard(ctx, &repo.MetricUNS{
					Name:    "Met_UNS1",
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
			got, err := s.CreateMetricUserNominativeStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricUserNominativeStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricUserNominativeStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricUNS(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricUNS
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
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUNS(ctx, "Met_UNS1", "Scope1").Return(&repo.MetricUNS{
					ID:      "123",
					Name:    "Met_UNS1",
					Profile: "P10",
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricUNS(ctx, &repo.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricUNS - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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
		{name: "FAILURE - UpdateMetricUNS - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
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
		{name: "FAILURE - UpdateMetricUNS - Default Value True, Metric created by import can't be updated error",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
					Default: true,
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricUNS - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUNS(ctx, "Met_UNS1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},
		{name: "FAILURE - UpdateMetricUNS - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUNS(ctx, "Met_UNS1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			outErr: true,
		},

		{name: "FAILURE - UpdateMetricUNS - cannot update metric inm",
			input: args{
				ctx: ctx,
				req: &v1.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P1",
					Scopes:  []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUNS(ctx, "Met_UNS1", "Scope1").Return(&repo.MetricUNS{
					Name:    "Met_UNS1",
					Profile: "P5",
				}, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricUNS(ctx, &repo.MetricUNS{
					Name:    "Met_UNS1",
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
			got, err := s.UpdateMetricUserNominativeStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.UpdateMetricUserNominativeStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricUserNominativeStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestGetDescriptionUNS(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockMetricRepo := mock.NewMockMetric(mockCtrl)
	s := &metricServiceServer{
		metricRepo: mockMetricRepo,
	}

	tests := []struct {
		name          string
		metricName    string
		scope         string
		metric        string
		expected      string
		expectedError error
	}{
		{
			name:          "Success",
			metricName:    "MetricName",
			scope:         "Scope",
			metric:        "Profile1",
			expected:      "ExpectedDescription",
			expectedError: nil,
		},
		{
			name:          "Error",
			metricName:    "MetricName",
			scope:         "Scope",
			metric:        "",
			expected:      "",
			expectedError: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetricRepo.EXPECT().GetMetricConfigUNS(ctx, tt.metricName, tt.scope).Return(&repo.MetricUNS{}, tt.expectedError)

			_, err := s.getDescriptionUNS(ctx, tt.metricName, tt.scope)
			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("metricServiceServer.CreateMetricWindowServerStandard() error = %v, wantErr %v", err, tt.expected)
				return
			}
		})
	}
}
