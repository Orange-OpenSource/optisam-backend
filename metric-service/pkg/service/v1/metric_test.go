package v1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1"
	accmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/account-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1/mock"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_DropMetrics(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric

	tests := []struct {
		name    string
		s       *metricServiceServer
		input   *v1.DropMetricDataRequest
		setup   func()
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "ScopeNotFound",
			input:   &v1.DropMetricDataRequest{Scope: "Scope6"},
			setup:   func() {},
			ctx:     ctx,
			wantErr: true,
		},

		{
			name:    "ClaimsNotFound",
			input:   &v1.DropMetricDataRequest{Scope: "Scope6"},
			setup:   func() {},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:  "DBError",
			input: &v1.DropMetricDataRequest{Scope: "Scope1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropMetrics(ctx, "Scope1").Return(errors.New("DBError")).Times(1)
			},
			ctx:     ctx,
			wantErr: true,
		},
		{
			name:  "SuccessFully metrics Deletion",
			input: &v1.DropMetricDataRequest{Scope: "Scope1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropMetrics(ctx, "Scope1").Times(1).Return(nil).Times(1)
			},
			ctx:     ctx,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			_, err := s.DropMetricData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.DropMetricData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_metricServiceServer_ListMetricType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	var acc accv1.AccountServiceClient
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
		{name: "SUCCESS: generic scope",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "acs",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "ibm",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "sps",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				mockRepo.EXPECT().ListMetricTypeInfo(ctx, repo.GetScopeType(accv1.ScopeType_GENERIC.String()), "Scope1", false).Times(1).Return(repo.MetricTypesGeneric, nil)
			},
			want: &v1.ListMetricTypeResponse{
				// Types: []*v1.MetricType{
				// 	{
				// 		Name:        string(repo.MetricOPSOracleProcessorStandard),
				// 		Description: repo.MetricDescriptionOracleProcessorStandard.String(),
				// 		Href:        "/api/v1/metric/ops",
				// 		TypeId:      v1.MetricType_Oracle_Processor,
				// 	},
				// 	{
				// 		Name:        string(repo.MetricSPSSagProcessorStandard),
				// 		Description: repo.MetricDescriptionSagProcessorStandard.String(),
				// 		Href:        "/api/v1/metric/sps",
				// 		TypeId:      v1.MetricType_SAG_Processor,
				// 	},
				// 	{
				// 		Name:        string(repo.MetricIPSIbmPvuStandard),
				// 		Description: repo.MetricDescriptionIbmPvuStandard.String(),
				// 		Href:        "/api/v1/metric/ips",
				// 		TypeId:      v1.MetricType_IBM_PVU,
				// 	},
				// 	{
				// 		Name:        string(repo.MetricOracleNUPStandard),
				// 		Description: repo.MetricDescriptionOracleNUPStandard.String(),
				// 		Href:        "/api/v1/metric/oracle_nup",
				// 		TypeId:      v1.MetricType_Oracle_NUP,
				// 	},
				// 	{
				// 		Name:        string(repo.MetricAttrCounterStandard),
				// 		Description: repo.MetricDescriptionAttrCounterStandard.String(),
				// 		Href:        "/api/v1/metric/acs",
				// 		TypeId:      v1.MetricType_Attr_Counter,
				// 	},
				// 	{
				// 		Name:        string(repo.MetricInstanceNumberStandard),
				// 		Description: repo.MetricDescriptionInstanceNumberStandard.String(),
				// 		Href:        "/api/v1/metric/inm",
				// 		TypeId:      v1.MetricType_Instance_Number,
				// 	},
				// 	{
				// 		Name:        repo.MetricAttrSumStandard.String(),
				// 		Description: repo.MetricDescriptionAttrSumStandard.String(),
				// 		Href:        "/api/v1/metric/attr_sum",
				// 		TypeId:      v1.MetricType_Attr_Sum,
				// 	},
				// 	{
				// 		Name:        repo.MetricUserSumStandard.String(),
				// 		Description: repo.MetricDescriptionUserSumStandard.String(),
				// 		Href:        "/api/v1/metric/uss",
				// 		TypeId:      v1.MetricType_User_Sum,
				// 	},
				// 	{
				// 		Name:        repo.MetricStaticStandard.String(),
				// 		Description: repo.MetricDescriptionStaticStandard.String(),
				// 		Href:        "/api/v1/metric/ss",
				// 		TypeId:      v1.MetricType_Static_Standard,
				// 	},
				// 	{
				// 		Name:        repo.MetricEquipAttrStandard.String(),
				// 		Description: repo.MetricDescriptionEquipAttrStandard.String(),
				// 		Href:        "/api/v1/metric/equip_attr",
				// 		TypeId:      v1.MetricType_Equip_Attr,
				// 	},
				// 	{
				// 		Name:        repo.MetricUserNomStandard.String(),
				// 		Description: repo.MetricDescriptionUserNomStandard.String(),
				// 		Href:        "/api/v1/metric/uns",
				// 		TypeId:      v1.MetricType_Nominative_User,
				// 	},
				// 	{
				// 		Name:        repo.MetricUserConcurentStandard.String(),
				// 		Description: repo.MetricDescriptionUserConcurentStandard.String(),
				// 		Href:        "/api/v1/metric/user_conc",
				// 		TypeId:      v1.MetricType_User_Concurent,
				// 	},
				// },
			},
		},
		{name: "SUCCESS: specific scope",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_SPECIFIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name:    "acs",
						Type:    repo.MetricAttrCounterStandard,
						Default: false,
					},
					{
						Name:    "ibm",
						Type:    repo.MetricIPSIbmPvuStandard,
						Default: true,
					},
					{
						Name:    "sps",
						Type:    repo.MetricSPSSagProcessorStandard,
						Default: false,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, repo.GetScopeType(accv1.ScopeType_SPECIFIC.String()), "Scope1", false).Times(1).Return(repo.MetricTypesSpecific, nil)
			},
			want: &v1.ListMetricTypeResponse{
				Types: []*v1.MetricType{
					{
						Name:           string(repo.MetricOPSOracleProcessorStandard),
						Description:    repo.MetricDescriptionOracleProcessorStandard.String(),
						Href:           "/api/v1/metric/ops",
						TypeId:         v1.MetricType_Oracle_Processor,
						IsExist:        false,
						DefaultMetrics: []string{"oracle.processor"},
					},
					{
						Name:           string(repo.MetricSPSSagProcessorStandard),
						Description:    repo.MetricDescriptionSagProcessorStandard.String(),
						Href:           "/api/v1/metric/sps",
						TypeId:         v1.MetricType_SAG_Processor,
						IsExist:        false,
						DefaultMetrics: []string{"sag.processor"},
					},
					{
						Name:           string(repo.MetricIPSIbmPvuStandard),
						Description:    repo.MetricDescriptionIbmPvuStandard.String(),
						Href:           "/api/v1/metric/ips",
						TypeId:         v1.MetricType_IBM_PVU,
						IsExist:        true,
						DefaultMetrics: []string{"ibm.pvu"},
					},
					{
						Name:           string(repo.MetricOracleNUPStandard),
						Description:    repo.MetricDescriptionOracleNUPStandard.String(),
						Href:           "/api/v1/metric/oracle_nup",
						TypeId:         v1.MetricType_Oracle_NUP,
						IsExist:        false,
						DefaultMetrics: []string{"oracle.nup"},
					},
					{
						Name:           string(repo.MetricAttrCounterStandard),
						Description:    repo.MetricDescriptionAttrCounterStandard.String(),
						Href:           "/api/v1/metric/acs",
						TypeId:         v1.MetricType_Attr_Counter,
						IsExist:        false,
						DefaultMetrics: []string{"attribute.counter"},
					},
					{
						Name:           string(repo.MetricInstanceNumberStandard),
						Description:    repo.MetricDescriptionInstanceNumberStandard.String(),
						Href:           "/api/v1/metric/inm",
						TypeId:         v1.MetricType_Instance_Number,
						IsExist:        false,
						DefaultMetrics: []string{"one_instance"},
					},
					{
						Name:           repo.MetricAttrSumStandard.String(),
						Description:    repo.MetricDescriptionAttrSumStandard.String(),
						Href:           "/api/v1/metric/attr_sum",
						TypeId:         v1.MetricType_Attr_Sum,
						IsExist:        false,
						DefaultMetrics: []string{"attribute.sum"},
					},
					{
						Name:           repo.MetricStaticStandard.String(),
						Description:    repo.MetricDescriptionStaticStandard.String(),
						Href:           "/api/v1/metric/ss",
						TypeId:         v1.MetricType_Static_Standard,
						IsExist:        false,
						DefaultMetrics: []string{"static.standard"},
					},
					{
						Name:           repo.MetricEquipAttrStandard.String(),
						Description:    repo.MetricDescriptionEquipAttrStandard.String(),
						Href:           "/api/v1/metric/equip_attr",
						TypeId:         v1.MetricType_Equip_Attr,
						IsExist:        false,
						DefaultMetrics: []string{"openshift.premium", "openshift.standard"},
					},
					{
						Name:           repo.MetricUserNomStandard.String(),
						Description:    repo.MetricDescriptionUserNomStandard.String(),
						Href:           "/api/v1/metric/uns",
						TypeId:         v1.MetricType_Nominative_User,
						IsExist:        false,
						DefaultMetrics: []string{"user.nominative.standard"},
					},
					{
						Name:           repo.MetricUserConcurentStandard.String(),
						Description:    repo.MetricDescriptionUserConcurentStandard.String(),
						Href:           "/api/v1/metric/user_conc",
						TypeId:         v1.MetricType_User_Concurent,
						IsExist:        false,
						DefaultMetrics: []string{"user.concurrent.standard"},
					},
				},
			},
		},
		// {name: "SUCCESS: specific scope with NUP metric already present",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListMetricTypeRequest{
		// 			Scopes: []string{"Scope1"},
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockMetric(mockCtrl)
		// 		mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
		// 		acc = mockAcc
		// 		rep = mockRepo
		// 		mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
		// 			ScopeCode:  "Scope1",
		// 			ScopeName:  "Scope 1",
		// 			CreatedBy:  "admin@test.com",
		// 			GroupNames: []string{"ROOT"},
		// 			ScopeType:  accv1.ScopeType_SPECIFIC.String(),
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
		// 			{
		// 				Name: "acs",
		// 				Type: repo.MetricAttrCounterStandard,
		// 			},
		// 			{
		// 				Name: "ibm",
		// 				Type: repo.MetricIPSIbmPvuStandard,
		// 			},
		// 			{
		// 				Name: "sps",
		// 				Type: repo.MetricSPSSagProcessorStandard,
		// 			},
		// 			{
		// 				Name: "nup",
		// 				Type: repo.MetricOracleNUPStandard,
		// 			},
		// 		}, nil)

		// 		mockRepo.EXPECT().ListMetricTypeInfo(ctx, repo.GetScopeType(accv1.ScopeType_SPECIFIC.String()), false, true, "Scope1").Times(1).Return(append(repo.MetricTypesSpecific, repo.MertricTypeOPS), nil)
		// 	},
		// 	want: &v1.ListMetricTypeResponse{
		// 		Types: []*v1.MetricType{
		// 			{
		// 				Name:        string(repo.MetricSPSSagProcessorStandard),
		// 				Description: repo.MetricDescriptionSagProcessorStandard.String(),
		// 				Href:        "/api/v1/metric/sps",
		// 				TypeId:      v1.MetricType_SAG_Processor,
		// 			},
		// 			{
		// 				Name:        string(repo.MetricIPSIbmPvuStandard),
		// 				Description: repo.MetricDescriptionIbmPvuStandard.String(),
		// 				Href:        "/api/v1/metric/ips",
		// 				TypeId:      v1.MetricType_IBM_PVU,
		// 			},
		// 			{
		// 				Name:        string(repo.MetricAttrCounterStandard),
		// 				Description: repo.MetricDescriptionAttrCounterStandard.String(),
		// 				Href:        "/api/v1/metric/acs",
		// 				TypeId:      v1.MetricType_Attr_Counter,
		// 			},
		// 			{
		// 				Name:        string(repo.MetricInstanceNumberStandard),
		// 				Description: repo.MetricDescriptionInstanceNumberStandard.String(),
		// 				Href:        "/api/v1/metric/inm",
		// 				TypeId:      v1.MetricType_Instance_Number,
		// 			},
		// 			{
		// 				Name:        repo.MetricAttrSumStandard.String(),
		// 				Description: repo.MetricDescriptionAttrSumStandard.String(),
		// 				Href:        "/api/v1/metric/attr_sum",
		// 				TypeId:      v1.MetricType_Attr_Sum,
		// 			},
		// 			{
		// 				Name:        string(repo.MetricOPSOracleProcessorStandard),
		// 				Description: string(repo.MetricDescriptionOracleProcessorStandard),
		// 				Href:        "/api/v1/metric/ops",
		// 				TypeId:      v1.MetricType_Oracle_Processor,
		// 			},
		// 		},
		// 	},
		// },
		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - scopevalidation error",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope3"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch scope type",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(nil, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/ListMetrices cannot fetch list metrics",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return(nil, errors.New("FAILURE - cannot fetch list metrics"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metric types info",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricTypeRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "Scope1"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "Scope1",
					ScopeName:  "Scope 1",
					CreatedBy:  "admin@test.com",
					GroupNames: []string{"ROOT"},
					ScopeType:  accv1.ScopeType_GENERIC.String(),
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "acs",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "ibm",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "sps",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricTypeInfo(ctx, repo.GetScopeType(accv1.ScopeType_GENERIC.String()), "Scope1", false).Times(1).Return(nil, errors.New("FAILURE - cannot fetch metric types info"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &metricServiceServer{
				metricRepo: rep,
				account:    acc,
			}
			_, err := s.ListMetricType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.ListMetricType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareMetricTypesResponse(t, "ListMetricTypeResponse", tt.want, got)
			// }
		})
	}

}
func Test_metricServiceServer_ListMetrices(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	var acc accv1.AccountServiceClient
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
				req: &v1.ListMetricRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "INM",
						Type: repo.MetricInstanceNumberStandard,
					},
					{
						Name: "NUP",
						Type: repo.MetricOracleNUPStandard,
					},
					{
						Name: "ACS",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "ATT",
						Type: repo.MetricAttrSumStandard,
					},
				}, nil)
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "INM", "Scope1").Times(1).Return(&repo.MetricINM{
					ID:          "021",
					Name:        "INM",
					Coefficient: 8,
				}, nil)
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP", "Scope1").Times(1).Return(&repo.MetricNUPConfig{
					ID:            "3222",
					Name:          "NUP",
					NumberOfUsers: 5,
				}, nil)
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "ACS", "Scope1").Times(1).Return(&repo.MetricACS{
					ID:            "543",
					Name:          "ACS",
					EqType:        "Equip1",
					AttributeName: "att1",
					Value:         "6",
				}, nil)
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "ATT", "Scope1").Times(1).Return(&repo.MetricAttrSumStand{
					ID:             "521",
					Name:           "ATT",
					EqType:         "Equipment_type",
					AttributeName:  "attribute_value",
					ReferenceValue: 5.00,
				}, nil)
			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "OPS",
						Description: repo.MetricDescriptionOracleProcessorStandard.String(),
					},
					{
						Type:        string(repo.MetricInstanceNumberStandard),
						Name:        "INM",
						Description: "Number of licenses required = Sum of product installations / 8",
					},
					{
						Type:        string(repo.MetricOracleNUPStandard),
						Name:        "NUP",
						Description: "Number Of licenses required = MAX(CPU nb x Core(per CPU) nb x CoreFactor x 5, given number of users)",
					},
					{
						Type:        string(repo.MetricAttrCounterStandard),
						Name:        "ACS",
						Description: "Number of licenses required = Number of equipment of type Equip1 with att1 = 6.",
					},
					{
						Type:        string(repo.MetricAttrSumStandard),
						Name:        "ATT",
						Description: "Number of licenses required = Ceil( Sum( on all equipments of type Equipment_type) of attribute_value)/ 5.00",
					},
				},
			},
		},
		{name: "FAILURE - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListMetricRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - scope validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricRequest{
					Scopes: []string{"Scope5"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "SUCCESS - description not found",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.MetricInfo{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WIN",
						Type: repo.MetricType("Windows.processor"),
					},
				}, nil)

			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					{
						Type:        string(repo.MetricOPSOracleProcessorStandard),
						Name:        "OPS",
						Description: repo.MetricDescriptionOracleProcessorStandard.String(),
					},
					{
						Type: "Windows.processor",
						Name: "WIN",
					},
				},
			},
		},
		{name: "FAILURE - cannot fetch metrices",
			args: args{
				ctx: ctx,
				req: &v1.ListMetricRequest{
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &metricServiceServer{
				metricRepo: rep,
				account:    acc,
			}
			got, err := s.ListMetrices(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.ListMetrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListMetricResponse(t, "ListMetricResponse", tt.want, got)
			}
		})
	}
}

