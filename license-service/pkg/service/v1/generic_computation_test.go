package v1

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/mock"

	"github.com/golang/mock/gomock"
	prov1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	mockpro "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1/mock"
)

func Test_licenseServiceServer_InsMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient

	type args struct {
		ctx   context.Context
		input map[string]interface{}
		equip []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcqRightsForAggregationResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - metric type OPS",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockCtrl = gomock.NewController(t)
				metrics := []*repo.MetricINM{
					{
						Name: "metric1",
						ID:   "oracle.processor.standard",
					},
				}
				mockLicense.EXPECT().ListMetricINM(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockLicense.EXPECT().MetricINMComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := insMetricCalulation(tt.args.ctx, s, tt.args.equip, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				fmt.Println("test case passed : [", tt.name, "]")
			}
		})
	}
}
func Test_licenseServiceServer_EquipAttrMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient

	type args struct {
		ctx   context.Context
		input map[string]interface{}
		equip []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcqRightsForAggregationResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - metric type eqip ot found",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockCtrl = gomock.NewController(t)
				metrics := []*repo.MetricEquipAttrStand{
					{
						Name:   "metric1",
						ID:     "oracle.processor.standard",
						EqType: "cpu",
					},
				}
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockLicense.EXPECT().MetricEquipAttrComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS ",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "cpu",
						Type:     "cpu",
						SourceID: "cpu",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "cpu",
							},
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "environment",
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockCtrl = gomock.NewController(t)
				metrics := []*repo.MetricEquipAttrStand{
					{
						Name:          "metric1",
						ID:            "oracle.processor.standard",
						EqType:        "cpu",
						AttributeName: "cpu",
					},
				}
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockLicense.EXPECT().MetricEquipAttrComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS err 2 ",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "cpu",
						Type:     "cpu",
						SourceID: "cpu",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "cpu",
							},
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "environment",
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockCtrl = gomock.NewController(t)
				metrics := []*repo.MetricEquipAttrStand{
					{
						Name:          "metric1",
						ID:            "oracle.processor.standard",
						EqType:        "cpu",
						AttributeName: "cpu",
					},
				}
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, gomock.Any()).AnyTimes().Return(metrics, repo.ErrNoData)
				mockLicense.EXPECT().MetricEquipAttrComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
			wantErr: false,
		},
		{
			wantErr: true,
			name:    "SUCCESS err 3 ",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "cpu",
						Type:     "cpu",
						SourceID: "cpu",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "cpu",
							},
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "environment",
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockCtrl = gomock.NewController(t)
				metrics := []*repo.MetricEquipAttrStand{
					{
						Name:          "metric1",
						ID:            "oracle.processor.standard",
						EqType:        "cpu",
						AttributeName: "cpu",
					},
				}
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockLicense.EXPECT().MetricEquipAttrComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), repo.ErrNoData)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := equipAttrMetricCalulation(tt.args.ctx, s, tt.args.equip, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				fmt.Println("test case passed : [", tt.name, "]")
			}
		})
	}
}

