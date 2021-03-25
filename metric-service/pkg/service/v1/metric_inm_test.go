// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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
		req *v1.CreateINM
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.CreateINM
		outErr bool
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.CreateINM{
					Name:        "Met_INM1",
					Coefficient: 1.5,
					Scopes:      []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					&repo.MetricInfo{
						Name: "ONS",
					},
					&repo.MetricInfo{
						Name: "WS",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1.5,
				}, "Scope1").Return(&repo.MetricINM{
					ID:          "Met_INM1ID",
					Name:        "Met_INM1",
					Coefficient: 1.5,
				}, nil).Times(1)
			},
			output: &v1.CreateINM{
				ID:          "Met_INM1ID",
				Name:        "Met_INM1",
				Coefficient: 1.5,
			},
		},
		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.CreateINM{
					Name:        "Met_INM1",
					Coefficient: 1.6,
					Scopes:      []string{"Scope1"},
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.CreateINM{
					Name:        "Met_INM1",
					Coefficient: 1.6,
					Scopes:      []string{"Scope1"},
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
				req: &v1.CreateINM{
					Name:        "Met_INM1",
					Coefficient: 1.6,
					Scopes:      []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					&repo.MetricInfo{
						Name: "ONS",
					},
					&repo.MetricInfo{
						Name: "Met_INM1",
					},
				}, nil).Times(1)
			},
			outErr: true,
		},

		{name: "FAILURE - CreateMetricInstanceNumberStandard - cannot create metric acs",
			input: args{
				ctx: ctx,
				req: &v1.CreateINM{
					Name:        "Met_INM1",
					Coefficient: 1.5,
					Scopes:      []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					&repo.MetricInfo{
						Name: "ONS",
					},
					&repo.MetricInfo{
						Name: "WS",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, &repo.MetricINM{
					Name:        "Met_INM1",
					Coefficient: 1.5,
				}, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			outErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep)
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
