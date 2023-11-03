package v1

import (
	"context"
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

func Test_metricServiceServer_CreateScopeMetric(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	var acc accv1.AccountServiceClient
	type args struct {
		ctx context.Context
		req *v1.CreateScopeMetricRequest
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.CreateScopeMetricResponse
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.CreateScopeMetricRequest{
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockRepo.EXPECT().CreateMetricSQLForScope(ctx, &repo.ScopeMetric{
					MetricType: "microsoft.sql.enterprise",
					MetricName: "microsoft.sql.enterprise.2019",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				}).Return(&repo.ScopeMetric{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricDataCenterForScope(ctx, &repo.ScopeMetric{
					MetricType: "windows.server.datacenter",
					MetricName: "windows.server.datacenter.2016",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				}).Return(&repo.ScopeMetric{}, nil).AnyTimes()
			},
			output: &v1.CreateScopeMetricResponse{
				Success: true,
			},
			outErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &metricServiceServer{
				metricRepo: rep,
				account:    acc,
			}
			got, err := s.CreateScopeMetric(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricForScope() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricForScope() = %v, want %v", got, tt.output)
			}
		})
	}
}
