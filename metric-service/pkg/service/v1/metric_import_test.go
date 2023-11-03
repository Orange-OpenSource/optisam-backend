package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"

	equipv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/equipment-service/pkg/api/v1"
	equipmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/equipment-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1/mock"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/golang/mock/gomock"
)

func Test_metricServiceServer_CreateMetricImport(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Metric
	var equip equipv1.EquipmentServiceClient
	metadata := repo.GlobalMetricMetadata("Scope1")

	eqTypes := []*repo.EquipmentType{
		{
			ID:         "1",
			Type:       "server",
			ParentID:   "p1",
			ParentType: "typ_parent",
			Attributes: []*repo.Attribute{
				{
					Name:         "cores_per_processor",
					Type:         repo.DataTypeInt,
					IsSearchable: true,
					IntVal:       10,
				},
				{
					Name:         "hyperthreading",
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
		req *v1.MetricImportRequest
	}
	tests := []struct {
		name   string
		serObj *metricServiceServer
		input  args
		setup  func()
		output *v1.MetricImportResponse
		outErr bool
	}{
		{name: "SUCCESS-INM",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"itnstance.number.sandard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().CreateMetricInstanceNumberStandard(ctx, &repo.MetricINM{
					Name:        metadata["instance.number.standard"].MetadataINM.Name,
					Coefficient: metadata["instance.number.standard"].MetadataINM.Num_Of_Deployments,
					Default:     true,
				}, "Scope1").Return(&repo.MetricINM{
					ID:          "Met_INM1ID",
					Name:        "one_instance",
					Coefficient: 1,
					Default:     true,
				}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "OPS",
					},
				}, nil).AnyTimes()
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-sum_standard",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"static.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricStaticStandard(ctx, &repo.MetricSS{
					Name:           metadata["static.standard"].MetadataSS.Name,
					ReferenceValue: int32(metadata["static.standard"].MetadataSS.Reference),
					Default:        true,
				}, "Scope1").Return(&repo.MetricSS{
					ID:             "Met_SS1ID",
					Name:           "static.standard",
					ReferenceValue: 1,
					Default:        true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-AttrSum_standard",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"attribute.sum.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricAttrSum(ctx, &repo.MetricAttrSumStand{
					Name:           metadata["attribute.sum.standard"].MetadataAttrSum.Name,
					EqType:         metadata["attribute.sum.standard"].MetadataAttrSum.Eq_type,
					AttributeName:  metadata["attribute.sum.standard"].MetadataAttrSum.Attribute_name,
					ReferenceValue: metadata["attribute.sum.standard"].MetadataAttrSum.ReferenceValue,
					Default:        true,
				}, eqTypes[0].Attributes[0], "Scope1").Return(&repo.MetricAttrSumStand{
					ID:             "Met_AttrSum1ID",
					Name:           "attribute.sum",
					EqType:         "server",
					AttributeName:  "cores_per_processor",
					ReferenceValue: 8,
					Default:        true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-ACS",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"attribute.counter.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
						Type: repo.MetricAttrCounterStandard,
					},
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil).AnyTimes()
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Return(eqTypes, nil).Times(1)
				mockRepo.EXPECT().CreateMetricACS(ctx, &repo.MetricACS{
					Name:          metadata["attribute.counter.standard"].MetadataACS.Name,
					EqType:        metadata["attribute.counter.standard"].MetadataACS.Eq_type,
					AttributeName: metadata["attribute.counter.standard"].MetadataACS.Attribute_name,
					Value:         metadata["attribute.counter.standard"].MetadataACS.Value,
					Default:       true,
				}, eqTypes[0].Attributes[1], "Scope1").Return(&repo.MetricACS{
					ID:            "Met_ACS1ID",
					Name:          "attribute.counter",
					EqType:        "server",
					AttributeName: "hyperthreading",
					Value:         "5",
					Default:       true,
				}, nil).AnyTimes()
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).AnyTimes()
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-OPS",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"oracle.processor.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "e1",
							Type:           "virtualmachine",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "a1",
									Name:             "cores_per_processor",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "a2",
									Name:             "server_processors_numbers",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "a3",
									Name:             "oracle_core_factor",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "e2",
							Type:       "server",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
						{
							ID:         "e3",
							Type:       "cluster",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
						{
							ID:         "e4",
							Type:       "vcenter",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				myMap := make(map[string]string)
				myMap["virtualmachine"] = "e1"
				myMap["server"] = "e2"
				myMap["cores_per_processor"] = "a1"
				myMap["server_processors_numbers"] = "a2"
				myMap["oracle_core_factor"] = "a3"
				myMap["cluster"] = "e3"
				myMap["vcenter"] = "e4"
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
				mockRepo.EXPECT().CreateMetricOPS(ctx, &repo.MetricOPS{
					Name:                  metadata["oracle.processor.standard"].MetadataOPS.Name,
					NumCoreAttrID:         myMap[metadata["oracle.processor.standard"].MetadataOPS.Num_core_attr_id],
					NumCPUAttrID:          myMap[metadata["oracle.processor.standard"].MetadataOPS.NumCPU_attr_id],
					CoreFactorAttrID:      myMap[metadata["oracle.processor.standard"].MetadataOPS.Core_factor_attr_id],
					StartEqTypeID:         myMap[metadata["oracle.processor.standard"].MetadataOPS.Start_eq_type_id],
					BaseEqTypeID:          myMap[metadata["oracle.processor.standard"].MetadataOPS.Base_eq_type_id],
					AggerateLevelEqTypeID: myMap[metadata["oracle.processor.standard"].MetadataOPS.AggerateLevel_eq_type_id],
					EndEqTypeID:           myMap[metadata["oracle.processor.standard"].MetadataOPS.End_eq_type_id],
					Default:               true,
				}, "Scope1").Return(&repo.MetricOPS{
					ID:                    "Met_INM1ID",
					Name:                  "oracle.processor",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					BaseEqTypeID:          "e2",
					AggerateLevelEqTypeID: "e3",
					EndEqTypeID:           "e4",
					Default:               true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-NUP",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"oracle.nup.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "e1",
							Type:           "virtualmachine",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "a1",
									Name:             "cores_per_processor",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "a2",
									Name:             "server_processors_numbers",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "a3",
									Name:             "oracle_core_factor",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "e2",
							Type:       "server",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
						{
							ID:         "e3",
							Type:       "cluster",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
						{
							ID:         "e4",
							Type:       "vcenter",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"Scope1"},
							Attributes: []*equipv1.Attribute{
								{},
							},
						},
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, "Scope1").Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "e1",
						ParentID: "e2",
					},
					{
						ID:       "e2",
						ParentID: "e3",
						Attributes: []*repo.Attribute{
							{
								ID:   "a1",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a2",
								Type: repo.DataTypeInt,
							},
							{
								ID:   "a3",
								Type: repo.DataTypeInt,
							},
						},
					},
					{
						ID:       "e3",
						ParentID: "e4",
					},
					{
						ID: "e4",
					},
				}, nil)
				myMap := make(map[string]string)
				myMap["virtualmachine"] = "e1"
				myMap["server"] = "e2"
				myMap["cores_per_processor"] = "a1"
				myMap["server_processors_numbers"] = "a2"
				myMap["oracle_core_factor"] = "a3"
				myMap["cluster"] = "e3"
				myMap["vcenter"] = "e4"
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
				mockRepo.EXPECT().CreateMetricOracleNUPStandard(ctx, &repo.MetricNUPOracle{
					Name:                  metadata["oracle.nup.standard"].MetadataNUP.Name,
					NumCoreAttrID:         myMap[metadata["oracle.nup.standard"].MetadataNUP.Num_core_attr_id],
					NumCPUAttrID:          myMap[metadata["oracle.nup.standard"].MetadataNUP.NumCPU_attr_id],
					CoreFactorAttrID:      myMap[metadata["oracle.nup.standard"].MetadataNUP.Core_factor_attr_id],
					StartEqTypeID:         myMap[metadata["oracle.nup.standard"].MetadataNUP.Start_eq_type_id],
					BaseEqTypeID:          myMap[metadata["oracle.nup.standard"].MetadataNUP.Base_eq_type_id],
					AggerateLevelEqTypeID: myMap[metadata["oracle.nup.standard"].MetadataNUP.AggerateLevel_eq_type_id],
					EndEqTypeID:           myMap[metadata["oracle.nup.standard"].MetadataNUP.End_eq_type_id],
					NumberOfUsers:         metadata["oracle.nup.standard"].MetadataNUP.Number_of_users,
					Transform:             metadata["oracle.nup.standard"].MetadataNUP.Transform,
					TransformMetricName:   metadata["oracle.nup.standard"].MetadataNUP.Transform_metric_name,
					Default:               true,
				}, "Scope1").Return(&repo.MetricNUPOracle{
					ID:                    "Met_INM1ID",
					Name:                  "oracle.nup",
					NumCoreAttrID:         "a1",
					NumCPUAttrID:          "a2",
					CoreFactorAttrID:      "a3",
					StartEqTypeID:         "e1",
					BaseEqTypeID:          "e2",
					AggerateLevelEqTypeID: "e3",
					EndEqTypeID:           "e4",
					NumberOfUsers:         5,
					Transform:             false,
					TransformMetricName:   "",
					Default:               true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-SQL-Standard",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"microsoft.sql.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricSQLStandard(ctx, &repo.MetricSQLStand{
					MetricName: metadata["microsoft.sql.standard"].MetadataSQL.MetricName,
					MetricType: "microsoft.sql.standard",
					Reference:  metadata["microsoft.sql.standard"].MetadataSQL.Reference,
					Core:       metadata["microsoft.sql.standard"].MetadataSQL.Core,
					CPU:        metadata["microsoft.sql.standard"].MetadataSQL.CPU,
					Scope:      "Scope1",
					Default:    true,
				}).Return(&repo.MetricSQLStand{
					ID:         "Met_INM1ID",
					MetricType: "microsoft.sql.standard",
					MetricName: "microsoft.sql.standard",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "Scope1",
					Default:    true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-WindowServer-Standard",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"windows.server.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricWindowServerStandard(ctx, &repo.MetricWSS{
					MetricName: metadata["windows.server.standard"].MetadataSQL.MetricName,
					MetricType: "windows.server.standard",
					Reference:  metadata["windows.server.standard"].MetadataSQL.Reference,
					Core:       metadata["windows.server.standard"].MetadataSQL.Core,
					CPU:        metadata["windows.server.standard"].MetadataSQL.CPU,
					Scope:      "Scope1",
					Default:    true,
				}).Return(&repo.MetricWSS{
					ID:         "Met_INM1ID",
					MetricType: "windows.server.standard",
					MetricName: "windows.server.standard",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "Scope1",
					Default:    true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-UNS",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"user.nominative.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricUserNominativeStandard(ctx, &repo.MetricUNS{
					Name:    metadata["user.nominative.standard"].MetadataUNS.Name,
					Profile: metadata["user.nominative.standard"].MetadataUNS.Profile,
					Default: true,
				}, "Scope1").Return(&repo.MetricUNS{
					ID:      "Met_INM1ID",
					Name:    "user.nominative.standard",
					Profile: "ALL",
					Default: true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-UCS",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"user.concurrent.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
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
				mockRepo.EXPECT().CreateMetricUserConcurentStandard(ctx, &repo.MetricUCS{
					Name:    metadata["user.concurrent.standard"].MetadataUNS.Name,
					Profile: metadata["user.concurrent.standard"].MetadataUNS.Profile,
					Default: true,
				}, "Scope1").Return(&repo.MetricUCS{
					ID:      "Met_INM1ID",
					Name:    "user.concurrent.standard",
					Profile: "ALL",
					Default: true,
				}, nil).Times(1)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Return([]*repo.MetricInfo{
					{
						Name: "ONS",
					},
					{
						Name: "WS",
					},
				}, nil).Times(1)
			},
			output: &v1.MetricImportResponse{
				Success: true,
			},
		},
		{name: "FAILURE - CreateMetricImport - cannot find claims in context",
			input: args{
				ctx: context.Background(),
				req: &v1.MetricImportRequest{
					Metric: []string{"sag.processor.standard"},
					Scope:  "Scope1",
				},
			},
			setup:  func() {},
			outErr: true,
		},
		{name: "FAILURE - CreateMetricImport - cannot fetch eqtypes",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"sag.processor.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(nil, errors.New("service error"))
			},
			outErr: true,
		},
		{name: "FAILURE - cannot fetch metric types",
			input: args{
				ctx: ctx,
				req: &v1.MetricImportRequest{
					Metric: []string{"sag.processor.standard"},
					Scope:  "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockMetric(mockCtrl)
				mockEquipment := equipmock.NewMockEquipmentServiceClient(mockCtrl)
				rep = mockRepo
				equip = mockEquipment
				mockEquipment.EXPECT().EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
					Scopes: []string{"Scope1"},
				}).Times(1).Return(&equipv1.EquipmentTypesResponse{
					EquipmentTypes: []*equipv1.EquipmentType{
						{
							ID:             "1",
							Type:           "typ1",
							ParentId:       "p1",
							MetadataId:     "s1",
							ParentType:     "typ_parent",
							MetadataSource: "equip1.csv",
							Scopes:         []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
								{
									ID:               "2",
									Name:             "attr_2",
									DataType:         equipv1.DataTypes_INT,
									PrimaryKey:       false,
									Displayed:        true,
									Searchable:       false,
									ParentIdentifier: true,
									MappedTo:         "mapping_2",
								},
								{
									ID:               "3",
									Name:             "attr_3",
									DataType:         equipv1.DataTypes_FLOAT,
									PrimaryKey:       false,
									Displayed:        false,
									Searchable:       false,
									ParentIdentifier: false,
									MappedTo:         "mapping_3",
								},
							},
						},
						{
							ID:         "2",
							Type:       "typ2",
							ParentId:   "p2",
							MetadataId: "s2",
							Scopes:     []string{"A"},
							Attributes: []*equipv1.Attribute{
								{
									ID:               "1",
									Name:             "attr_1",
									DataType:         equipv1.DataTypes_STRING,
									PrimaryKey:       true,
									Displayed:        true,
									Searchable:       true,
									ParentIdentifier: true,
									MappedTo:         "mapping_1",
								},
							},
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, "Scope1").Times(1).Return(nil, errors.New("Test error"))
			},
			outErr: true,
		},
		// {name: "FAILURE - CreateMetricInstanceNumberStandard - metric name already exists",
		// 	input: args{
		// 		ctx: ctx,
		// 		req: &v1.MetricINM{
		// 			Name:             "Met_INM1",
		// 			NumOfDeployments: 1,
		// 			Scopes:           []string{"Scope1"},
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockMetric(mockCtrl)
		// 		rep = mockRepo
		// 	},
		// 	outErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.serObj = &metricServiceServer{
				metricRepo: rep,
				equipments: equip,
			}
			got, err := tt.serObj.CreateMetricImport(tt.input.ctx, tt.input.req)
			if (err != nil) != tt.outErr {
				t.Errorf("metricServiceServer.CreateMetricImport() error = %v, wantErr %v", err, tt.outErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("metricServiceServer.CreateMetricImport() = %v, want %v", got, tt.output)
			}
		})
	}
}
