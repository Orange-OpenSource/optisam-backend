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

func Test_licenseServiceServer_ProductLicensesForMetric(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.ProductLicensesForMetricRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ProductLicensesForMetricResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - metric type OPS",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, "Scope1").Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
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
				mockLicense.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricOPSComputedLicenses(ctx, "ID1", mat, "Scope1").Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricOPS(ctx, "Scope1").Times(1).Return([]*repo.MetricOPS{
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
			want: &v1.ProductLicensesForMetricResponse{
				NumCptLicences: 10,
				TotalCost:      1000,
			},
		},
		// TODO : Need to check this test case
		// {
		// 	name: "SUCCESS - metric type SPS - licensesProd > licensesNonProd",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ProductLicensesForMetricRequest{
		// 			SwidTag:    "swidTag1",
		// 			MetricName: "SPS",
		// 			UnitCost:   100,
		// 			Scope:      "Scope1",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockLicense := mock.NewMockLicense(mockCtrl)
		// 		rep = mockLicense
		// 		mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1"}).Return("ID1", nil).Times(1)
		// 		mockLicense.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.Metric{
		// 			&repo.Metric{
		// 				Name: "SPS",
		// 				Type: repo.MetricSPSSagProcessorStandard,
		// 			},
		// 		}, nil).Times(1)
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
		// 		mockLicense.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
		// 		mat := &repo.MetricSPSComputed{
		// 			BaseType:       base,
		// 			NumCoresAttr:   cores,
		// 			CoreFactorAttr: corefactor,
		// 		}
		// 		mockLicense.EXPECT().MetricSPSComputedLicenses(ctx, "ID1", mat, "Scope1").Return(uint64(12), uint64(10), nil).Times(1)
		// 		mockLicense.EXPECT().ListMetricSPS(ctx, "Scope1").Times(1).Return([]*repo.MetricSPS{
		// 			&repo.MetricSPS{
		// 				Name:             "SPS",
		// 				BaseEqTypeID:     "e2",
		// 				CoreFactorAttrID: "corefactor",

		// 			},
		// 			&repo.MetricSPS{
		// 				Name: "IMB",
		// 			},
		// 		}, nil)
		// 	},
		// 	want: &v1.ProductLicensesForMetricResponse{
		// 		NumCptLicences: 12,
		// 		TotalCost:      1200,
		// 	},
		// },
		{
			name: "SUCCESS - metric type SPS - licensesProd < licensesNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "SPS",
					UnitCost:   100,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, "Scope1").Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.Metric{
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricSPSComputedLicenses(ctx, "ID1", mat, "Scope1").Return(uint64(8), uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricSPS(ctx, "Scope1").Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "SPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ProductLicensesForMetricResponse{
				NumCptLicences: 10,
				TotalCost:      1000,
			},
		},
		{
			name: "SUCCESS - metric type IPS",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "IPS",
					UnitCost:   100,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, "Scope1").Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.Metric{
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
				mockLicense.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricIPSComputedLicenses(ctx, "ID1", mat, "Scope1").Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricIPS(ctx, "Scope1").Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "SPS",
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
			want: &v1.ProductLicensesForMetricResponse{
				NumCptLicences: 10,
				TotalCost:      1000,
			},
		},
		{
			name: "SUCCESS - metric type NUP",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "NUP",
					UnitCost:   100,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, "Scope1").Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.Metric{
					&repo.Metric{
						Name: "NUP",
						Type: repo.MetricOracleNUPStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, "Scope1").Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricNUPComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricNUPComputedLicenses(ctx, "ID1", mat, "Scope1").Return(uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricNUP(ctx, "Scope1").Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						Name:                  "NUP",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricNUPOracle{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ProductLicensesForMetricResponse{
				NumCptLicences: 10,
				TotalCost:      1000,
			},
		},
		{
			name: "SUCCESS - computeLicenseACS",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "ORAC001",
					MetricName: "ACS",
					UnitCost:   10,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ProductIDForSwidtag(ctx, "ORAC001", &repo.QueryProducts{}, "Scope1").Return("ID1", nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "ACS",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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

				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, "Scope1").Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "ACS",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					&repo.MetricACS{
						Name:          "ACS1",
						EqType:        "server",
						AttributeName: "cpu",
						Value:         "2",
					},
				}, nil)

				mat := &repo.MetricACSComputed{
					Name:      "ACS",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "ID1", mat, []string{"Scope1"}).Times(1).Return(uint64(10), nil)
			},
			want: &v1.ProductLicensesForMetricResponse{
				NumCptLicences: 10,
				TotalCost:      100,
			},
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot get product id for swid tag",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("", errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - metric name does not exist",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "SPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "IPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil).Times(1)
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot find metric for computation",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot compute OPS licenses",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "OPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricOPSComputedLicenses(ctx, "ID1", mat, []string{"Scope1", "Scope2", "Scope3"}).Return(uint64(0), nil).Times(1)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Intenal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot compute SPS licenses",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "SPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricSPSComputedLicenses(ctx, "ID1", mat, []string{"Scope1", "Scope2", "Scope3"}).Return(uint64(12), uint64(10), nil).Times(1)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1", "Scope2", "Scope3"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot compute IPS licenses",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "IPS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "SPS",
						Type: repo.MetricSPSSagProcessorStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricIPSComputedLicenses(ctx, "ID1", mat, []string{"Scope1", "Scope2", "Scope3"}).Return(uint64(0), errors.New("Internal")).Times(1)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1", "Scope2", "Scope3"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "SPS",
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
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot compute NUP licenses",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "NUP",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "NUP",
						Type: repo.MetricOracleNUPStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricNUPComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockLicense.EXPECT().MetricNUPComputedLicenses(ctx, "ID1", mat, []string{"Scope1", "Scope2", "Scope3"}).Return(uint64(0), nil).Times(1)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Intenal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - ProductLicensesForMetric - cannot compute ACS licenses",
			args: args{
				ctx: ctx,
				req: &v1.ProductLicensesForMetricRequest{
					SwidTag:    "swidTag1",
					MetricName: "ACS",
					UnitCost:   100,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductIDForSwidtag(ctx, "swidTag1", &repo.QueryProducts{}, []string{"Scope1", "Scope2", "Scope3"}).Return("ID1", nil).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.Metric{
					&repo.Metric{
						Name: "NUP",
						Type: repo.MetricOracleNUPStandard,
					},
					&repo.Metric{
						Name: "ACS",
						Type: repo.MetricAttrCounterStandard,
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
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1", "Scope2", "Scope3"}).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil).Times(1)
				mat := &repo.MetricACSComputed{
					Name:      "ACS",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockLicense.EXPECT().MetricACSComputedLicenses(ctx, "ID1", mat, []string{"Scope1", "Scope2", "Scope3"}).Return(uint64(0), errors.New("Internal")).Times(1)
				mockLicense.EXPECT().ListMetricACS(ctx, []string{"Scope1", "Scope2", "Scope3"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "ACS",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					&repo.MetricACS{
						Name:          "ACS1",
						EqType:        "server",
						AttributeName: "cpu",
						Value:         "2",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ProductLicensesForMetric(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ProductLicensesForMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				compareProductLicensesForMetricResponse(t, "ProductLicensesForMetric", got, tt.want)
			} else {
				fmt.Println("Test case passed : [", tt.name, "]")
			}

		})
	}
}

func compareProductLicensesForMetricResponse(t *testing.T, name string, exp *v1.ProductLicensesForMetricResponse, act *v1.ProductLicensesForMetricResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.NumCptLicences, act.NumCptLicences, "%s.NumCptLicences are not same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost are not same", name)
}
