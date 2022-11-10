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

func Test_metricServiceServer_CreateMetricEquipAttrStand(t *testing.T) {
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
		req *v1.MetricEquipAtt
	}
	tests := []struct {
		name    string
		s       *metricServiceServer
		args    args
		setup   func()
		want    *v1.MetricEquipAtt
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         5,
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
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().CreateMetricEquipAttrStandard(ctx, &repo.MetricEquipAttrStand{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         5,
				}, eqTypes[0].Attributes[0], "Scope1").Return(&repo.MetricEquipAttrStand{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         5,
				}, nil).Times(1)
			},
			want: &v1.MetricEquipAtt{
				Name:          "Met_EquipAttr1",
				EqType:        "eqType2",
				AttributeName: "a1",
				Environment:   "env1",
				Value:         5,
			},
		},
		{name: "FAILURE - CreateMetricEquipAttrStand - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricEquipAttrStand - scope validation error",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
					Scopes:        []string{"Scope5"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricEquipAttrStand - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - metric name already exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
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
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "Met_EquipAttr1",
						Type: repo.MetricAttrSumStandard,
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateMetricEquipAttrStand - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - cannot find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType1",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - attribute name is empty",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - attribute doesn't exists",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a4",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - string type attribute is not allowed",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a3",
					Environment:   "env1",
					Value:         2,
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
		{name: "FAILURE - CreateMetricEquipAttrStand - cannot create metric attr sum",
			args: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
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
				mockRepo.EXPECT().CreateMetricEquipAttrStandard(ctx, &repo.MetricEquipAttrStand{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
				}, eqTypes[0].Attributes[0], "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.CreateMetricEquipAttrStandard(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetricEquipAttrStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.CreateMetricEquipAttrStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricServiceServer_UpdateMetricEquipAttr(t *testing.T) {
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
		req *v1.MetricEquipAtt
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
				req: &v1.MetricEquipAtt{
					Name:          "Met_EA1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EA1", "Scope1").Return(&repo.MetricEquipAttrStand{
					ID:            "123",
					Name:          "Met_EA1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         3,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricEquipAttr(ctx, &repo.MetricEquipAttrStand{
					Name:          "Met_EA1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env1",
					Value:         2,
				}, "Scope1").Return(nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateMetricEquipAttr - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricEquipAtt{
					Name:          "Met_EA1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - scope validation error",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope5"},
				},
			},
			setup: func() {},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - cannot fetch metrics",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - metric name already exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, repo.ErrNoData).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - cannot fetch equipment types",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - cannot find equipment type",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType1",
					AttributeName: "a1",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - attribute name is empty",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - attribute doesn't exists",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a4",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - string type attribute is not allowed",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a3",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(nil, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
			},
			output: &v1.UpdateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateMetricEquipAttr - cannot update metric attr sum",
			input: args{
				ctx: ctx,
				req: &v1.MetricEquipAtt{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a2",
					Environment:   "env",
					Value:         2,
					Scopes:        []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, "Met_EquipAttr1", "Scope1").Return(&repo.MetricEquipAttrStand{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a2",
					Environment:   "env",
					Value:         1,
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().UpdateMetricEquipAttr(ctx, &repo.MetricEquipAttrStand{
					Name:          "Met_EquipAttr1",
					EqType:        "eqType2",
					AttributeName: "a2",
					Environment:   "env",
					Value:         2,
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
			got, err := s.UpdateMetricEquipAttrStandard(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.UpdateMetricEquipAttrStandard() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.UpdateMetricEquipAttrStandard() got = %v, want %v", got, tt.output)
			}
		})
	}
}
