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

func Test_licenseServiceServer_LicensesForEquipAndMetric(t *testing.T) {
	coresAttr := &repo.Attribute{
		ID:   "1A",
		Type: repo.DataTypeInt,
		Name: "numCores",
	}
	cpuAttr := &repo.Attribute{
		ID:   "1B",
		Type: repo.DataTypeInt,
		Name: "numCPU",
	}
	coreFactorAttr := &repo.Attribute{
		ID:   "1C",
		Type: repo.DataTypeFloat,
		Name: "coreFactor",
	}
	coresAttrSim := &repo.Attribute{
		ID:          "1A",
		Type:        repo.DataTypeInt,
		IsSimulated: true,
		IntVal:      3,
		IntValOld:   1,
		Name:        "numCores",
	}
	cpuAttrSim := &repo.Attribute{
		ID:          "1B",
		Type:        repo.DataTypeInt,
		IsSimulated: true,
		IntVal:      2,
		IntValOld:   1,
		Name:        "numCPU",
	}
	coreFactorAttrSim := &repo.Attribute{
		ID:          "1C",
		Type:        repo.DataTypeFloat,
		IsSimulated: true,
		FloatVal:    0.25,
		FloatValOld: 1,
		Name:        "coreFactor",
	}
	envAttrSim := &repo.Attribute{
		ID:           "1D",
		Type:         repo.DataTypeString,
		IsSimulated:  true,
		StringVal:    "Production",
		StringValOld: "Development",
		Name:         "environment",
	}
	serverEquipment := &repo.EquipmentType{
		ID:       "2",
		Type:     "Server",
		ParentID: "3",
		Attributes: []*repo.Attribute{
			coresAttr,
			cpuAttr,
			coreFactorAttr,
		},
	}
	serverEquipmentWithEnvAttr := &repo.EquipmentType{
		ID:       "2",
		Type:     "Server",
		ParentID: "3",
		Attributes: []*repo.Attribute{
			coresAttr,
			cpuAttr,
			coreFactorAttr,
			envAttrSim,
		},
	}
	clusterEquipment := &repo.EquipmentType{
		ID:       "3",
		Type:     "Cluster",
		ParentID: "4",
	}
	eqTypeTree := []*repo.EquipmentType{
		{
			ID:       "1",
			Type:     "Partition",
			ParentID: "2",
		},
		serverEquipment,
		clusterEquipment,
		{
			ID:       "4",
			Type:     "Vcenter",
			ParentID: "5",
			Attributes: []*repo.Attribute{
				{
					ID:   "1A",
					Type: repo.DataTypeString,
					Name: "environment",
				},
			},
		},
		{
			ID:   "5",
			Type: "Datacenter",
		},
	}
	eqTypeTreeForEnvOnSameLevel := []*repo.EquipmentType{
		{
			ID:       "1",
			Type:     "Partition",
			ParentID: "2",
		},
		serverEquipmentWithEnvAttr,
		clusterEquipment,
		{
			ID:       "4",
			Type:     "Vcenter",
			ParentID: "5",
			Attributes: []*repo.Attribute{
				{
					ID:   "1A",
					Type: repo.DataTypeString,
					Name: "environment",
				},
			},
		},
		{
			ID:   "5",
			Type: "Datacenter",
		},
	}
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.LicensesForEquipAndMetricRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.LicensesForEquipAndMetricResponse
		wantErr bool
	}{
		{name: "SUCCESS - For OPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(350), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(100), 100.5, nil)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "SUCCESS OPS - Atleast one attribute is not simulable",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 6,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: false,
							DataType:  v1.DataTypes_INT,
							// Val: &v1.Attribute_IntVal{
							// 	IntVal: 1,
							// },
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(350), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(100), 100.5, nil)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "oracle.processor.standard",
						OldLicences: int64(350),
						NewLicenses: int64(351),
						Delta:       int64(1),
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "FAILURE - For OPS metric - cannot simulate OPS metric for types other than base type",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Cluster",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE-cannot fetch OPS metrics",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE-metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "windows.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE-cannot fetch computed metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "2",
						BaseEqTypeID:          "1",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE-Simulation not allowed for equipment other than base equipment type",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "3",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILURE-equipment does not exist",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(nil, repo.ErrNodeNotFound)
			},
			wantErr: true,
		},
		{
			name: "FAILURE-can not fetch equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE-cannot fetch products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE- for OPS metric - no data for products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(nil, repo.ErrNoData)
			},
			want: &v1.LicensesForEquipAndMetricResponse{},
		},
		{name: "FAILURE-cannot fetch old licenses for OPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name: "Oracle1",
					},
					{
						Name: "Oracle2",
					},
					{
						Name: "Oracle3",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(0), errors.New("Internal"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE-cannot fetch new licenses for OPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricOPS(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricOPS{
					{
						ID:                    "1M",
						Name:                  "oracle.processor.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name: "Oracle1",
					},
					{
						Name: "Oracle2",
					},
					{
						Name: "Oracle3",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(350), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						Name:           "oracle.processor.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(0), float64(0), errors.New("Internal"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE-Metric is not supported for simulation",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: "NoNameMetric",
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - For NUP metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(2000), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(1000), 1000.5, nil)
				gomock.InOrder(
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O1", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "1",
							UserID:    "U1",
							UserCount: int64(100000),
						},
						{
							ID:        "2",
							UserID:    "U2",
							UserCount: int64(200000),
						},
					}, nil),
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O2", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "3",
							UserID:    "U3",
							UserCount: int64(100000),
						},
						{
							ID:        "4",
							UserID:    "U4",
							UserCount: int64(200000),
						},
					}, nil),
				)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: int64(400000),
						NewLicenses: int64(400200),
						Delta:       int64(200),
						SwidTag:     "O1",
					},
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: int64(400000),
						NewLicenses: int64(400200),
						Delta:       int64(200),
						SwidTag:     "O2",
					},
				},
			},
		},
		{name: "SUCCESS NUP - Atleast one attribute is non simulable",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 6,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: false,
							DataType:  v1.DataTypes_INT,
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(2000), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(1000), 1000.5, nil)
				gomock.InOrder(
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O1", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr: &repo.Attribute{
							ID:          "1A",
							Type:        repo.DataTypeInt,
							IsSimulated: false,
							IntValOld:   1,
							Name:        "numCores",
						},
						NumCPUAttr: &repo.Attribute{
							ID:          "1B",
							Type:        repo.DataTypeInt,
							IsSimulated: true,
							IntVal:      6,
							IntValOld:   1,
							Name:        "numCPU",
						},
						NumOfUsers: 100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "1",
							UserID:    "U1",
							UserCount: int64(100000),
						},
						{
							ID:        "2",
							UserID:    "U2",
							UserCount: int64(200000),
						},
					}, nil),
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O2", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr: &repo.Attribute{
							ID:          "1A",
							Type:        repo.DataTypeInt,
							IsSimulated: false,
							IntValOld:   1,
							Name:        "numCores",
						},
						NumCPUAttr: &repo.Attribute{
							ID:          "1B",
							Type:        repo.DataTypeInt,
							IsSimulated: true,
							IntVal:      6,
							IntValOld:   1,
							Name:        "numCPU",
						},
						NumOfUsers: 100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "3",
							UserID:    "U3",
							UserCount: int64(100000),
						},
						{
							ID:        "4",
							UserID:    "U4",
							UserCount: int64(200000),
						},
					}, nil),
				)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: int64(400000),
						NewLicenses: int64(400200),
						Delta:       int64(200),
						SwidTag:     "O1",
					},
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: int64(400000),
						NewLicenses: int64(400200),
						Delta:       int64(200),
						SwidTag:     "O2",
					},
				},
			},
		},
		{name: "SUCCESS - For NUP metric product does not have user nodes",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(2000), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(1000), 1000.5, nil)
				gomock.InOrder(
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O1", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "1",
							UserID:    "U1",
							UserCount: int64(100000),
						},
						{
							ID:        "2",
							UserID:    "U2",
							UserCount: int64(200000),
						},
					}, nil),
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O2", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return(nil, repo.ErrNoData),
				)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: 400000,
						NewLicenses: 400200,
						Delta:       200,
						SwidTag:     "O1",
					},
					{
						MetricName:  "oracle.nup.standard",
						OldLicences: 200000,
						NewLicenses: 200100,
						Delta:       100,
						SwidTag:     "O2",
					},
				},
			},
		},
		{name: "Failure - Getting NUP metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric not found",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard_xyz",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric cannot get computed nup metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "10",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - For NUP metric - cannot simulate NUP metric for types other than base type",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Cluster",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric equipment not found",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(nil, repo.ErrNodeNotFound)
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric failed to fetch parents",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric equipment not found",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(nil, repo.ErrNodeNotFound)
			},
			wantErr: true,
		},
		{name: "FAILURE - For NUP metric - cannot fetch products",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - For NUP metric - no data for products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return(nil, repo.ErrNoData)
			},
			want: &v1.LicensesForEquipAndMetricResponse{},
		},
		{name: "Failure - For NUP metric failed in getting old licenses",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(0), errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure - For NUP metric getting simulated licenses",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(2000), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(0), float64(0), errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "Failure - For NUP metric getting user nodes for product",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOracleNUPStandard.String(),
					MetricName: "oracle.nup.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricNUP(ctx, []string{"Scope1"}).Times(1).Return([]*repo.MetricNUPOracle{
					{
						ID:                    "1M",
						Name:                  "oracle.nup.standard",
						NumCoreAttrID:         "1A",
						NumCPUAttrID:          "1B",
						CoreFactorAttrID:      "1C",
						StartEqTypeID:         "1",
						BaseEqTypeID:          "2",
						AggerateLevelEqTypeID: "3",
						EndEqTypeID:           "5",
						NumberOfUsers:         100,
					},
				}, nil)
				mockLicense.EXPECT().ParentsHirerachyForEquipment(ctx, "e1ID", "Server", uint8(4), []string{"Scope1"}).Times(1).Return(
					&repo.Equipment{
						Type:    "Server",
						EquipID: "e1ID",
						Parent: &repo.Equipment{
							Type:    "Cluster",
							EquipID: "e2ID",
							Parent: &repo.Equipment{
								Type:    "Vcenter",
								EquipID: "e3ID",
								Parent: &repo.Equipment{
									Type:    "Datacenter",
									EquipID: "e4ID",
									Parent:  nil,
								},
							},
						},
					}, nil)
				mockLicense.EXPECT().ProductsForEquipmentForMetricOracleNUPStandard(ctx, "e4ID", "Datacenter", uint8(5),
					&repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.ProductData{
					{
						Name:    "Oracle1",
						Swidtag: "O1",
					},
					{
						Name:    "Oracle2",
						Swidtag: "O2",
					},
				}, nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, "e4ID", "Datacenter",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(2000), nil)
				mockLicense.EXPECT().ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, "e2ID", "Cluster",
					&repo.MetricOPSComputed{
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttr,
						NumCoresAttr:   coresAttr,
						NumCPUAttr:     cpuAttr,
					}, []string{"Scope1"}).Times(1).Return(int64(3000), float64(0), nil)
				gomock.InOrder(
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O1", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return([]*repo.User{
						{
							ID:        "1",
							UserID:    "U1",
							UserCount: int64(100000),
						},
						{
							ID:        "2",
							UserID:    "U2",
							UserCount: int64(200000),
						},
					}, nil),
					mockLicense.EXPECT().UsersForEquipmentForMetricOracleNUP(ctx, "e4ID", "Datacenter", "O2", uint8(5), &repo.MetricNUPComputed{
						Name:           "oracle.nup.standard",
						EqTypeTree:     eqTypeTree,
						BaseType:       serverEquipment,
						AggregateLevel: clusterEquipment,
						CoreFactorAttr: coreFactorAttrSim,
						NumCoresAttr:   coresAttrSim,
						NumCPUAttr:     cpuAttrSim,
						NumOfUsers:     100,
					}, []string{"Scope1"}).Times(1).Return(nil, errors.New("no data")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE-cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-requested scopes are outside the scope of user",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type does not exist",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server1",
					EquipId:    "e1ID",
					MetricType: repo.MetricOPSOracleProcessorStandard.String(),
					MetricName: "oracle.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 0.25,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - For IPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						BaseEqTypeID:     "2",
						CoreFactorAttrID: "1C",
						NumCPUAttrID:     "1B",
						NumCoreAttrID:    "1A",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricIPSStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricIPSComputed{
					Name:     "ibm.pvu.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCPUAttr: &repo.Attribute{
						ID:          "1B",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      2,
						IntValOld:   1,
						Name:        "numCPU",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "ibm.pvu.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "ibm.pvu.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "ibm.pvu.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "FAILURE - For IPS metric - cannot fetch IPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For IPS metric - metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "abc",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For IPS metric -cannot compute IPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "6",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For IPS metric - cannot simulate IPS metric for types other than base type",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Cluster",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For IPS metric - cannot fetch products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricIPSStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricIPSComputed{
					Name:     "ibm.pvu.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
				}, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For IPS metric - no data for products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricIPSIbmPvuStandard.String(),
					MetricName: "ibm.pvu.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricIPS(ctx, []string{"Scope1"}).Return([]*repo.MetricIPS{
					{
						ID:               "1M",
						Name:             "ibm.pvu.standard",
						NumCoreAttrID:    "1A",
						NumCPUAttrID:     "1B",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricIPSStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricIPSComputed{
					Name:     "ibm.pvu.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCPUAttr: &repo.Attribute{
						ID:          "1B",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      2,
						IntValOld:   1,
						Name:        "numCPU",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
				}, []string{"Scope1"}).Return(nil, repo.ErrNoData).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{},
		},
		{name: "SUCCESS - For SPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						NumCPUAttrID:     "1B",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricSAGStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricSPSComputed{
					Name:     "sag.processor.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					NumCPUAttr: &repo.Attribute{
						ID:          "1B",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      2,
						IntValOld:   1,
						Name:        "numCPU",
					},
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "sag.processor.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "sag.processor.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "sag.processor.standard",
						OldLicences: 7,
						NewLicenses: 3,
						Delta:       -4,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "FAILURE - For SPS metric - cannot fetch SPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For SPS metric - metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "abc",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For SPS metric -cannot compute SPS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "6",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For SPS metric - cannot simulate SPS metric for types other than base type",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Cluster",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For SPS metric - cannot fetch products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricSAGStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricSPSComputed{
					Name:     "sag.processor.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
				}, []string{"Scope1"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - For SPS metric - no data for products for equipment",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricSPSSagProcessorStandard.String(),
					MetricName: "sag.processor.standard",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricSPS(ctx, []string{"Scope1"}).Return([]*repo.MetricSPS{
					{
						ID:               "1M",
						Name:             "sag.processor.standard",
						NumCoreAttrID:    "1A",
						NumCPUAttrID:     "1B",
						CoreFactorAttrID: "1C",
						BaseEqTypeID:     "2",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricSAGStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricSPSComputed{
					Name:     "sag.processor.standard",
					BaseType: serverEquipment,
					CoreFactorAttr: &repo.Attribute{
						ID:          "1C",
						Type:        repo.DataTypeFloat,
						IsSimulated: true,
						FloatVal:    1.25,
						FloatValOld: 1.5,
						Name:        "coreFactor",
					},
					NumCoresAttr: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					NumCPUAttr: &repo.Attribute{
						ID:          "1B",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      2,
						IntValOld:   1,
						Name:        "numCPU",
					},
				}, []string{"Scope1"}).Return(nil, repo.ErrNoData).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{},
		},
		{name: "SUCCESS - For Attr Sum metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricAttrSumStandard.String(),
					MetricName: "attrsum1",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricAttrSum(ctx, []string{"Scope1"}).Return([]*repo.MetricAttrSumStand{
					{
						ID:             "1M",
						Name:           "attrsum1",
						EqType:         "Server",
						AttributeName:  "numCores",
						ReferenceValue: 2,
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricAttrSumStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricAttrSumStandComputed{
					Name:     "attrsum1",
					BaseType: serverEquipment,
					Attribute: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					ReferenceValue: 2,
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "attrsum1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "attrsum1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "attrsum1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "SUCCESS - For ACS metric",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricAttrCounterStandard.String(),
					MetricName: "attr1",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricACS(ctx, []string{"Scope1"}).Return([]*repo.MetricACS{
					{
						ID:            "1M",
						Name:          "attr1",
						EqType:        "Server",
						AttributeName: "numCores",
						Value:         "3",
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricAttrCounterStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricACSComputed{
					Name:     "attr1",
					BaseType: serverEquipment,
					Attribute: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					Value: "3",
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "attr1",
						OldLicences: 0,
						NewLicenses: 1,
						Delta:       1,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "attr1",
						OldLicences: 0,
						NewLicenses: 1,
						Delta:       1,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "attr1",
						OldLicences: 0,
						NewLicenses: 1,
						Delta:       1,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "SUCCESS - For Equip Sum metric - environment attribute is on parent",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricEquipAttrStandard.String(),
					MetricName: "equipattr1",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTree, nil)
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, []string{"Scope1"}).Return([]*repo.MetricEquipAttrStand{
					{
						ID:            "1M",
						Name:          "equipattr1",
						EqType:        "Server",
						Environment:   "Production",
						AttributeName: "numCores",
						Value:         2,
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricEquipAttrStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricEquipAttrStandComputed{
					Name:     "equipattr1",
					BaseType: serverEquipment,
					EqTypeTree: []*repo.EquipmentType{
						serverEquipment,
						clusterEquipment,
						{
							ID:       "4",
							Type:     "Vcenter",
							ParentID: "5",
							Attributes: []*repo.Attribute{
								{
									ID:   "1A",
									Type: repo.DataTypeString,
									Name: "environment",
								},
							},
						},
					},
					Environment: "Production",
					Attribute: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					Value: 2,
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "equipattr1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "equipattr1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "equipattr1",
						OldLicences: 3,
						NewLicenses: 2,
						Delta:       -1,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
		{name: "SUCCESS - For Equip Sum metric - change in env attribute",
			args: args{
				ctx: ctx,
				req: &v1.LicensesForEquipAndMetricRequest{
					EquipType:  "Server",
					EquipId:    "e1ID",
					MetricType: repo.MetricEquipAttrStandard.String(),
					MetricName: "equipattr1",
					Attributes: []*v1.Attribute{
						{
							ID:        "1C",
							Name:      "coreFactor",
							Simulated: true,
							DataType:  v1.DataTypes_FLOAT,
							Val: &v1.Attribute_FloatVal{
								FloatVal: 1.25,
							},
							OldVal: &v1.Attribute_FloatValOld{
								FloatValOld: 1.5,
							},
						},
						{
							ID:        "1B",
							Name:      "numCPU",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 2,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 1,
							},
						},
						{
							ID:        "1A",
							Name:      "numCores",
							Simulated: true,
							DataType:  v1.DataTypes_INT,
							Val: &v1.Attribute_IntVal{
								IntVal: 3,
							},
							OldVal: &v1.Attribute_IntValOld{
								IntValOld: 5,
							},
						},
						{
							ID:        "1D",
							Name:      "environment",
							Simulated: true,
							DataType:  v1.DataTypes_STRING,
							Val: &v1.Attribute_StringVal{
								StringVal: "Production",
							},
							OldVal: &v1.Attribute_StringValOld{
								StringValOld: "Development",
							},
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypeTreeForEnvOnSameLevel, nil)
				mockLicense.EXPECT().ListMetricEquipAttr(ctx, []string{"Scope1"}).Return([]*repo.MetricEquipAttrStand{
					{
						ID:            "1M",
						Name:          "equipattr1",
						EqType:        "Server",
						Environment:   "Production",
						AttributeName: "numCores",
						Value:         2,
					},
				}, nil).Times(1)
				mockLicense.EXPECT().ProductsForEquipmentForMetricEquipAttrStandard(ctx, "e1ID", "Server", uint8(1), &repo.MetricEquipAttrStandComputed{
					Name:     "equipattr1",
					BaseType: serverEquipmentWithEnvAttr,
					EqTypeTree: []*repo.EquipmentType{
						serverEquipmentWithEnvAttr,
					},
					Environment: "Production",
					Attribute: &repo.Attribute{
						ID:          "1A",
						Type:        repo.DataTypeInt,
						IsSimulated: true,
						IntVal:      3,
						IntValOld:   5,
						Name:        "numCores",
					},
					Value: 2,
				}, []string{"Scope1"}).Return([]*repo.ProductData{
					{
						Swidtag: "Oracle1",
					},
					{
						Swidtag: "Oracle2",
					},
					{
						Swidtag: "Oracle3",
					},
				}, nil).Times(1)
			},
			want: &v1.LicensesForEquipAndMetricResponse{
				Licenses: []*v1.ProductLicenseForEquipAndMetric{
					{
						MetricName:  "equipattr1",
						OldLicences: 0,
						NewLicenses: 2,
						Delta:       2,
						SwidTag:     "Oracle1",
					},
					{
						MetricName:  "equipattr1",
						OldLicences: 0,
						NewLicenses: 2,
						Delta:       2,
						SwidTag:     "Oracle2",
					},
					{
						MetricName:  "equipattr1",
						OldLicences: 0,
						NewLicenses: 2,
						Delta:       2,
						SwidTag:     "Oracle3",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.LicensesForEquipAndMetric(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.LicensesForEquipAndMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareLicensesForEquipAndMetricResponse(t, "LicensesForEquipAndMetric", tt.want, got)
			}
		})
	}
}

func compareLicensesForEquipAndMetricResponse(t *testing.T, name string, exp, act *v1.LicensesForEquipAndMetricResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	for i := range exp.Licenses {
		compareLicensesForEquipAndMetric(t, fmt.Sprintf("%s Licenses[%d]", name, i), exp.Licenses[i], act.Licenses[i])
	}

}

func compareLicensesForEquipAndMetric(t *testing.T, name string, exp, act *v1.ProductLicenseForEquipAndMetric) {
	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.MetricName are not same", name)
	assert.Equalf(t, exp.OldLicences, act.OldLicences, "%s.OldLicences are not same", name)
	assert.Equalf(t, exp.NewLicenses, act.NewLicenses, "%s.NewLicenses are not same", name)
	assert.Equalf(t, exp.Delta, act.Delta, "%s.Delta are not same", name)
	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
}
