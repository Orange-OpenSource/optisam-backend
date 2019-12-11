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
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_CreateMetricOracleProcessorStandard(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.CreateMetricOPS
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.CreateMetricOPS
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

				mockRepo.EXPECT().CreateMetricOPS(ctx, &repo.MetricOPS{
					Name:                  "OPS",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
				}, []string{"A", "B"}).Times(1).Return(&repo.MetricOPS{
					ID:                    "m1",
					Name:                  "OPS",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
				}, nil)
			},
			want: &v1.CreateMetricOPS{
				ID:                    "m1",
				Name:                  "OPS",
				NumCoreAttrId:         "a1",
				NumCPUAttrId:          "a2",
				CoreFactorAttrId:      "a3",
				StartEqTypeId:         "e1",
				AggerateLevelEqTypeId: "e3",
				BaseEqTypeId:          "e2",
				EndEqTypeId:           "e4",
			},
		},
		{name: "SUCCESS - only one level present",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e2",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e2",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

				mockRepo.EXPECT().CreateMetricOPS(ctx, &repo.MetricOPS{
					Name:                  "OPS",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e2",
					AggerateLevelEqTypeID: "e2",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e2",
				}, []string{"A", "B"}).Times(1).Return(&repo.MetricOPS{
					ID:                    "m1",
					Name:                  "OPS",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e2",
					AggerateLevelEqTypeID: "e2",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e2",
				}, nil)
			},
			want: &v1.CreateMetricOPS{
				ID:                    "m1",
				Name:                  "OPS",
				NumCoreAttrId:         "a1",
				NumCPUAttrId:          "a2",
				CoreFactorAttrId:      "a3",
				StartEqTypeId:         "e2",
				AggerateLevelEqTypeId: "e2",
				BaseEqTypeId:          "e2",
				EndEqTypeId:           "e2",
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - starttype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

			},
			wantErr: true,
		},
		{name: "FAILURE - basetype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - aggtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:             "OPS",
					NumCoreAttrId:    "server.cores.number",
					NumCPUAttrId:     "server.processors.number",
					CoreFactorAttrId: "server.corefactor",
					StartEqTypeId:    "e1",
					BaseEqTypeId:     "e2",
					EndEqTypeId:      "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - endtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					BaseEqTypeId:          "e2",
					AggerateLevelEqTypeId: "e3",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - metric name already exists - case insensitive",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "ops",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},

			wantErr: true,
		},
		{name: "FAILURE - start type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e11",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - base type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e22",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - agg type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e33",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e44",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - parent hierachy not found",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e33",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of start level",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e33",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
					&repo.EquipmentType{
						ID: "e33",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - agg level is not ancestor of base level",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e3",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of agg level",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e5",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID:       "e4",
						ParentID: "e5",
					},
					&repo.EquipmentType{
						ID: "e5",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - empty attribut",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - attr1 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a11",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr1 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeString,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a22",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeString,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a33",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 data type is not int/float",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeString,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot create metric",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricOPS{
					Name:                  "OPS",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "ONS",
					},
					&repo.Metric{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "e1",
						ParentID: "e2",
					},
					&repo.EquipmentType{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							&repo.Attribute{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					&repo.EquipmentType{
						ID:       "e3",
						ParentID: "e4",
					},
					&repo.EquipmentType{
						ID: "e4",
					},
				}, nil)

				mockRepo.EXPECT().CreateMetricOPS(ctx, &repo.MetricOPS{
					Name:                  "OPS",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
				}, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.CreateMetricOracleProcessorStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.CreateMetricOracleProcessorStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricdata(t, "CreateMetricOPS", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func compareMetricdata(t *testing.T, name string, exp *v1.CreateMetricOPS, act *v1.CreateMetricOPS) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metric is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.NumCoreAttrId, act.NumCoreAttrId, "%s.numofcores attribute are not same", name)
	assert.Equalf(t, exp.NumCPUAttrId, act.NumCPUAttrId, "%s.numofprocessors attribute are not same", name)
	assert.Equalf(t, exp.CoreFactorAttrId, act.CoreFactorAttrId, "%s.Corefactor attribute are not same", name)
	assert.Equalf(t, exp.StartEqTypeId, act.StartEqTypeId, "%s.StartEqtype id are not same", name)
	assert.Equalf(t, exp.BaseEqTypeId, act.BaseEqTypeId, "%s.BaseEqtype id are not same", name)
	assert.Equalf(t, exp.AggerateLevelEqTypeId, act.AggerateLevelEqTypeId, "%s.AggLevelType id are not same", name)
	assert.Equalf(t, exp.EndEqTypeId, act.EndEqTypeId, "%s.EndLevelType id are not same", name)

}
