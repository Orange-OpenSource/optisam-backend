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

func Test_metricServiceServer_CreateMetricAttrSumStandard(t *testing.T) {
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
					IntVal:       10,
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
		req *v1.MetricAttrSum
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.MetricAttrSum
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 5,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().CreateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 5,
				}, eqTypes[0].Attributes[0], "Scope1").Return(&repo.MetricAttrSumStand{
					ID:             "Met_AttrSum1ID",
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 5,
				}, nil).Times(1)
			},
			want: &v1.MetricAttrSum{
				ID:             "Met_AttrSum1ID",
				Name:           "Met_AttrSum1",
				EqType:         "eqType2",
				AttributeName:  "a1",
				ReferenceValue: 5,
			},
		},
		{name: "FAILURE - CreateMetricAttrSumStandard - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrSumStandard - scope validation error",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope5"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrSumStandard - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "Met_AttrSum1",
						Type: repo.MetricAttrSumStandard,
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricAttrSumStandard - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - cannot find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType1",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - attribute name is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - attribute doesn't exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a4",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - string type attribute is not allowed",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a3",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
		{name: "FAILURE - CreateMetricAttrSumStandard - cannot create metric attr sum",
			args: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
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
				mockRepo.EXPECT().CreateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
				}, eqTypes[0].Attributes[0], "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricAttrSumStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetricAttrSumStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.CreateMetricAttrSumStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricAttrSum(t *testing.T) {
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
					IntVal:       10,
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
		req *v1.MetricAttrSum
	}
	tests := []struct {
		name    string
		serObj  *metricServiceServer
		input   args
		setup   func()
		wantErr bool
		output  *v1.UpdateMetricResponse
	}{
		{
			name: "SUCCESS",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_ATT1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 0.24,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_ATT1", "Scope1").Return(&repo.MetricAttrSumStand{
					ID:             "123",
					Name:           "Met_ATT1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 2,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
					Name:           "Met_ATT1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 0.24,
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricAttrSum - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope5"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - Default Value True, Metric created by import can't be updated",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
					Default:        true,
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - cannot fetch equipment types",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - cannot find equipment type",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType1",
					AttributeName:  "a1",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - attribute name is empty",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - attribute doesn't exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a4",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - string type attribute is not allowed",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a3",
					ReferenceValue: 2,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricAttrSum - cannot update metric attr sum",
			input: args{
				ctx: ctx,
				req: &v1.MetricAttrSum{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 0.24,
					Scopes:         []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "Met_AttrSum1", "Scope1").Return(&repo.MetricAttrSumStand{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 1,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
					Name:           "Met_AttrSum1",
					EqType:         "eqType2",
					AttributeName:  "a2",
					ReferenceValue: 0.24,
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
			got, err := s.UpdateMetricAttrSumStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.UpdateMetricAttributeStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricAttributeStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
