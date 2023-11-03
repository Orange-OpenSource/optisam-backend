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

func Test_metricServiceServer_CreateMetricWindowServerStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.MetricWSS
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricWSS
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope1"},
					Default:    true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, &repo.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				}).Return(&repo.MetricWSS{
					ID:         "Met_ScopeSQL1ID",
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				}, nil).Times(1)
			},
			output: &v1.MetricWSS{
				ID:         "Met_ScopeSQL1ID",
				MetricType: "window.server.standard",
				MetricName: "window.server.standard.2019",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Scopes:     []string{"scope1"},
				Default:    true,
			},
		},
		{name: "FAILURE - CreateMetricWindowServerStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope1"},
					Default:    true,
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricWindowServerStandard - do not have access to the scope",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope3"},
					Default:    true,
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricWindowServerStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope1"},
					Default:    true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricWindowServerStandard - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope1"},
					Default:    true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "window.server.standard.2019",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricWindowServerStandard - cannot create metric sql_standard",
			input: args{
				ctx: ctx,
				req: &v1.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scopes:     []string{"scope1"},
					Default:    true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, &repo.MetricWSS{
					MetricType: "window.server.standard",
					MetricName: "window.server.standard.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricWindowServerStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricWindowServerStandard() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricWindowServerStandard() = %v, want %v", got, tt.output)
			}
		})
	}
}