func Test_metricServiceServer_GetMetricConfiguration(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS1", "Scope1").Return(&repo.MetricOPSConfig{
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP1", "Scope1").Return(&repo.MetricNUPConfig{
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
					"NumberOfUsers":       10,
					"Transform": false,
					"TransformMetricName":""
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigSPS(ctx, "SPS", "Scope1").Return(&repo.MetricSPSConfig{
					ID:             "SPSID",
					Name:           "SPS",
					NumCoreAttr:    "coreattr",
					NumCPUAttr:     "cpuattr",
					CoreFactorAttr: "corefactorattr",
					BaseEqType:     "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "SPSID",
					"Name":                "SPS",
					"NumCoreAttr":         "coreattr",
					"NumCPUAttr":          "cpuattr",
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(&repo.MetricIPSConfig{
					ID:             "IPSID",
					Name:           "IPS",
					NumCoreAttr:    "coreattr",
					NumCPUAttr:     "cpuattr",
					CoreFactorAttr: "corefactorattr",
					BaseEqType:     "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":                  "IPSID",
					"Name":                "IPS",
					"NumCoreAttr":         "coreattr",
					"NumCPUAttr":          "cpuattr",
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "ACS", "Scope1").Return(&repo.MetricACS{
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
					"EqType":       	"s1",
					"Default":          false
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "INM", "Scope1").Return(&repo.MetricINM{
					ID:          "INMID",
					Name:        "INM",
					Coefficient: 10,
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":          "INMID",
					"Name":        "INM",
					"Coefficient": 10,
					"Default":false
				}`,
			},
		},
		{name: "SUCCESS - metric attr sum",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "attribute.sum.standard",
						Name: "attrsum",
					},
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
						ID:   "AtteSumID",
						Name: "attrsum",
						Type: repo.MetricAttrSumStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "attrsum", "Scope1").Return(&repo.MetricAttrSumStand{
					ID:             "AtteSumID",
					Name:           "attrsum",
					AttributeName:  "corefactor",
					ReferenceValue: 4,
					EqType:         "s1",
				}, nil).Times(1)
			},
			want: &v1.GetMetricConfigurationResponse{
				MetricConfig: `{
					"ID":               "AtteSumID",
					"Name":             "attrsum",
					"AttributeName": 	"corefactor",
					"ReferenceValue":    4,
					"EqType":       	"s1",
					"Default":          false
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
					Scopes: []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - scope validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS1",
					},
					Scopes: []string{"Scope5"},
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
					Scopes: []string{"Scope1"},
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
					Scopes: []string{"Scope1"},
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
		{name: "FAILURE - GetMetricConfiguration - ListMetrices - metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "oracle.processor.standard",
						Name: "OPS5",
					},
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigOPS(ctx, "OPS1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigNUP(ctx, "NUP1", "Scope1").Return(nil, errors.New("Internal")).Times(1)
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigSPS(ctx, "SPS", "Scope1").Return(nil, errors.New("Internal")).Times(1)
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigIPS(ctx, "IPS", "Scope1").Return(nil, errors.New("Internal")).Times(1)
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigACS(ctx, "ACS", "Scope1").Return(nil, errors.New("Internal")).Times(1)
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
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
				mockRepo.EXPECT().GetMetricConfigINM(ctx, "INM", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMetricConfiguration - GetMetricConfigAttrSum - cannot fetch config metric attr sum standard",
			args: args{
				ctx: ctx,
				req: &v1.GetMetricConfigurationRequest{
					MetricInfo: &v1.Metric{
						Type: "attribute.sum.standard",
						Name: "attrsum",
					},
					Scopes: []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
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
						ID:   "AttrSumID",
						Name: "attrsum",
						Type: repo.MetricAttrSumStandard,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, "attrsum", "Scope1").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
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

func Test_metricServiceServer_DeleteMetric(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.DeleteMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.DeleteMetricResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{
					ID:   "Metric1ID",
					Name: "Metric1",
					Type: repo.MetricOPSOracleProcessorStandard,
				}, nil)
				mockRepo.EXPECT().GetMetricNUPByTransformMetricName(ctx, "Metric1", "Scope1").Times(1).Return(nil, nil)
				mockRepo.EXPECT().DeleteMetric(ctx, "Metric1", "Scope1").Times(1).Return(nil)
			},
			want: &v1.DeleteMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "FAILURE - can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope3",
				},
			},
			setup: func() {},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - MetricInfoWithAcqAndAgg - can not get metric info",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{}, nil)
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - metric is being used by aquired right/aggregation",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{
					ID:                "Metric1ID",
					Name:              "Metric1",
					Type:              repo.MetricOPSOracleProcessorStandard,
					TotalAggregations: 0,
					TotalAcqRights:    2,
				}, nil)
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - Default Value True, Metric created by import can't be updated error",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{
					ID:      "Metric1ID",
					Name:    "Metric1",
					Type:    repo.MetricOPSOracleProcessorStandard,
					Default: true,
				}, nil)
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - metric is being used for transform",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{
					ID:   "Metric1ID",
					Name: "Metric1",
					Type: repo.MetricOPSOracleProcessorStandard,
				}, nil)
				mockRepo.EXPECT().GetMetricNUPByTransformMetricName(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricNUPOracle{
					Name: "Metric1",
				}, nil)
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - unable to delete metric",
			args: args{
				ctx: ctx,
				req: &v1.DeleteMetricRequest{
					MetricName: "Metric1",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetricInfoWithAcqAndAgg(ctx, "Metric1", "Scope1").Times(1).Return(&repo.MetricInfoFull{
					ID:   "Metric1ID",
					Name: "Metric1",
					Type: repo.MetricOPSOracleProcessorStandard,
				}, nil)
				mockRepo.EXPECT().GetMetricNUPByTransformMetricName(ctx, "Metric1", "Scope1").Times(1).Return(nil, nil)
				mockRepo.EXPECT().DeleteMetric(ctx, "Metric1", "Scope1").Times(1).Return(errors.New("Internal"))
			},
			want: &v1.DeleteMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			got, err := s.DeleteMetric(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.DeleteMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricServiceServer.DeleteMetric() = %v, want %v", got, tt.want)
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

func Test_metricServiceServer_CreateMetric(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Metric
	type args struct {
		ctx context.Context
		req *v1.CreateMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.CreateMetricResponse
		wantErr bool
	}{
		{
			name: "FAIL CTX ",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAIL metric name blank ",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "",
					},
				},
			},
			setup: func() {},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS OPS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigOPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricOPS{}, nil)
				mockRepo.EXPECT().CreateMetricOPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricOPS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail OPS -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigOPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricOPS{}, errors.New("err"))
				mockRepo.EXPECT().CreateMetricOPS(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricOPS{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail OPS -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigOPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricOPS{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricOPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricOPS{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS NUP",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOracleNUPStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUPID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricNUPOracle{}, nil)
				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricNUPOracle{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail NUP -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOracleNUPStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUPID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricNUPOracle{}, errors.New("err"))
				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricNUPOracle{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail NUP -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricOracleNUPStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigNUPID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricNUPOracle{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricNUPOracle{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS SPS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricSPSSagProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSPS{}, nil)
				mockRepo.EXPECT().CreateMetricSPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSPS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail sps -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricSPSSagProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSPS{}, errors.New("err"))
				mockRepo.EXPECT().CreateMetricSPS(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricSPS{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail sps -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricSPSSagProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSPS{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricSPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSPS{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS IPS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricIPSIbmPvuStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricIPS{}, nil)
				mockRepo.EXPECT().CreateMetricIPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricIPS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail IPS -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricIPSIbmPvuStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricIPS{}, errors.New("err"))
				mockRepo.EXPECT().CreateMetricIPS(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricIPS{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail IPS -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricIPSIbmPvuStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigIPSID(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricIPS{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricIPS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricIPS{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS ACS",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrCounterStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigACS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricACS{EqType: "string", AttributeName: "string"}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{{Type: "string", Attributes: []*repo.Attribute{{Name: "string"}}}}, nil)
				mockRepo.EXPECT().CreateMetricACS(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricACS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail ACS -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrCounterStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigACS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricACS{}, errors.New("err"))
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).AnyTimes().Return([]*repo.EquipmentType{}, nil)

				mockRepo.EXPECT().CreateMetricACS(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricACS{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail ACS -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrCounterStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigACS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricACS{}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{}, nil)
				mockRepo.EXPECT().CreateMetricACS(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricACS{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS INM",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricInstanceNumberStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricINM{}, nil)
				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricINM{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail INM -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricInstanceNumberStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricINM{}, errors.New("err"))

				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricINM{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail INM -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricInstanceNumberStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigINM(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricINM{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricINM{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS ATTRSUMSTAND",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricAttrSumStand{EqType: "string", AttributeName: "string"}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{{Type: "string", Attributes: []*repo.Attribute{{Name: "string"}}}}, nil)
				mockRepo.EXPECT().CreateMetricAttrSum(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricAttrSumStand{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail ATTRSUMSTAND -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricAttrSumStand{}, errors.New("err"))
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).AnyTimes().Return([]*repo.EquipmentType{}, nil)

				mockRepo.EXPECT().CreateMetricAttrSum(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricAttrSumStand{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail ATTRSUMSTAND -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricAttrSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigAttrSum(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricAttrSumStand{}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{}, nil)
				mockRepo.EXPECT().CreateMetricAttrSum(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricAttrSumStand{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS ATTRSUMSTAND",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricEquipAttrStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricEquipAttrStand{EqType: "string", AttributeName: "string"}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{{Type: "string", Attributes: []*repo.Attribute{{Name: "string", Type: repo.DataTypeInt}}}}, nil)
				mockRepo.EXPECT().CreateMetricEquipAttrStandard(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricEquipAttrStand{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail ATTRSUMSTAND -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricEquipAttrStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricEquipAttrStand{}, errors.New("err"))
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).AnyTimes().Return([]*repo.EquipmentType{}, nil)

				mockRepo.EXPECT().CreateMetricEquipAttrStandard(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricEquipAttrStand{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail ATTRSUMSTAND -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricEquipAttrStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigEquipAttr(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricEquipAttrStand{}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Times(1).Return([]*repo.EquipmentType{}, nil)
				mockRepo.EXPECT().CreateMetricEquipAttrStandard(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricEquipAttrStand{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS USERSUMSTANDARD",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricUserSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUSS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricUSS{}, nil)
				mockRepo.EXPECT().CreateMetricUSS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricUSS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail USERSUMSTANDARD -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricUserSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUSS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricUSS{}, errors.New("err"))
				mockRepo.EXPECT().CreateMetricUSS(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricUSS{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail USERSUMSTANDARD -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricUserSumStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigUSS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricUSS{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricUSS(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricUSS{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - can not find claims in context",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope11",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1_not_found",
					Metric: &v1.Metric{
						Type: repo.MetricOPSOracleProcessorStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS SQLSTANDARD",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSQLStand{}, nil)
				mockRepo.EXPECT().CreateMetricSQLStandard(ctx, gomock.Any()).Times(1).Return(&repo.MetricSQLStand{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail SQLSTANDARD -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSQLStand{}, errors.New("err"))

				mockRepo.EXPECT().CreateMetricSQLStandard(ctx, gomock.Any()).AnyTimes().Return(&repo.MetricSQLStand{}, nil)
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail SQLSTANDARD -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLStandard(ctx, gomock.Any(), gomock.Any()).Times(1).Return(&repo.MetricSQLStand{}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricSQLStandard(ctx, gomock.Any()).Times(1).Return(&repo.MetricSQLStand{}, errors.New("err")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS SQL",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLEnterprise.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil)
				mockRepo.EXPECT().CreateMetricSQLForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail SQL -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLEnterprise.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, errors.New("abc"))
				mockRepo.EXPECT().CreateMetricSQLForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail SQL -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricMicrosoftSQLEnterprise.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigSQLForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil)
				mockRepo.EXPECT().CreateMetricSQLForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, errors.New("error")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS WinDcenter",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerDataCenter.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigDataCenterForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil)
				mockRepo.EXPECT().CreateMetricDataCenterForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail WinDcenter -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerDataCenter.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigDataCenterForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, errors.New("error"))
				mockRepo.EXPECT().CreateMetricDataCenterForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail WinDcenter -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerDataCenter.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigDataCenterForScope(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, nil)
				mockRepo.EXPECT().CreateMetricDataCenterForScope(ctx, gomock.Any()).AnyTimes().Return(&repo.ScopeMetric{}, errors.New("error")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS WinServerStand",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigWindowServerStandard(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, nil)
				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "fail WinServerStand -1",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigWindowServerStandard(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, errors.New("error"))
				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, nil).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "fail WinServerStand -2",
			args: args{
				ctx: ctx,
				req: &v1.CreateMetricRequest{
					SenderScope: "Scope1",
					Metric: &v1.Metric{
						Type: repo.MetricWindowsServerStandard.String(),
						Name: "m1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetMetricConfigWindowServerStandard(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, nil)
				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, gomock.Any()).AnyTimes().Return(&repo.MetricWSS{}, errors.New("error")).AnyTimes()
			},
			want: &v1.CreateMetricResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewMetricServiceServer(rep, nil)
			_, err := s.CreateMetric(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("metricServiceServer.CreateMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
