package v1

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/mock"
	prodv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	prov1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	mockpro "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1/mock"

	"github.com/golang/mock/gomock"
)

func Test_licenseServiceServer_GetOverAllCompliance(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx context.Context
		req *v1.GetOverAllComplianceRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, nil)
				mockRepo.EXPECT().GetAcqRights(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Acqrights{&repo.Acqrights{Swidtag: "s", ProductName: "p1"}}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).AnyTimes().Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "OPS",
						Type: "attribute.sum.standard",
					},
					{
						Name: "OPS",
						Type: repo.MetricInstanceNumberStandard,
					},
				}, nil)

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prodv1.GetAvailableLicensesResponse{}, nil).AnyTimes()

				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockRepo.EXPECT().GetProductsByEditorProductName(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.ProductDetail{}, nil).AnyTimes()

				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().GetProductInformationFromAcqRight(ctx, gomock.Any(), gomock.Any()).Return(&repo.ProductAdditionalInfo{Products: []repo.ProductAdditionalData{{Name: "string"}}}, nil).AnyTimes()

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

				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
				}, nil).AnyTimes()

				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
					Name:           "OPS",
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), mat, gomock.Any()).Return(uint64(8), nil).AnyTimes()
				mockRepo.EXPECT().MetricOPSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()
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
						Metric:            "attribute.sum.standard",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s3",
						Metric:            "instance.number.standard",
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
					{
						SKU:               "s4",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
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
		},
		{
			name: "SUCCESS - fail",
			args: args{
				ctx: ctx,
				req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, sql.ErrNoRows)

			},
			wantErr: true,
		},
		{
			name: "SUCCESS - scope err",
			args: args{
				ctx: ctx,
				req: &v1.GetOverAllComplianceRequest{Scope: "notfound", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, sql.ErrNoRows)

			},
			wantErr: true,
		},
		{
			name: "SUCCESS - clains",
			args: args{
				ctx: context.Background(),
				req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, sql.ErrNoRows)

			},
			wantErr: true,
		},
		{
			name: "SUCCESS - fail",
			args: args{
				ctx: ctx,
				req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, sql.ErrNoRows)

			},
			wantErr: true,
		},
		// {name: "SUCCESS - error2",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: false},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockLicense(mockCtrl)
		// 		rep = mockRepo
		// 		mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
		// 		prod = mockProdClient
		// 		mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, nil)
		// 		mockRepo.EXPECT().GetAcqRights(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Acqrights{&repo.Acqrights{Swidtag: "s", ProductName: "p1"}}, repo.ErrNoData)
		// 		mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).AnyTimes().Return([]*repo.Metric{
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricOPSOracleProcessorStandard,
		// 			},
		// 			{
		// 				Name: "OPS",
		// 				Type: "attribute.sum.standard",
		// 			},
		// 			{
		// 				Name: "OPS",
		// 				Type: repo.MetricInstanceNumberStandard,
		// 			},
		// 		}, nil)
		// 		mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("err")).AnyTimes()
		// 		mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		// 	},
		// 	wantErr: true,
		// },
		{
			name: "SUCCESS - error in other fucn",
			args: args{
				ctx: ctx,
				req: &v1.GetOverAllComplianceRequest{Scope: "scope1", Editor: "e", Simulation: true},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().GetAggregations(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Aggregation{&repo.Aggregation{Name: "p1", Swidtags: []string{"s"}}}, nil)
				mockRepo.EXPECT().GetAcqRights(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]*repo.Acqrights{&repo.Acqrights{Swidtag: "s", ProductName: "p1"}}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, gomock.Any()).AnyTimes().Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "OPS",
						Type: "attribute.sum.standard",
					},
					{
						Name: "OPS",
						Type: repo.MetricInstanceNumberStandard,
					},
				}, nil)
				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prodv1.GetAvailableLicensesResponse{}, errors.New("no data")).AnyTimes()

				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				mockRepo.EXPECT().ProductAcquiredRights(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockRepo.EXPECT().GetProductsByEditorProductName(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*repo.ProductDetail{}, nil).AnyTimes()

				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any(), gomock.Any()).AnyTimes().Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().GetProductInformationFromAcqRight(ctx, gomock.Any(), gomock.Any()).Return(&repo.ProductAdditionalInfo{Products: []repo.ProductAdditionalData{{Name: "string"}}}, nil).AnyTimes()

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

				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
				}, nil).AnyTimes()

				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
					Name:           "OPS",
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, gomock.Any(), mat, gomock.Any()).Return(uint64(8), nil).AnyTimes()
				mockRepo.EXPECT().MetricOPSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(8), nil).AnyTimes()
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
						Metric:            "attribute.sum.standard",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       true,
					},
					{
						SKU:               "s3",
						Metric:            "instance.number.standard",
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
					{
						SKU:               "s4",
						Metric:            "OPS",
						AcqLicenses:       1197,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
						Repartition:       false,
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, gomock.Any()).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).AnyTimes()
				mockProdClient.EXPECT().GetMetric(ctx, gomock.Any()).Return(&prov1.GetMetricResponse{Metric: "OPS"}, nil).AnyTimes()

				mockProdClient.EXPECT().GetAvailableLicenses(ctx, gomock.Any()).Return(&prov1.GetAvailableLicensesResponse{AvailableLicenses: 100}, nil).AnyTimes()

				mockRepo.EXPECT().ListMetricOPS(ctx, gomock.Any()).Return([]*repo.MetricOPS{
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, err := s.GetOverAllCompliance(tt.args.ctx, tt.args.req)
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