func Test_licenseServiceServer_SsMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	metrics := []*repo.MetricSS{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	type args struct {
		ctx   context.Context
		input map[string]interface{}
		equip []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcqRightsForAggregationResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - metric type eqip ot found",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{&repo.EquipmentType{}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricSS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS err",
			args: args{
				ctx: ctx,
				input: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
				},
				equip: []*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "cpu",
						Type:     "cpu",
						SourceID: "cpu",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "cpu",
							},
							&repo.Attribute{
								Type: repo.DataTypeString,
								Name: "environment",
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricSS(ctx, gomock.Any()).AnyTimes().Return(metrics, repo.ErrNoData)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := ssMetricCalulation(tt.args.ctx, s, tt.args.equip, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				fmt.Println("test case passed : [", tt.name, "]")
			}
		})
	}
}
func Test_licenseServiceServer_userSumMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.MetricUserSumStand{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx     context.Context
		req     map[string]interface{}
		eqTypes []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
				eqTypes: []*repo.EquipmentType{&repo.EquipmentType{ID: "s1", Type: "s1"}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUserSum(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUserSumComputedLicensesAgg(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric2",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUserSum(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUserSumComputedLicensesAgg(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - individual iss agg false",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        false,
					"IS_SA":         false,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUserSum(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUserSumComputedLicenses(ctx, gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - err",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUserSum(ctx, gomock.Any()).AnyTimes().Return(metrics, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUserSum(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUserSumComputedLicensesAgg(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), sql.ErrNoRows)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := userSumMetricCalulation(tt.args.ctx, s, tt.args.eqTypes, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
			// }
		})
	}
}

// func Test_licenseServiceServer_InsMetricCalulation(t *testing.T) {
// 	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
// 		UserID: "admin@superuser.com",
// 		Role:   "SuperAdmin",
// 		Socpes: []string{"Scope1"},
// 	})

// 	var mockCtrl *gomock.Controller
// 	var rep repo.License
// 	var prod prov1.ProductServiceClient

//		type args struct {
//			ctx   context.Context
//			input map[string]interface{}
//			equip []*repo.EquipmentType
//		}
//		tests := []struct {
//			name    string
//			s       *licenseServiceServer
//			args    args
//			setup   func()
//			want    *v1.ListAcqRightsForAggregationResponse
//			wantErr bool
//		}{
//			{
//				name: "SUCCESS - metric type OPS",
//				args: args{
//					ctx: ctx,
//					input: map[string]interface{}{
//						"scopes":        []string{"scope1", "scope2"},
//						"METRIC_NAME":   "metric1",
//						"PROD_AGG_NAME": "agg",
//						"IS_AGG":        true,
//					},
//					equip: []*repo.EquipmentType{},
//				},
//				setup: func() {
//					mockCtrl = gomock.NewController(t)
//					mockLicense := mock.NewMockLicense(mockCtrl)
//					rep = mockLicense
//					mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
//					prod = mockProdClient
//					mockCtrl = gomock.NewController(t)
//					metrics := []*repo.MetricINM{
//						{
//							Name: "metric1",
//							ID:   "oracle.processor.standard",
//						},
//					}
//					mockLicense.EXPECT().ListMetricINM(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
//					mockLicense.EXPECT().MetricINMComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
//				},
//				wantErr: false,
//			},
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				tt.setup()
//				s := &licenseServiceServer{
//					licenseRepo:   rep,
//					productClient: prod,
//				}
//				_, err := insMetricCalulation(tt.args.ctx, s, tt.args.equip, tt.args.input)
//				if (err != nil) != tt.wantErr {
//					t.Errorf("licenseServiceServer.ListAcqRightsForAggregation() error = %v, wantErr %v", err, tt.wantErr)
//					return
//				} else {
//					fmt.Println("test case passed : [", tt.name, "]")
//				}
//			})
//		}
//	}
func Test_licenseServiceServer_unsMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.MetricUNS{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx     context.Context
		req     map[string]interface{}
		eqTypes []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
				eqTypes: []*repo.EquipmentType{&repo.EquipmentType{ID: "s1", Type: "s1"}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUNS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUNSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric2",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUNS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUNSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - individual iss agg false",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        false,
					"IS_SA":         false,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUNS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUNSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - err",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUNS(ctx, gomock.Any()).AnyTimes().Return(metrics, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUNS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUNSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), sql.ErrNoRows)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := unsMetricCalulation(tt.args.ctx, s, tt.args.eqTypes, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
			// }
		})
	}
}
func Test_licenseServiceServer_ucsMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.MetricUCS{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx     context.Context
		req     map[string]interface{}
		eqTypes []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
				eqTypes: []*repo.EquipmentType{&repo.EquipmentType{ID: "s1", Type: "s1"}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric2",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - individual iss agg false",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        false,
					"IS_SA":         false,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - err",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), sql.ErrNoRows)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := ucsMetricCalulation(tt.args.ctx, s, tt.args.eqTypes, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
			// }
		})
	}
}

func Test_licenseServiceServer_mseMetricCalulation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.MetricMSE{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx     context.Context
		req     map[string]interface{}
		eqTypes []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
				eqTypes: []*repo.EquipmentType{&repo.EquipmentType{ID: "s1", Type: "s1"}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricMSE(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricMSEComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric2",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricMSE(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricMSEComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - individual iss agg false",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        false,
					"IS_SA":         false,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricMSE(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricMSEComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), nil)
			},
		},
		{name: "SUCCESS - err",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricMSE(ctx, gomock.Any()).AnyTimes().Return(metrics, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricMSE(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricMSEComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), sql.ErrNoRows)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := mseMetricCalulation(tt.args.ctx, s, tt.args.eqTypes, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
			// }
		})
	}
}
