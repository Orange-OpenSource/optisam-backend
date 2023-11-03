package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/mock"
	prov1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	mockpro "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1/mock"
)

func TestAdminRightsRequired(t *testing.T) {
	// Populate adminRPCMap with sample data
	adminRPCMap = map[string]struct{}{
		"ServiceA.MethodA": {},
		"ServiceB.MethodB": {},
		"ServiceC.MethodC": {},
	}

	tests := []struct {
		method      string
		expected    bool
		description string
	}{
		{"ServiceA.MethodA", true, "Admin rights required"},
		{"ServiceB.MethodB", true, "Admin rights required"},
		{"ServiceC.MethodC", true, "Admin rights required"},
		{"ServiceD.MethodD", false, "Method not in adminRPCMap"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := AdminRightsRequired(test.method)
			if result != test.expected {
				t.Errorf("For method %s, expected %v but got %v", test.method, test.expected, result)
			}
		})
	}
}

func Test_licenseServiceServer_ListAcqRightsForApplicationsProduct(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx context.Context
		req *v1.ListAcqRightsForApplicationsProductRequest
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		// want    *v1.ListAcqRightsForApplicationsProductResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
					{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WSD",
						Type: repo.MetricWindowsServerDataCenter,
					},
					{
						Name: "INS",
						Type: repo.MetricInstanceNumberStandard,
					},
					{
						Name: "SS",
						Type: repo.MetricStaticStandard,
					},
					{
						Name: "MSQ",
						Type: repo.MetricMicrosoftSqlStandard,
					},
					{
						Name: "MSS",
						Type: repo.MetricUserSumStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Type:       "server",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:         "e1",
					Type:       "partition",
					ParentID:   "e2",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				agg := &repo.EquipmentType{
					ID:         "e3",
					Type:       "cluster",
					ParentID:   "e4",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				end := &repo.EquipmentType{
					ID:         "e4",
					Type:       "vcenter",
					ParentID:   "e5",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				endP := &repo.EquipmentType{
					ID:         "e5",
					Type:       "datacenter",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
					[]*repo.Equipment{
						{
							ID:      "ue1",
							EquipID: "ee1",
							Type:    "partition",
						},
					}, nil)
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				mockRepo.EXPECT().GetProductsByEditorProductName(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.ProductDetail{}, nil).AnyTimes()
				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformationFromAcqRight(ctx, gomock.Any(), gomock.Any()).Return(&repo.ProductAdditionalInfo{Products: []repo.ProductAdditionalData{{Name: "string"}}}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return([]*repo.MetricOPS{&repo.MetricOPS{Name: "OPS", ID: "e1", StartEqTypeID: "e1", BaseEqTypeID: "e1", AggerateLevelEqTypeID: "e1", EndEqTypeID: "e1", CoreFactorAttrID: "corefactor", NumCoreAttrID: "cores", NumCPUAttrID: "cpus"}}, nil)
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()

			},
		},
		{
			name: "error ProductAcquiredRights ",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
					{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, errors.New("err"))
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WSD",
						Type: repo.MetricWindowsServerDataCenter,
					},
					{
						Name: "INS",
						Type: repo.MetricInstanceNumberStandard,
					},
					{
						Name: "SS",
						Type: repo.MetricStaticStandard,
					},
					{
						Name: "MSQ",
						Type: repo.MetricMicrosoftSqlStandard,
					},
					{
						Name: "MSS",
						Type: repo.MetricUserSumStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Type:       "server",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:         "e1",
					Type:       "partition",
					ParentID:   "e2",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				agg := &repo.EquipmentType{
					ID:         "e3",
					Type:       "cluster",
					ParentID:   "e4",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				end := &repo.EquipmentType{
					ID:         "e4",
					Type:       "vcenter",
					ParentID:   "e5",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				endP := &repo.EquipmentType{
					ID:         "e5",
					Type:       "datacenter",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
					[]*repo.Equipment{
						{
							ID:      "ue1",
							EquipID: "ee1",
							Type:    "partition",
						},
					}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return([]*repo.MetricOPS{&repo.MetricOPS{Name: "OPS", ID: "e1", StartEqTypeID: "e1", BaseEqTypeID: "e1", AggerateLevelEqTypeID: "e1", EndEqTypeID: "e1", CoreFactorAttrID: "corefactor", NumCoreAttrID: "cores", NumCPUAttrID: "cpus"}}, nil)
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()

			},
			wantErr: true,
		},
		{
			name: "scope not found",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "notfound",
				},
			},
			setup: func() {

			},
			wantErr: true,
		},
		{
			name: "fail 6",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, sql.ErrNoRows)

			},
			wantErr: true,
		},
		{
			name: "ProductExistsForApplication_ERROR",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, errors.New("not found"))

			},
			wantErr: true,
		},
		{
			name: "ProductExistsForApplication_FALSE",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)

			},
			wantErr: true,
		},
		{
			name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcqRightsForApplicationsProductRequest{
					AppId:  "a1",
					ProdId: "p1",
					Scope:  "B",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		// {
		// 	name: "Fail",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
		// 		mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ProductAdditionalInfo{
		// 			Products: []repo.ProductAdditionalData{
		// 				{
		// 					NumofEquipments: 56,
		// 				},
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().GetProductsByEditorProductName(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.ProductDetail{}, nil).AnyTimes()
		// 		mockRepo.EXPECT().GetProductInformationFromAcqRight(ctx, gomock.Any(), gomock.Any()).Return(&repo.ProductAdditionalInfo{Products: []repo.ProductAdditionalData{{Name: "string"}}}, nil).AnyTimes()

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)

		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, nil).AnyTimes()
		// 		mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return([]*repo.MetricOPS{&repo.MetricOPS{Name: "OPS", ID: "e1", StartEqTypeID: "e1", BaseEqTypeID: "e1", AggerateLevelEqTypeID: "e1", EndEqTypeID: "e1", CoreFactorAttrID: "corefactor", NumCoreAttrID: "cores", NumCPUAttrID: "cpus"}}, nil)
		// 		mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), errors.New(" Not found")).AnyTimes()
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 2",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)
		// 		mockRepo.EXPECT().GetProductInformationFromAcqRight(ctx, "p1", []string{"A"}).Return(&repo.ProductAdditionalInfo{Products: []repo.ProductAdditionalData{{Name: "string"}}}, nil).AnyTimes()
		// 		mockRepo.EXPECT().GetProductsByEditorProductName(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.ProductDetail{}, nil).AnyTimes()

		// 		mockRepo.EXPECT().GetProductInformation(ctx, "p1", []string{"A"}).AnyTimes().Return(&repo.ProductAdditionalInfo{
		// 			Products: []repo.ProductAdditionalData{
		// 				{
		// 					NumofEquipments: 0,
		// 				},
		// 			},
		// 		}, nil)
		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, nil).AnyTimes()
		// 		mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return(nil, errors.New("Not found"))
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 3",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)

		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, sql.ErrNoRows).AnyTimes()
		// 		mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return([]*repo.MetricOPS{&repo.MetricOPS{Name: "OPS", ID: "e1", StartEqTypeID: "e1", BaseEqTypeID: "e1", AggerateLevelEqTypeID: "e1", EndEqTypeID: "e1", CoreFactorAttrID: "corefactor", NumCoreAttrID: "cores", NumCPUAttrID: "cpus"}}, nil)
		// 		mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()

		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "Fail 4",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)

		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, sql.ErrNoRows).AnyTimes()
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 5",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)

		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, sql.ErrNoRows).AnyTimes()
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 8",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, sql.ErrNoRows)

		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 9",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, sql.ErrNoRows)
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 10",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, sql.ErrNoRows)
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "Fail 10.1",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:          "s1",
		// 				Metric:       "OPS",
		// 				AcqLicenses:  5,
		// 				TotalCost:    20,
		// 				AvgUnitPrice: 4,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, errors.New("no data found"))
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "fail 11",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return("p1", "", []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:              "s1",
		// 				Metric:           "OPS",
		// 				AcqLicenses:      5,
		// 				TotalCost:        20,
		// 				AvgUnitPrice:     4,
		// 				TransformDetails: "not blank",
		// 			},
		// 		}, nil)

		// 		mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "WSD",
		// 				Type: repo.MetricWindowsServerDataCenter,
		// 			},
		// 			{
		// 				Name: "INS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 			{
		// 				Name: "SS",
		// 				Type: repo.MetricStaticStandard,
		// 			},
		// 			{
		// 				Name: "MSQ",
		// 				Type: repo.MetricMicrosoftSqlStandard,
		// 			},
		// 			{
		// 				Name: "MSS",
		// 				Type: repo.MetricUserSumStandard,
		// 			},
		// 		}, nil)

		// 		cores := &repo.Attribute{
		// 			ID:   "cores",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		cpu := &repo.Attribute{
		// 			ID:   "cpus",
		// 			Type: repo.DataTypeInt,
		// 		}
		// 		corefactor := &repo.Attribute{
		// 			ID:   "corefactor",
		// 			Type: repo.DataTypeInt,
		// 		}

		// 		base := &repo.EquipmentType{
		// 			ID:         "e2",
		// 			ParentID:   "e3",
		// 			Type:       "server",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:         "e1",
		// 			Type:       "partition",
		// 			ParentID:   "e2",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:         "e3",
		// 			Type:       "cluster",
		// 			ParentID:   "e4",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:         "e4",
		// 			Type:       "vcenter",
		// 			ParentID:   "e5",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID:         "e5",
		// 			Type:       "datacenter",
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}

		// 		mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).AnyTimes().Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
		// 		mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).AnyTimes().Return(
		// 			[]*repo.Equipment{
		// 				{
		// 					ID:      "ue1",
		// 					EquipID: "ee1",
		// 					Type:    "partition",
		// 				},
		// 			}, nil)

		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any(), gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{}, nil).AnyTimes()
		// 		mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).AnyTimes().Return([]*repo.MetricOPS{&repo.MetricOPS{Name: "OPS", ID: "e1", StartEqTypeID: "e1", BaseEqTypeID: "e1", AggerateLevelEqTypeID: "e1", EndEqTypeID: "e1", CoreFactorAttrID: "corefactor", NumCoreAttrID: "cores", NumCPUAttrID: "cpus"}}, nil)
		// 		mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()

		// 	},
		// },
		// {
		// 	name: "Fail 12",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 			Scope:  "A",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().ProductExistsForApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(true, sql.ErrNoRows)
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "FAILURE: Can not find user claims",
		// 	args: args{
		// 		ctx: context.Background(),
		// 		req: &v1.ListAcqRightsForApplicationsProductRequest{
		// 			AppId:  "a1",
		// 			ProdId: "p1",
		// 		},
		// 	},
		// 	setup:   func() {},
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			// s := NewLicenseServiceServer(rep, nil)
			_, err := s.ListAcqRightsForApplicationsProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForApplicationsProduct() error = %v, wantErr %v , failed test case %s", err, tt.wantErr, tt.name)
			} else {
				fmt.Println("Test case passed  : [", tt.name, "]")
			}
		})
	}
}
