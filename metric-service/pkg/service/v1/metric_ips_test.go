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

func Test_metricServiceServer_CreateMetricIBMPvuStandard(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	eqTypes := []*repo.EquipmentType{
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
	}

	type args struct {
		ctx context.Context
		req *v1.MetricIPS
	}
	tests := []struct {
		name    string
		args    args
		want    *v1.MetricIPS
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
				mockRepo.EXPECT().CreateMetricIPS(ctx, &repo.MetricIPS{
					Name:             "IPS",
					NumCoreAttrID:    "a1",
					CoreFactorAttrID: "a3",
					BaseEqTypeID:     "e2",
				}, "Scope1").Times(1).Return(&repo.MetricIPS{
					ID:               "IPS",
					Name:             "IPS",
					NumCoreAttrID:    "a1",
					CoreFactorAttrID: "a3",
					BaseEqTypeID:     "e2",
				}, nil)
			},
			want: &v1.MetricIPS{
				ID:               "IPS",
				Name:             "IPS",
				NumCoreAttrId:    "a1",
				CoreFactorAttrId: "a3",
				BaseEqTypeId:     "e2",
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metric types",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return(nil, errors.New("Test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(nil, errors.New("Test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot create metric",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
				mockRepo.EXPECT().CreateMetricIPS(ctx, &repo.MetricIPS{
					Name:             "IPS",
					NumCoreAttrID:    "a1",
					CoreFactorAttrID: "a3",
					BaseEqTypeID:     "e2",
				}, "Scope1").Times(1).Return(nil, errors.New("Test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "IPS",
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
				req: &v1.MetricIPS{
					Name:             "ips",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "IPS",
					},
					{
						Name: "WS",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
		{name: "FAILURE - num of cores attribute is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - numofcores attribute doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a4",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - numofcores attribute doesnt have valid data type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a4",
					BaseEqTypeId:     "e2",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute doesnt have valid data type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
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
				}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			if tt.setup == nil {
				defer mockCtrl.Finish()
			}
			got, err := s.CreateMetricIBMPvuStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetricIBMPvuStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.CreateMetricIBMPvuStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricIPS(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	eqTypes := []*repo.EquipmentType{
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
	}
	type args struct {
		ctx context.Context
		req *v1.MetricIPS
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
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e1",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricIPS(ctx, &repo.MetricIPS{
					Name:             "IPS",
					NumCoreAttrID:    "a1",
					CoreFactorAttrID: "a3",
					BaseEqTypeID:     "e2",
				}, "Scope1").Return(nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot get metric config",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(nil, errors.New("Test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e5",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("Test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot update metric",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e4",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
				mockRepo.EXPECT().UpdateMetricIPS(ctx, &repo.MetricIPS{
					Name:             "IPS",
					NumCoreAttrID:    "a1",
					CoreFactorAttrID: "a3",
					BaseEqTypeID:     "e2",
				}, "Scope1").Return(errors.New("Test error")).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - metric name does not exist",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e4",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e7",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - num of cores attribute is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e2",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e2",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - numofcores attribute doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a6",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a9",
					CoreFactorAttr: "a3",
					BaseEqType:     "e2",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - numofcores attribute doesnt have valid data type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "1",
					CoreFactorAttrId: "a3",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e2",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute doesnt exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "a9",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a8",
					BaseEqType:     "e3",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - core factor attribute doesnt have valid data type",
			args: args{
				ctx: ctx,
				req: &v1.MetricIPS{
					Name:             "IPS",
					NumCoreAttrId:    "a1",
					CoreFactorAttrId: "1",
					BaseEqTypeId:     "e2",
					Scopes:           []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					Name:           "IPS",
					NumCoreAttr:    "a1",
					CoreFactorAttr: "a3",
					BaseEqType:     "e2",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return(eqTypes, nil).Times(1)
			},
			want: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.UpdateMetricIBMPvuStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.UpdateMetricIBMPvuStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.UpdateMetricIBMPvuStandard() got = %v, want %v", got, tt.want)
			}
		})
	}
}
