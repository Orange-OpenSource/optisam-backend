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

func Test_metricServiceServer_CreateMetricAttrCounterStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	eqTypes := []*repo.EquipmentType{
		&repo.EquipmentType{
			ID:       "e2",
			Type:     "eqType2",
			ParentID: "e3",
			Attributes: []*repo.Attribute{
				&repo.Attribute{
					Name:         "a1",
					Type:         repo.DataTypeInt,
					IsSearchable: true,
				},
				&repo.Attribute{
					Name:         "a2",
					Type:         repo.DataTypeFloat,
					IsSearchable: true,
				},
				&repo.Attribute{
					Name:         "a3",
					Type:         repo.DataTypeString,
					IsSearchable: true,
				},
			},
		},
	}

	type args struct {
		ctx context.Context
		req *v1.CreateMetricACS
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.CreateMetricACS
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().CreateMetricACS(ctx, &repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
				}, eqTypes[0].Attributes[0], "Scope1").Return(&repo.MetricACS{
					ID:            "Met_ACS1ID",
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
				}, nil).Times(1)
			},
			want: &v1.CreateMetricACS{
				ID:            "Met_ACS1ID",
				Name:          "Met_ACS1",
				EqType:        "eqType2",
				AttributeName: "a1",
				Value:         "2",
			},
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
						Name: "Met_ACS1",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType1",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - attribute name is empty",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - attribute doesn't exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a4",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - invalid value type - type should be int",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2.5",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - invalid value type - type should be float",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a2",
					Value:         "abc",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot create metric acs",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().CreateMetricACS(ctx, &repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
				}, eqTypes[0].Attributes[0], "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep)
			got, err := s.CreateMetricAttrCounterStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetricAttrCounterStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.CreateMetricAttrCounterStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}
