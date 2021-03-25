// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	"fmt"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_ListAcqRightsForProductAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.ListAcqRightsForProductAggregationRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcqRightsForProductAggregationResponse
		wantErr bool
	}{
		{name: "SUCCESS - metric type OPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricOPSComputedLicensesAgg(ctx, "pro1", "OPS", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "OPS",
						NumCptLicences: 10,
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
						DeltaNumber:    1187,
						DeltaCost:      2.2558935e+06,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - metric type SPS - licensesProd > licensesNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "SPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricSPSComputedLicensesAgg(ctx, "pro1", "SPS", mat, []string{"Scope1"}).Return(uint64(12), uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "SPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "SPS",
						NumCptLicences: 12,
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
						DeltaNumber:    1185,
						DeltaCost:      2.2520925e+06,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - metric type SPS - licensesProd <= licensesNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "SPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricSPSComputedLicensesAgg(ctx, "pro1", "SPS", mat, []string{"Scope1"}).Return(uint64(8), uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "SPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "SPS",
						NumCptLicences: 10,
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
						DeltaNumber:    1187,
						DeltaCost:      2.2558935e+06,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - metric type ACS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "acs1",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "acs1",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "acs1",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "acs1",
						Type: repo.MetricAttrCounterStandard,
					},
				}, nil).Times(1)
				cores := &repo.Attribute{
					Name: "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					Name: "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "Server",
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricACSComputed{
					Name:      "acs1",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockLicense.EXPECT().MetricACSComputedLicensesAgg(ctx, "pro1", "acs1", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricACS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "acs1",
						EqType:        "Server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					&repo.MetricACS{
						Name:          "acs2",
						EqType:        "Server",
						AttributeName: "cores",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "acs1",
						NumCptLicences: 10,
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
						DeltaNumber:    1187,
						DeltaCost:      2.2558935e+06,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - metric type IPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "IPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "ibm.pvu.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "ibm.pvu.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricIPSComputedLicensesAgg(ctx, "pro1", "IPS", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "IPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "IPS",
						NumCptLicences: 10,
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
						DeltaNumber:    1187,
						DeltaCost:      2.2558935e+06,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - metric name doesnt exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "SPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "SPS",
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - no equipments linked with product",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID:    "proAggID1",
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   0,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
			},
			want: &v1.ListAcqRightsForProductAggregationResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001PROC,ORAC002PROC",
						SwidTag:        "ORAC001,ORAC002",
						Metric:         "OPS",
						NumAcqLicences: 1197,
						TotalCost:      4.0567515e+07,
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - failed to get product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(nil, errors.New(("Internal"))).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch metric OPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricOPSComputedLicensesAgg(ctx, "pro1", "OPS", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "SPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "sag.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricSPSComputedLicensesAgg(ctx, "pro1", "SPS", mat, []string{"Scope1"}).Return(uint64(12), uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch metric IPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "IPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "ibm.pvu.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "ibm.pvu.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricIPSComputedLicensesAgg(ctx, "pro1", "IPS", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot fetch metric ACS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "acs1",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "acs1",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "acs1",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "acs1",
						Type: repo.MetricAttrCounterStandard,
					},
				}, nil).Times(1)
				cores := &repo.Attribute{
					Name: "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					Name: "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "Server",
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricACSComputed{
					Name:      "acs1",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockLicense.EXPECT().MetricACSComputedLicensesAgg(ctx, "pro1", "acs1", mat, []string{"Scope1"}).Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricACS(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRightsForProductAggregation - cannot find metric for computation",
			args: args{
				ctx: ctx,
				req: &v1.ListAcqRightsForProductAggregationRequest{
					ID: "proAggID1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationDetails(ctx, "proAggID1", &repo.QueryProductAggregations{}, []string{"Scope1"}).Return(&repo.ProductAggregation{
					ID:                "proAggID1",
					Name:              "pro1",
					Editor:            "e1",
					Product:           "productName",
					Metric:            "OPS",
					NumOfApplications: 1,
					NumOfEquipments:   1,
					TotalCost:         1000,
					Products:          []string{"Scope1", "Scope2"},
					AcqRightsFull: []*repo.AcquiredRights{
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC001PROC",
							SwidTag:                        "ORAC001",
							ProductName:                    "Oracle Client",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         1016,
							LicensesUnderMaintenanceNumber: 1008,
							AvgLicenesUnitPrice:            2042,
							AvgMaintenanceUnitPrice:        14294,
							TotalPurchaseCost:              2074672,
							TotalMaintenanceCost:           14408352,
							TotalCost:                      35155072,
						},
						&repo.AcquiredRights{
							Entity:                         "",
							SKU:                            "ORAC002PROC",
							SwidTag:                        "ORAC002",
							ProductName:                    "Oracle XML Development Kit",
							Editor:                         "oracle",
							Metric:                         "oracle.processor.standard",
							AcquiredLicensesNumber:         181,
							LicensesUnderMaintenanceNumber: 181,
							AvgLicenesUnitPrice:            1759,
							AvgMaintenanceUnitPrice:        12313,
							TotalPurchaseCost:              318379,
							TotalMaintenanceCost:           2228653,
							TotalCost:                      5412443,
						},
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: "abc",
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).Times(1)
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListAcqRightsForProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareAcqRightforProAggResponse(t, "ListAcqRightsForProductAggregation", got, tt.want)
			} else {
				fmt.Println("test case passed : [", tt.name, "]")
			}
		})
	}
}

func compareAcqRightforProAggResponse(t *testing.T, name string, exp *v1.ListAcqRightsForProductAggregationResponse, act *v1.ListAcqRightsForProductAggregationResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	compareAcqRightforProAggAll(t, name+".AcqRights", exp.AcqRights, act.AcqRights)
}

func compareAcqRightforProAggAll(t *testing.T, name string, exp []*v1.ProductAcquiredRights, act []*v1.ProductAcquiredRights) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAcqRightforProAgg(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareAcqRightforProAgg(t *testing.T, name string, exp *v1.ProductAcquiredRights, act *v1.ProductAcquiredRights) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
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
