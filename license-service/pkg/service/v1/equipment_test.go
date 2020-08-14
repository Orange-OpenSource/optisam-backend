// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_licenseServiceServer_MetricesForEqType(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.MetricesForEqTypeRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		setup   func()
		args    args
		want    *v1.ListMetricResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							{
								ID: "1A",
							},
							{
								ID: "1B",
							},
							{
								ID: "1C",
							},
						},
					},
					{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							{
								ID: "2A",
							},
							{
								ID: "2B",
							},
							{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						ID:                    "3M",
						Name:                  "MetricNUP1",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
					&repo.MetricNUPOracle{
						ID:                    "4M",
						Name:                  "MetricNUP2",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         200,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						ID:               "5M",
						Name:             "MetricIPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricIPS{
						ID:               "6M",
						Name:             "MetricIPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "3",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						ID:               "7M",
						Name:             "MetricSPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "3",
					},
					&repo.MetricSPS{
						ID:               "8M",
						Name:             "MetricSPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
				}, nil)
			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					&v1.Metric{
						Type:        "oracle.processor.standard",
						Name:        "MetricOPS1",
						Description: "xyz",
					},
					&v1.Metric{
						Type:        "oracle.nup.standard",
						Name:        "MetricNUP2",
						Description: "uvw",
					},
					&v1.Metric{
						Type:        "ibm.pvu.standard",
						Name:        "MetricIPS1",
						Description: "pqr",
					},
					&v1.Metric{
						Type:        "sag.processor.standard",
						Name:        "MetricSPS2",
						Description: "abc",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - With One OPS metric giving error",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "1",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						ID:                    "3M",
						Name:                  "MetricNUP1",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
					&repo.MetricNUPOracle{
						ID:                    "4M",
						Name:                  "MetricNUP2",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         200,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						ID:               "5M",
						Name:             "MetricIPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricIPS{
						ID:               "6M",
						Name:             "MetricIPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						ID:               "7M",
						Name:             "MetricSPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricSPS{
						ID:               "8M",
						Name:             "MetricSPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
				}, nil)

			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					&v1.Metric{
						Type:        "oracle.processor.standard",
						Name:        "MetricOPS1",
						Description: "xyz",
					},
					&v1.Metric{
						Type:        "oracle.nup.standard",
						Name:        "MetricNUP2",
						Description: "uvw",
					},
					&v1.Metric{
						Type:        "ibm.pvu.standard",
						Name:        "MetricIPS1",
						Description: "pqr",
					},
					&v1.Metric{
						Type:        "ibm.pvu.standard",
						Name:        "MetricIPS2",
						Description: "pqr",
					},
					&v1.Metric{
						Type:        "sag.processor.standard",
						Name:        "MetricSPS1",
						Description: "abc",
					},
					&v1.Metric{
						Type:        "sag.processor.standard",
						Name:        "MetricSPS2",
						Description: "abc",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - With one NUP metric failing",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						ID:                    "3M",
						Name:                  "MetricNUP1",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "1",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
					&repo.MetricNUPOracle{
						ID:                    "4M",
						Name:                  "MetricNUP2",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         200,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						ID:               "5M",
						Name:             "MetricIPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricIPS{
						ID:               "6M",
						Name:             "MetricIPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						ID:               "7M",
						Name:             "MetricSPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricSPS{
						ID:               "8M",
						Name:             "MetricSPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
				}, nil)

			},
			want: &v1.ListMetricResponse{
				Metrices: []*v1.Metric{
					&v1.Metric{
						Type:        "oracle.processor.standard",
						Name:        "MetricOPS1",
						Description: "xyz",
					},
					&v1.Metric{
						Type:        "oracle.nup.standard",
						Name:        "MetricNUP2",
						Description: "uvw",
					},
					&v1.Metric{
						Type:        "ibm.pvu.standard",
						Name:        "MetricIPS1",
						Description: "pqr",
					},
					&v1.Metric{
						Type:        "ibm.pvu.standard",
						Name:        "MetricIPS2",
						Description: "pqr",
					},
					&v1.Metric{
						Type:        "sag.processor.standard",
						Name:        "MetricSPS1",
						Description: "abc",
					},
					&v1.Metric{
						Type:        "sag.processor.standard",
						Name:        "MetricSPS2",
						Description: "abc",
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type does not exist",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "abc",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch OPS metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch NUP metrics",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch IPS network",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						ID:                    "3M",
						Name:                  "MetricNUP1",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
					&repo.MetricNUPOracle{
						ID:                    "4M",
						Name:                  "MetricNUP2",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         200,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - can not fetch sps metric",
			args: args{
				ctx: ctx,
				req: &v1.MetricesForEqTypeRequest{
					Type: "Server",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "Partition",
						ParentID: "2",
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "Server",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "1A",
							},
							&repo.Attribute{
								ID: "1B",
							},
							&repo.Attribute{
								ID: "1C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "3",
						Type:     "Cluster",
						ParentID: "4",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID: "2A",
							},
							&repo.Attribute{
								ID: "2B",
							},
							&repo.Attribute{
								ID: "2C",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "4",
						Type:     "Vcenter",
						ParentID: "5",
					},
					&repo.EquipmentType{
						ID:   "5",
						Type: "Datacenter",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						ID:                    "1M",
						Name:                  "MetricOPS1",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
					&repo.MetricOPS{
						ID:                    "2M",
						Name:                  "MetricOPS2",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricNUP(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricNUPOracle{
					&repo.MetricNUPOracle{
						ID:                    "3M",
						Name:                  "MetricNUP1",
						NumCoreAttrID:         "2A",
						NumCPUAttrID:          "2B",
						CoreFactorAttrID:      "2C",
						StartEqTypeID:         "3",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "4",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
					&repo.MetricNUPOracle{
						ID:                    "4M",
						Name:                  "MetricNUP2",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         200,
					},
				}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						ID:               "5M",
						Name:             "MetricIPS1",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "2",
					},
					&repo.MetricIPS{
						ID:               "6M",
						Name:             "MetricIPS2",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1B",
						BaseEqTypeID:     "3",
					},
				}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.MetricesForEqType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.MetricesForEqType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListMetricResponse(t, "MetricesForEqType", tt.want, got)
			}
		})
	}
}
