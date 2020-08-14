// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"
	"optisam-backend/metric-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_metricServiceServer_ListMetricType(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	type args struct {
		ctx context.Context
		req *v1.ListMetricTypeRequest
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
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
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
			},
			want: &v1.ListMetricTypeResponse{
				Types: []*v1.MetricType{
					&v1.MetricType{
						Name:        string(repo.MetricOPSOracleProcessorStandard),
						Description: "Number of processor licenses required = CPU nb x Core(per CPU) nb x CoreFactor",
						Href:        "/api/v1/metric/ops",
						TypeId:      v1.MetricType_Oracle_Processor,
					},
					&v1.MetricType{
						Name:        string(repo.MetricSPSSagProcessorStandard),
						Description: "Number of processor licenses required = MAX(Prod_licenses, NonProd_licenses) : licenses = CPU nb x Core(per CPU) nb x CoreFactor",
						Href:        "/api/v1/metric/sps",
						TypeId:      v1.MetricType_SAG_Processor,
					},
					&v1.MetricType{
						Name:        string(repo.MetricIPSIbmPvuStandard),
						Description: "Number of licenses required = CPU nb x Core(per CPU) nb x CoreFactor",
						Href:        "/api/v1/metric/ips",
						TypeId:      v1.MetricType_IBM_PVU,
					},
					&v1.MetricType{
						Name:        string(repo.MetricOracleNUPStandard),
						Description: "Named User Plus licenses required = MAX(A,B) : A = CPU nb x Core(per CPU) nb x CoreFactor x minimum number of NUP per processor, B = total number of current users with access to the Oracle product",
						Href:        "/api/v1/metric/oracle_nup",
						TypeId:      v1.MetricType_Oracle_NUP,
					},
					&v1.MetricType{
						Name:        string(repo.MetricAttrCounterStandard),
						Description: repo.MetricDescriptionAttrCounterStandard.String(),
						Href:        "/api/v1/metric/acs",
						TypeId:      v1.MetricType_Attr_Counter,
					},
					&v1.MetricType{
						Name:        string(repo.MetricInstanceNumberStandard),
						Description: repo.MetricDescriptionInstanceNumberStandard.String(),
						Href:        "/api/v1/metric/inm",
						TypeId:      v1.MetricType_Instance_Number,
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
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("FAILURE - cannot fetch metric types info"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep)
			got, err := s.ListMetricType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.ListMetricType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricTypesResponse(t, "ListMetricTypeResponse", got, tt.want)
			}
		})
	}

}
func Test_metricServiceServer_ListMetrices(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.ListMetricRequest
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
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
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricInfo{
					&repo.MetricInfo{
						Name: "Oracle Type1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.MetricInfo{
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
						Description: "Number of processor licenses required = CPU nb x Core(per CPU) nb x CoreFactor",
					},
					&v1.Metric{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "Oracle Type2",
						Description: "Number of processor licenses required = CPU nb x Core(per CPU) nb x CoreFactor",
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
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, []string{"A", "B"}).Times(1).Return(repo.MetricTypes, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricInfo{
					&repo.MetricInfo{
						Name: "Oracle Type1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.MetricInfo{
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
						Description: "Number of processor licenses required = CPU nb x Core(per CPU) nb x CoreFactor",
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
				mockRepo := mock.NewMockMetric(mockCtrl)
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
				mockRepo := mock.NewMockMetric(mockCtrl)
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
			s := NewMetricServiceServer(rep)
			got, err := s.ListMetrices(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.ListMetrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListMetricResponse(t, "ListMetricResponse", got, tt.want)
			}
		})
	}
}

func Test_metricServiceServer_GetMetricConfiguration(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.GetMetricConfigurationRequest
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.GetMetricConfigurationResponse
		wantErr bool
	}{
		{name: "SUCCESS - metric OPS",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS1", []string{"Scope1", "Scope2"}).Return(&repo.MetricOPSConfig{
					ID:                  "OPS1Id",
					Name:                "OPS1",
					NumCPUAttr:          "cpuattr",
					NumCoreAttr:         "coreattr",
					CoreFactorAttr:      "corefactorattr",
					BaseEqType:          "s1",
					StartEqType:         "p1",
					EndEqType:           "d1",
					AggerateLevelEqType: "a1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "OPS1Id",
					"Name":                "OPS1",
					"NumCPUAttr":          "cpuattr",
					"NumCoreAttr":         "coreattr",
					"CoreFactorAttr":      "corefactorattr",
					"BaseEqType":          "s1",
					"StartEqType":         "p1",
					"EndEqType":           "d1",
					"AggerateLevelEqType": "a1"
				}`,
			},
		},
		{name: "SUCCESS - metric NUP",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.nup.standard",
						Name: "NUP1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP1", []string{"Scope1", "Scope2"}).Return(&repo.MetricNUPConfig{
					ID:                  "NUP1ID",
					Name:                "NUP1",
					NumCPUAttr:          "cpuattr",
					NumCoreAttr:         "coreattr",
					CoreFactorAttr:      "corefactorattr",
					BaseEqType:          "s1",
					StartEqType:         "p1",
					EndEqType:           "d1",
					AggerateLevelEqType: "a1",
					NumberOfUsers:       10,
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "NUP1ID",
					"Name":                "NUP1",
					"NumCPUAttr":          "cpuattr",
					"NumCoreAttr":         "coreattr",
					"CoreFactorAttr":      "corefactorattr",
					"BaseEqType":          "s1",
					"StartEqType":         "p1",
					"EndEqType":           "d1",
					"AggerateLevelEqType": "a1",
					"NumberOfUsers":       10
				}`,
			},
		},
		{name: "SUCCESS - metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "sag.processor.standard",
						Name: "SPS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
					{
						ID:   "SPSID",
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigSPS(ctx, "SPS", []string{"Scope1", "Scope2"}).Return(&repo.MetricSPSConfig{
					ID:             "SPSID",
					Name:           "SPS",
					NumCoreAttr:    "coreattr",
					CoreFactorAttr: "corefactorattr",
					BaseEqType:     "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "SPSID",
					"Name":                "SPS",
					"NumCoreAttr":         "coreattr",
					"CoreFactorAttr":      "corefactorattr",
					"BaseEqType":          "s1"
				}`,
			},
		},
		{name: "SUCCESS - metric IPS",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "ibm.pvu.standard",
						Name: "IPS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
					{
						ID:   "IPSID",
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", []string{"Scope1", "Scope2"}).Return(&repo.MetricIPSConfig{
					ID:             "IPSID",
					Name:           "IPS",
					NumCoreAttr:    "coreattr",
					CoreFactorAttr: "corefactorattr",
					BaseEqType:     "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "IPSID",
					"Name":                "IPS",
					"NumCoreAttr":         "coreattr",
					"CoreFactorAttr":      "corefactorattr",
					"BaseEqType":          "s1"
				}`,
			},
		},
		{name: "SUCCESS - metric ACS",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "attribute.counter.standard",
						Name: "ACS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
					{
						ID:   "ACSID",
						Name: "ACS",
						Type: repo.MetricAttrCounterStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "ACS", []string{"Scope1", "Scope2"}).Return(&repo.MetricACS{
					ID:            "ACSID",
					Name:          "ACS",
					AttributeName: "corefactor",
					Value:         "4",
					EqType:        "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":               "ACSID",
					"Name":             "ACS",
					"AttributeName": 	"corefactor",
					"Value":         	"4",
					"EqType":       	"s1"
				}`,
			},
		},
		{name: "SUCCESS - metric INM",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "instance.number.standard",
						Name: "INM",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
					{
						ID:   "INMID",
						Name: "INM",
						Type: repo.MetricInstanceNumberStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "INM", []string{"Scope1", "Scope2"}).Return(&repo.MetricINMConfig{
					ID:          "INMID",
					Name:        "INM",
					Coefficient: 10,
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":          "INMID",
					"Name":        "INM",
					"Coefficient": 10
				}`,
			},
		},
		{name: "FAILURE - GetMetricConfiguration - can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS1",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - metric name and type can not be empty",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "",
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - ListMetrices - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - ListMetrices - metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS5",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - ListMetrices - invalid metric type",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "attribute.counter.standard",
						Name: "OPS1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigOPS - cannot fetch config metric ops",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS1", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigNUP - cannot fetch config metric nup",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.nup.standard",
						Name: "NUP1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "NUP1ID",
						Name: "NUP1",
						Type: repo.MetricOracleNUPStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP1", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigSPS - cannot fetch config metric sps",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "sag.processor.standard",
						Name: "SPS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "SPSID",
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigSPS(ctx, "SPS", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigIPS - cannot fetch config metric ips",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "ibm.pvu.standard",
						Name: "IPS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "IPSID",
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigACS - cannot fetch config metric acs",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "attribute.counter.standard",
						Name: "ACS",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "ACID",
						Name: "ACS",
						Type: repo.MetricAttrCounterStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "ACS", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigINM - cannot fetch config metric inm",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "instance.number.standard",
						Name: "INM",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2"}).Return([]*repo.MetricInfo{
					{
						ID:   "OPS1Id",
						Name: "OPS1",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "OPS2Id",
						Name: "OPS2",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						ID:   "INMID",
						Name: "INM",
						Type: repo.MetricInstanceNumberStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "INM", []string{"Scope1", "Scope2"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep)
			got, err := s.GetMetricConfiguration(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.GetMetricConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.JSONEqf(t, tt.want.MetricConfig, got.MetricConfig, "metricServiceServer.GetMetricConfiguration() metric config is not same")
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
