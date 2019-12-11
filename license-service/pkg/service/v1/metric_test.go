// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_ListMetricType(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.ListMetricTypeRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListMetricTypeResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
			},
			want: &v1.ListMetricTypeResponse{
				Types: []*v1.MetricType{
					&v1.MetricType{
						Name:        string(repo.MetricOPSOracleProcessorStandard),
						Description: "xyz",
						Href:        "/api/v1/metric/ops",
						TypeId:      v1.MetricType_Oracle_Processor,
					},
					&v1.MetricType{
						Name:        string(repo.MetricSPSSagProcessorStandard),
						Description: "abc",
						Href:        "/api/v1/metric/sps",
						TypeId:      v1.MetricType_SAG_Processor,
					},
					&v1.MetricType{
						Name:        string(repo.MetricIPSIbmPvuStandard),
						Description: "pqr",
						Href:        "/api/v1/metric/ips",
						TypeId:      v1.MetricType_IBM_PVU,
					},
				},
			},
		},
		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metric types info",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListMetricType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListMetricType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricTypesResponse(t, "ListMetricTypeResponse", got, tt.want)
			}
		})
	}

}
func Test_licenseServiceServer_ListMetrices(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.ListMetricRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListMetricResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "Oracle Type1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "Oracle Type2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					&v1.Metric{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "Oracle Type1",
						Description: "xyz",
					},
					&v1.Metric{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "Oracle Type2",
						Description: "xyz",
					},
				},
			},
		},
		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "SUCCESS - description not found",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "Oracle Type1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "Oracle Type2",
						Type: repo.MetricType("Windows.processor"),
					},
				}, nil)

			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					&v1.Metric{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "Oracle Type1",
						Description: "xyz",
					},
					&v1.Metric{
						Type: "Windows.processor",
						Name: "Oracle Type2",
					},
				},
			},
		},
		{name: "FAILURE - cannot fetch metric types info",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrices",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListMetrices(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListMetrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListMetricResponse(t, "ListMetricResponse", got, tt.want)
			}
		})
	}
}

func compareMetricTypesResponse(t *testing.T, name string, exp *v1.ListMetricTypeResponse, act *v1.ListMetricTypeResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	compareMetricTypeAll(t, name+".Types", exp.Types, act.Types)
}

func compareMetricTypeAll(t *testing.T, name string, exp []*v1.MetricType, act []*v1.MetricType) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetricType(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetricType(t *testing.T, name string, exp *v1.MetricType, act *v1.MetricType) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Names are not same", name)
	assert.Equalf(t, exp.Description, act.Description, "%s.Descriptions are not same", name)
	assert.Equalf(t, exp.Href, act.Href, "%s.Href are not same", name)

}

func compareListMetricResponse(t *testing.T, name string, exp *v1.ListMetricResponse, act *v1.ListMetricResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	compareMetricAll(t, name+".Metrices", exp.Metrices, act.Metrices)
}

func compareMetricAll(t *testing.T, name string, exp []*v1.Metric, act []*v1.Metric) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetric(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetric(t *testing.T, name string, exp *v1.Metric, act *v1.Metric) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Names are not same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Types are not same", name)
	assert.Equalf(t, exp.Description, act.Description, "%s.Descriptions are not same", name)

}
