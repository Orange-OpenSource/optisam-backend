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
		{name: "FAILURE - UpdateMetricUCS - Default value true metric can't be updated",
			input: args{
				ctx: ctx,
				req: &v1.MetricUCS{
					Name:    "Met_UCS1",
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

func TestGetDescriptionUCS(t *testing.T) {
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
			mockMetricRepo.EXPECT().GetMetricConfigConcurentUser(ctx, tt.metricName, tt.scope).Return(&repo.MetricUCS{}, tt.expectedError)

			_, err := s.getDescriptionUCS(ctx, tt.metricName, tt.scope)
			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("metricServiceServer.CreateMetricWindowServerStandard() error = %v, wantErr %v", err, tt.expected)
				return
			}
		})
	}
}

// func Test_GetDescriptionUCS(t *testing.T) {
// 	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
// 		UserID: "admin@superuser.com",
// 		Role:   "SuperAdmin",
// 		Socpes: []string{"Scope1", "Scope2"},
// 	})

// 	var mockCtrl *gomock.Controller
// 	var rep repo.Metric

// 	tests := []struct {
// 		name    string
// 		s       *metricServiceServer
// 		input   *v1.DropMetricDataRequest
// 		setup   func()
// 		ctx     context.Context
// 		wantErr bool
// 	}{

// 		{
// 			name:  "DBError",
// 			input: &v1.DropMetricDataRequest{Scope: "Scope1"},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockMetric(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().DropMetrics(ctx, "Scope1").Return(errors.New("DBError")).Times(1)
// 			},
// 			ctx:     ctx,
// 			wantErr: true,
// 		},
// 		{
// 			name:  "SuccessFully",
// 			input: &v1.DropMetricDataRequest{Scope: "Scope1"},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockMetric(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().DropMetrics(ctx, "Scope1").Times(1).Return(nil).Times(1)
// 			},
// 			ctx:     ctx,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setup()
// 			s := NewMetricServiceServer(rep, nil)
// 			_, err := s.DropMetricData(tt.ctx, tt.input)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("metricServiceServer.DropMetricData() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 		})
// 	}
// }
