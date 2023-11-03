package v1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	prov1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	mockpro "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1/mock"
)

func Test_licenseServiceServer_ListAcqRightsForAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient

	type args struct {
		ctx context.Context
		req *v1.ListAcqRightsForAggregationRequest
	}
	metrics := []*repo.Metric{
		{
			Name: "OPS",
			Type: "oracle.processor.standard",
		},
		{
			Name: "NUP",
			Type: "oracle.nup.standard",
		},
		{
			Name: "WS",
			Type: "oracle.processor.standard",
		},
		{
			Name: "WSD",
			Type: repo.MetricWindowsServerDataCenter,
		},
		{
			Name: "WSS",
			Type: repo.MetricWindowsServerStandard,
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
			Name: "MSE",
			Type: repo.MetricMicrosoftSqlEnterprise,
		},
		{
			Name: "MSS",
			Type: repo.MetricUserSumStandard,
		},
		{
			Name: "UCS",
			Type: repo.MetricUserConcurentStandard,
		},
		{
			Name: "UNS",
			Type: repo.MetricUserNomStandard,
		},
		{
			Name: "attribute.sum.standard",
			Type: repo.MetricAttrSumStandard,
		},
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcqRightsForAggregationResponse
		wantErr bool
	}{
		// {
		// 	name: "SUCCESS - metric type OPS",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ListAcqRightsForAggregationRequest{
		// 			Name:       "OPS",
		// 			Scope:      "Scope1",
		// 			Simulation: true,
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockLicense := mock.NewMockLicense(mockCtrl)
		// 		rep = mockLicense
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
		// 		mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
		// 			ID:                1,
		// 			Name:              "OPS",
		// 			ProductNames:      []string{"p1,p2"},
		// 			Swidtags:          []string{"Swid1", "Swid2"},
		// 			ProductIDs:        []string{"PR1", "PR2"},
		// 			Editor:            "e1",
		// 			NumOfApplications: 3,
		// 			NumOfEquipments:   4,
		// 		}, []*repo.ProductAcquiredRight{
		// 			{
		// 				SKU:               "s1",
		// 				Metric:            "OPS",
		// 				AcqLicenses:       1197,
		// 				TotalCost:         20,
		// 				TotalPurchaseCost: 20,
		// 				AvgUnitPrice:      4,
		// 				Repartition:       true,
		// 			},
		// 			{
		// 				SKU:               "s1.1",
		// 				Metric:            "OPS",
		// 				AcqLicenses:       1197,
		// 				TotalCost:         20,
		// 				TotalPurchaseCost: 20,
		// 				AvgUnitPrice:      4,
		// 				Repartition:       true,
		// 			},
		// 			{
		// 				SKU:               "s2",
		// 				Metric:            "OPS",
		// 				AcqLicenses:       1197,
		// 				TotalCost:         20,
		// 				TotalPurchaseCost: 20,
		// 				AvgUnitPrice:      4,
		// 				Repartition:       false,
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
		// 			Attributes: []*repo.Attribute{cores, cpu, corefactor},
		// 		}
		// 		start := &repo.EquipmentType{
		// 			ID:       "e1",
		// 			ParentID: "e2",
		// 		}
		// 		agg := &repo.EquipmentType{
		// 			ID:       "e3",
		// 			ParentID: "e4",
		// 		}
		// 		end := &repo.EquipmentType{
		// 			ID:       "e4",
		// 			ParentID: "e5",
		// 		}
		// 		endP := &repo.EquipmentType{
		// 			ID: "e5",
		// 		}
		// 		mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
		// 		f := mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
		// 		mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes().After(f)

		// 		mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

		// 		mockLicense.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
		// 			{
		// 				ID:                    "1",
		// 				Name:                  "OPS",
		// 				NumCoreAttrID:         "1A",
		// 				NumCPUAttrID:          "1B",
		// 				CoreFactorAttrID:      "1C",
		// 				StartEqTypeID:         "1",
		// 				BaseEqTypeID:          "2",
		// 				AggerateLevelEqTypeID: "3",
		// 				EndEqTypeID:           "5",
		// 			},
		// 		}, nil).AnyTimes()
		// 	},
		// 	want: &v1.ListAcqRightsForAggregationResponse{
		// 		AcqRights: []*v1.AggregationAcquiredRights{
		// 			{
		// 				SKU:              "s1",
		// 				AggregationName:  "agg",
		// 				SwidTags:         "Swid1,Swid2",
		// 				Metric:           "OPS",
		// 				NumCptLicences:   1200,
		// 				NumAcqLicences:   1197,
		// 				TotalCost:        20,
		// 				DeltaNumber:      0,
		// 				DeltaCost:        0,
		// 				AvgUnitPrice:     4,
		// 				ComputedDetails:  "OPS",
		// 				MetricNotDefined: false,
		// 				NotDeployed:      false,
		// 				ProductNames:     "p1,p2",
		// 				ComputedCost:     0,
		// 				PurchaseCost:     20,
		// 			},
		// 		},
		// 	},
		// 	wantErr: false,
		// },
		{
			name: "SUCCESS - metric type attribute.sum.standard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "attribute.sum.standard",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s1.1",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s2",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				f := mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "attribute.sum.standard"}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "attribute.sum.standard"}, nil).AnyTimes().After(f)

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricAttrSum(ctx, gomock.Any()).Return([]*repo.MetricAttrSumStand{
					{
						ID:             "1",
						Name:           "attribute.sum.standard",
						EqType:         "e2",
						AttributeName:  "att",
						ReferenceValue: 1.0,
					},
				}, nil).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "SUCCESS - metric type microsoft.sql.standard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "MSQ",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "MSQ",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "tf",
					},
					{
						SKU:               "s1.1",
						Metric:            "MSQ",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "tf",
					},
					{
						SKU:               "s2",
						Metric:            "MSQ",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
						TransformDetails:  "tf",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				f := mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSQ"}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSQ"}, nil).AnyTimes().After(f)

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						ID:                    "1",
						Name:                  "MSQ",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "SUCCESS - metric type NUP Transformed to OPS ",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "NUP",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "NUP is tranformed to OPS",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "NUP"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricNUP(ctx, gomock.Any()).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						TransformMetricName:   "OPS",
						Transform:             true,
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil).AnyTimes()
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "NUP",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "NUP is tranformed to OPS",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type NUP Transformed to OPS ",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "NUP",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "NUP is tranformed to OPS",
					},
					{
						SKU:               "s2",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
						TransformDetails:  "NUP is tranformed to OPS",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricNUP(ctx, gomock.Any()).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						TransformMetricName:   "OPS",
						Transform:             true,
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil).AnyTimes()
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "NUP",
						NumCptLicences:   10,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  " ",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type WSD",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "WSD",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "WSD"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricWSD(ctx, gomock.Any()).Return([]*repo.MetricWSD{
					{
						ID:         "1",
						MetricType: "windows.server.datacenter",
						MetricName: "WSD",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricWSDComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "WSD",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type WSS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "WSS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "WSS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricWSS(ctx, gomock.Any()).Return([]*repo.MetricWSS{
					{
						ID:         "1",
						MetricType: repo.MetricMicrosoftSqlStandard.String(),
						MetricName: "WSS",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricWSSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "WSS",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type MSE",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "MSE",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSE"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricMSE(ctx, gomock.Any()).Return([]*repo.MetricMSE{
					{
						ID:        "1",
						Name:      "MSE",
						Reference: "server",
						Core:      "core_per_processor",
						CPU:       "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricMSEComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "MSE",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type MSQ.MetricMicrosoftSqlStandard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "MSQ",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSQ"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricMSS(ctx, gomock.Any()).Return([]*repo.MetricMSS{
					{
						ID:         "1",
						MetricType: repo.MetricMicrosoftSqlStandard.String(),
						MetricName: "MSQ",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricMSSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "MSQ",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type UCS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "UCS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "UCS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricUCS(ctx, gomock.Any()).Return([]*repo.MetricUCS{
					{
						ID:      "1",
						Name:    "UCS",
						Profile: "All",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), uint64(10), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "UCS",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - metric type UNS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "UNS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "UNS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricUNS(ctx, gomock.Any()).Return([]*repo.MetricUNS{
					{
						ID:      "1",
						Name:    "UNS",
						Profile: "All",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricUNSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), uint64(10), nil)
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "UNS",
						NumCptLicences:   1200,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - cannot fetch metrices  list",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"Scope1"}).AnyTimes().Return(nil, errors.New("test srror"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE - AggregationDetails",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(nil, nil, errors.New("Not found"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE - AggregationDetails when aggregation exists but products not exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, nil, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - EquipmentTypes",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s2",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return(nil, errors.New("not found")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - GetAvailableLicenses",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s2",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil)
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(nil, errors.New("not found")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "Failure - Scope validation",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "ACV",
					Simulation: true,
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "FAILURE - No equipment found",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   0,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s2",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "FAILURE - GetMetric",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockRepo.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s2",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(nil, errors.New("not found"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE - GetMaintenanceBySwidtag",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "NUP",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "NUP is tranformed to OPS",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "NUP"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

			},
			wantErr: true,
		},
		{
			name: "FAILURE - metric type NUP Transformed to OPS but equipment not found ",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   0,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "NUP",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "NUP is tranformed to OPS",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "NUP"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricNUP(ctx, gomock.Any()).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						TransformMetricName:   "OPS",
						Transform:             true,
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						ID:                    "1",
						Name:                  "OPS",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil).AnyTimes()

			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "NUP",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - GetMaintenanceBySwidtag metric type WSD",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "WSD",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "WSD"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				// mockLicense.EXPECT().ListMetricWSD(ctx, gomock.Any()).Return([]*repo.MetricWSD{
				// 	{
				// 		ID:         "1",
				// 		MetricType: "windows.server.datacenter",
				// 		MetricName: "WSD",
				// 		Reference:  "server",
				// 		Core:       "core_per_processor",
				// 		CPU:        "num_of_cores",
				// 	},
				// }, nil).AnyTimes()
				// mockLicense.EXPECT().MetricWSDComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil)
			},

			wantErr: true,
		},
		{
			name: "FAILURE - WSD",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "WSD",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "WSD"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricWSD(ctx, gomock.Any()).Return([]*repo.MetricWSD{
					{
						ID:         "1",
						MetricType: "windows.server.datacenter",
						MetricName: "WSD",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricWSDComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), errors.New("Error while computing licences for WSD Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "WSD",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - metric type WSS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "WSS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "WSS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricWSS(ctx, gomock.Any()).Return([]*repo.MetricWSS{
					{
						ID:         "1",
						MetricType: repo.MetricMicrosoftSqlStandard.String(),
						MetricName: "WSS",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricWSSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), errors.New("Error while computing licences for WSS Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "WSS",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - metric type MSE",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "MSE",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSE"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricMSE(ctx, gomock.Any()).Return([]*repo.MetricMSE{
					{
						ID:        "1",
						Name:      "MSE",
						Reference: "server",
						Core:      "core_per_processor",
						CPU:       "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricMSEComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), errors.New("Error while computing licences for MSE Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "MSE",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - metric type MSQ.MetricMicrosoftSqlStandard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "MSQ",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "MSQ"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricMSS(ctx, gomock.Any()).Return([]*repo.MetricMSS{
					{
						ID:         "1",
						MetricType: repo.MetricMicrosoftSqlStandard.String(),
						MetricName: "MSQ",
						Reference:  "server",
						Core:       "core_per_processor",
						CPU:        "num_of_cores",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricMSSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), errors.New("Error while computing licences for MSQ Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "MSQ",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - metric type UCS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "UCS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "UCS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricUCS(ctx, gomock.Any()).Return([]*repo.MetricUCS{
					{
						ID:      "1",
						Name:    "UCS",
						Profile: "All",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), uint64(0), errors.New("Error while computing licences for UCS Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "UCS",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "FAILURE - metric type UNS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForAggregationRequest{
					Name:       "OPS",
					Scope:      "Scope1",
					Simulation: true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockLicense.EXPECT().ListMetrices(ctx, gomock.Any()).Return(metrics, nil).AnyTimes()
				mockLicense.EXPECT().AggregationDetails(ctx, gomock.Any(), gomock.Any(), true, gomock.Any()).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "OPS",
					ProductNames:      []string{"p1,p2"},
					Swidtags:          []string{"Swid1", "Swid2"},
					ProductIDs:        []string{"PR1", "PR2"},
					Editor:            "e1",
					NumOfApplications: 3,
					NumOfEquipments:   4,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "UNS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
						TransformDetails:  "",
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
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}
				mockLicense.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "UNS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockProdClient.EXPECT().GetMaintenanceBySwidtag(ctx, gomock.Any()).Return(&prov1.GetMaintenanceBySwidtagResponse{AcqLicenses: 100, Success: true, UnitPrice: 50}, nil).AnyTimes()

				mockLicense.EXPECT().ListMetricUNS(ctx, gomock.Any()).Return([]*repo.MetricUNS{
					{
						ID:      "1",
						Name:    "UNS",
						Profile: "All",
					},
				}, nil).AnyTimes()
				mockLicense.EXPECT().MetricUNSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), uint64(0), errors.New("Error while computing licences for UNS Metric"))
			},
			want: &v1.ListAcqRightsForAggregationResponse{
				AcqRights: []*v1.AggregationAcquiredRights{
					{
						SKU:              "s1",
						AggregationName:  "agg",
						SwidTags:         "Swid1,Swid2",
						Metric:           "UNS",
						NumCptLicences:   0,
						NumAcqLicences:   1197,
						TotalCost:        20,
						DeltaNumber:      0,
						DeltaCost:        0,
						AvgUnitPrice:     4,
						ComputedDetails:  "",
						MetricNotDefined: false,
						NotDeployed:      false,
						ProductNames:     "p1,p2",
						ComputedCost:     0,
						PurchaseCost:     20,
					},
				},
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
			_, err := s.ListAcqRightsForAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				fmt.Println("test case passed : [", tt.name, "]")
			}
		})
	}
}

func compareAcqRightforProAggResponse(t *testing.T, name string, exp *v1.ListAcqRightsForAggregationResponse, act *v1.ListAcqRightsForAggregationResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	compareAcqRightforProAggAll(t, name+".AcqRights", exp.AcqRights, act.AcqRights)
}

func compareAcqRightforProAggAll(t *testing.T, name string, exp []*v1.AggregationAcquiredRights, act []*v1.AggregationAcquiredRights) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAcqRightforProAgg(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareAcqRightforProAgg(t *testing.T, name string, exp *v1.AggregationAcquiredRights, act *v1.AggregationAcquiredRights) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
	assert.Equalf(t, exp.SwidTags, act.SwidTags, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.NumCptLicences, act.NumCptLicences, "%s.NumCptLicences are not same", name)
	assert.Equalf(t, exp.NumAcqLicences, act.NumAcqLicences, "%s.NumAcqLicences are not same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost are not same", name)
	assert.Equalf(t, exp.DeltaNumber, act.DeltaNumber, "%s.DeltaNumber are not same", name)
	assert.Equalf(t, exp.DeltaCost, act.DeltaCost, "%s.DeltaCost are not same", name)
}

func compareQueryFiltersWithoutOrder(t *testing.T, name string, expFilter []repo.Queryable, actFilter []repo.Queryable) bool {
	for i := range expFilter {
		idx := queryFilterindex(expFilter[i].Key(), actFilter)
		if idx == -1 {
			return false
		}
		if !compareQueryFilter(t, fmt.Sprintf("%s[%d]", name, i), expFilter[i], actFilter[idx]) {
			return false
		}
	}
	return true
}

func queryFilterindex(key string, filter []repo.Queryable) int {
	for i := range filter {
		if key == filter[i].Key() {
			return i
		}
	}
	return -1
}
