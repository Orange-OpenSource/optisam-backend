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

func Test_metricServiceServer_CreateMetricAttrCounterStandard(t *testing.T) {
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
			Type:     "eqType2",
			ParentID: "e3",
			Attributes: []*repo.Attribute{
				{
					Name:         "a1",
					Type:         repo.DataTypeInt,
					IsSearchable: true,
				},
				{
					Name:         "a2",
					Type:         repo.DataTypeFloat,
					IsSearchable: true,
				},
				{
					Name:         "a3",
					Type:         repo.DataTypeString,
					IsSearchable: true,
				},
			},
		},
	}

	type args struct {
		ctx context.Context
		req *v1.MetricACS
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.MetricACS
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
			want: &v1.MetricACS{
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
				req: &v1.MetricACS{
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
				req: &v1.MetricACS{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
						Name: "Met_ACS1",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrCounterStandard - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
				req: &v1.MetricACS{
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
					{
						Name: "ONS",
					},
					{
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
			s := NewMetricServiceServer(rep, nil)
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

func Test_metricServiceServer_UpdateMetricACS(t *testing.T) {
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
			Type:     "eqType2",
			ParentID: "e3",
			Attributes: []*repo.Attribute{
				{
					Name:         "a1",
					Type:         repo.DataTypeInt,
					IsSearchable: true,
				},
				{
					Name:         "a2",
					Type:         repo.DataTypeFloat,
					IsSearchable: true,
				},
				{
					Name:         "a3",
					Type:         repo.DataTypeString,
					IsSearchable: true,
				},
			},
		},
	}
	type args struct {
		ctx context.Context
		req *v1.MetricACS
	}
	tests := []struct {
		name    string
		serObj  *metricServiceServer
		input   args
		setup   func()
		wantErr bool
		output  *v1.UpdateMetricResponse
	}{
		{name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricACS(ctx, &repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricACS - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope5"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - Default Value True, Metric created by import can't be updated error",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
					Scopes:        []string{"Scope1"},
					Default:       true,
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - cannot fetch metric config",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(nil, errors.New("internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - metric does not exist",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - cannot fetch equipment types",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - cannot find equipment type",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - attribute name is empty",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - attribute doesn't exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - invalid value type - type should be int",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - invalid value type - type should be float",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricACS - cannot update metric acs",
			input: args{
				ctx: ctx,
				req: &v1.MetricACS{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "Met_ACS1", "Scope1").Return(&repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "12",
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricACS(ctx, &repo.MetricACS{
					Name:          "Met_ACS1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Value:         "2",
				}, "Scope1").Return(errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.UpdateMetricAttrCounterStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.UpdateMetricAttrCounterStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricAttrCounterStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
