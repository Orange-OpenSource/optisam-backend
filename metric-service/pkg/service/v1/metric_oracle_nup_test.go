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
	"github.com/stretchr/testify/assert"
)

func Test_metricServiceServer_CreateMetricOracleNUPStandard(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.MetricNUP
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.MetricNUP
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Times(1).Return(&repo.MetricOPSConfig{
					Name:                "OPS",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
				}, nil)

				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
					NumberOfUsers:         2,
					Transform:             true,
					TransformMetricName:   "OPS",
				}, "Scope1").Times(1).Return(&repo.MetricNUPOracle{
					ID:                    "m1",
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
					NumberOfUsers:         2,
					Transform:             true,
					TransformMetricName:   "OPS",
				}, nil)
			},
			want: &v1.MetricNUP{
				ID:                    "m1",
				Name:                  "NUP",
				NumCoreAttrId:         "a1",
				NumCPUAttrId:          "a2",
				CoreFactorAttrId:      "a3",
				StartEqTypeId:         "e1",
				AggerateLevelEqTypeId: "e3",
				BaseEqTypeId:          "e2",
				EndEqTypeId:           "e4",
				NumberOfUsers:         2,
				Transform:             true,
				TransformMetricName:   "OPS",
			},
		},
		{name: "SUCCESS - only one level present",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e2",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e2",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Times(1).Return(&repo.MetricOPSConfig{
					Name:                "OPS",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
				}, nil)
				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e2",
					AggerateLevelEqTypeID: "e2",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e2",
					NumberOfUsers:         2,
					Transform:             true,
					TransformMetricName:   "OPS",
				}, "Scope1").Times(1).Return(&repo.MetricNUPOracle{
					ID:                    "m1",
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e2",
					AggerateLevelEqTypeID: "e2",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e2",
					NumberOfUsers:         2,
					Transform:             true,
					TransformMetricName:   "OPS",
				}, nil)
			},
			want: &v1.MetricNUP{
				ID:                    "m1",
				Name:                  "NUP",
				NumCoreAttrId:         "a1",
				NumCPUAttrId:          "a2",
				CoreFactorAttrId:      "a3",
				StartEqTypeId:         "e2",
				AggerateLevelEqTypeId: "e2",
				BaseEqTypeId:          "e2",
				EndEqTypeId:           "e2",
				NumberOfUsers:         2,
				Transform:             true,
				TransformMetricName:   "OPS",
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - starttype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

			},
			wantErr: true,
		},
		{name: "FAILURE - basetype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - aggtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:             "NUP",
					NumCoreAttrId:    "server.cores.number",
					NumCPUAttrId:     "server.processors.number",
					CoreFactorAttrId: "server.corefactor",
					StartEqTypeId:    "e1",
					BaseEqTypeId:     "e2",
					EndEqTypeId:      "e4",
					NumberOfUsers:    2,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - endtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					BaseEqTypeId:          "e2",
					AggerateLevelEqTypeId: "e3",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "NUP",
					},
					{
						Name: "WS",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - metric name already exists - case insensitive",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "nup",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "NUP",
					},
					{
						Name: "WS",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(nil, errors.New("test error"))

			},

			wantErr: true,
		},
		{name: "FAILURE - start type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e11",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - base type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e22",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - agg type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e33",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e44",
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - parent hierachy not found",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e33",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of start level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e33",
					},
					{
						ID: "e4",
					},
					{
						ID: "e33",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - agg level is not ancestor of base level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e3",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of agg level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e5",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID:       "e4",
						ParentID: "e5",
					},
					{
						ID: "e5",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - empty attribute NumCoreAttrId",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},

			wantErr: true,
		},
		{name: "FAILURE - empty attribute NumCPUAttrId",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - empty attribute CoreFactorAttrId",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr1 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a11",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr1 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeString,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a22",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeString,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a33",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 data type is not int/float",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeString,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - transform metric name cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - transform metric name not exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, repo.ErrNoData).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - transform metric name cannot be empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).AnyTimes()

			},
			wantErr: true,
		},
		{name: "FAILURE - transform metric name should be empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             false,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).AnyTimes()

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot create metric",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)

				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
					NumberOfUsers:         2,
				}, "Scope1").Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricOracleNUPStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetricOracleNUPStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricOracleNUPdata(t, "CreateMetricOracleNUP", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricNUP(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.MetricNUP
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.UpdateMetricResponse
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e3",
					AggerateLevelEqType: "e1",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Times(1).Return(&repo.MetricOPSConfig{
					Name:                "OPS",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
				}, nil)
				mockRepo.EXPECT().UpdateMetricNUP(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
					NumberOfUsers:         2,
					Transform:             true,
					TransformMetricName:   "OPS",
				}, "Scope1").Return(nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "SUCCESS - only one level present",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e2",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e2",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e2",
					BaseEqType:          "e2",
					EndEqType:           "e2",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricNUP(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e2",
					AggerateLevelEqTypeID: "e2",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e2",
					NumberOfUsers:         2,
				}, "Scope1").Return(nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - Default Value True, Metric created by import can't be updated error",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Default:               true,
				},
			},
			setup: func() {},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - starttype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).AnyTimes()

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).AnyTimes()
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - basetype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - aggtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:             "NUP",
					NumCoreAttrId:    "server.cores.number",
					NumCPUAttrId:     "server.processors.number",
					CoreFactorAttrId: "server.corefactor",
					StartEqTypeId:    "e1",
					BaseEqTypeId:     "e2",
					EndEqTypeId:      "e4",
					NumberOfUsers:    2,
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - endtype id is not given",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					BaseEqTypeId:          "e2",
					AggerateLevelEqTypeId: "e3",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(nil, errors.New("test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - metric name does not exist",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},

			wantErr: true,
		},
		{name: "FAILURE - start type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e11",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e9",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - base type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e22",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},

			wantErr: true,
		},
		{name: "FAILURE - agg type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e33",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - end type eq doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e44",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - parent hierachy not found",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e33",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of start level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e33",
					},
					{
						ID: "e4",
					},
					{
						ID: "e33",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - agg level is not ancestor of base level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e2",
					BaseEqTypeId:          "e3",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - end level is not ancestor of agg level",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "server.cores.number",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e5",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID:       "e4",
						ParentID: "e5",
					},
					{
						ID: "e5",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - empty attribute",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCPUAttrId:          "server.processors.number",
					CoreFactorAttrId:      "server.corefactor",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "server.cores.number",
					NumCPUAttr:          "server.processors.number",
					CoreFactorAttr:      "server.corefactor",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr1 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a11",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e2",
						ParentID: "e3",
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr1 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeString,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a22",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr2 data type is not int",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeString,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a33",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - attr3 data type is not int/float",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e4",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeString,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot update metric",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e1",
					AggerateLevelEqType: "e3",
					BaseEqType:          "e2",
					EndEqType:           "e2",
					NumberOfUsers:       2,
				}, nil).Times(1)

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)

				mockRepo.EXPECT().UpdateMetricNUP(ctx, &repo.MetricNUPOracle{
					Name:                  "NUP",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					AggerateLevelEqTypeID: "e3",
					BaseEqTypeID:          "e2",
					EndEqTypeID:           "e4",
					NumberOfUsers:         2,
				}, "Scope1").Return(errors.New("test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - transform metric name cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).Times(1)

			},
			want:    nil,
			wantErr: true,
		},
		{name: "FAILURE - transform metric name not exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, repo.ErrNoData).Times(1)

			},
			want:    nil,
			wantErr: true,
		},
		{name: "FAILURE - transform metric name cannot be empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             true,
					TransformMetricName:   "",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).AnyTimes()

			},
			want:    nil,
			wantErr: true,
		},
		{name: "FAILURE - transform metric name should be empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricNUP{
					Name:                  "NUP",
					NumCoreAttrId:         "a1",
					NumCPUAttrId:          "a2",
					CoreFactorAttrId:      "a3",
					StartEqTypeId:         "e1",
					AggerateLevelEqTypeId: "e3",
					BaseEqTypeId:          "e2",
					EndEqTypeId:           "e4",
					NumberOfUsers:         2,
					Scopes:                []string{"Scope1"},
					Transform:             false,
					TransformMetricName:   "OPS",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Return(&repo.MetricNUPConfig{
					Name:                "NUP",
					NumCoreAttr:         "a1",
					NumCPUAttr:          "a2",
					CoreFactorAttr:      "a3",
					StartEqType:         "e9",
					AggerateLevelEqType: "e8",
					BaseEqType:          "e7",
					EndEqType:           "e6",
					NumberOfUsers:       2,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS", "Scope1").Return(nil, errors.New("test error")).AnyTimes()

			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.UpdateMetricOracleNUPStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.UpdateMetricOracleNUPStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.UpdateMetricOracleNUPStandard() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareMetricOracleNUPdata(t *testing.T, name string, exp *v1.MetricNUP, act *v1.MetricNUP) {
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
	assert.Equalf(t, exp.NumberOfUsers, act.NumberOfUsers, "%s.NumberOfUsers id are not same", name)
}
