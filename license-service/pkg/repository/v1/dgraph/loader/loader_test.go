// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

// import (
// 	"context"
// 	"encoding/csv"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io/ioutil"
// 	v1 "optisam-backend/license-service/pkg/repository/v1"
// 	"optisam-backend/license-service/pkg/repository/v1/dgraph"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"testing"

// 	"github.com/dgraph-io/dgo/v2/protos/api"
// 	"github.com/stretchr/testify/assert"
// )

// func TestAggregateLoader_Load(t *testing.T) {
// 	// instances, acquired rights
// 	schemaFiles, _ := getAllFilesWithSuffixFullPath("../schema", ".schema")
// 	eqType := &v1.EquipmentType{}
// 	tests := []struct {
// 		name    string
// 		al      *AggregateLoader
// 		setup   func() (func() error, error)
// 		verify  func() error
// 		wantErr bool
// 	}{
// 		{name: "schema - success",
// 			al: &AggregateLoader{
// 				config: &Config{
// 					BadgerDir:    "badger",
// 					Zero:         ":5080",
// 					Alpha:        []string{":9080"},
// 					DropSchema:   true,
// 					CreateSchema: true,
// 					SchemaFiles:  []string{"testdata/schema/files.schema", "testdata/schema/file2.schema"},
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {
// 				predicates := []string{
// 					"product.name",
// 					"product.version",
// 					"application.id",
// 					"application.name",
// 					"equipment.type",
// 				}
// 				wantSchemaNodes := []*SchemaNode{
// 					&SchemaNode{
// 						Predicate: "product.name",
// 						Type:      "string",
// 						Index:     true,
// 						Tokenizer: []string{"trigram"},
// 					},
// 					&SchemaNode{
// 						Predicate: "product.version",
// 						Type:      "int",
// 					},
// 					&SchemaNode{
// 						Predicate: "application.id",
// 						Type:      "string",
// 						Index:     true,
// 						Tokenizer: []string{"exact"},
// 					},
// 					&SchemaNode{
// 						Predicate: "application.name",
// 						Type:      "string",
// 						Index:     true,
// 						Tokenizer: []string{"trigram"},
// 					},
// 					&SchemaNode{
// 						Predicate: "equipment.type",
// 						Type:      "string",
// 						Index:     true,
// 						Tokenizer: []string{"exact"},
// 					},
// 				}
// 				sns, err := querySchema(predicates...)
// 				if err != nil {
// 					return errors.New("error is not expect while quering schema for predicates")
// 				}
// 				compareSchemaNodeAll(t, "schemaNodes", wantSchemaNodes, sns)

// 				return nil
// 			},
// 		},
// 		{name: "metadata - success",
// 			al: &AggregateLoader{
// 				config: &Config{
// 					BadgerDir:    "badger",
// 					Zero:         ":5080",
// 					Alpha:        []string{":9080"},
// 					LoadMetadata: true,
// 					// add schema files
// 					DropSchema:    true,
// 					CreateSchema:  true,
// 					SchemaFiles:   schemaFiles,
// 					ScopeSkeleten: "testdata/skeletonschema",
// 					MetadataFiles: &MetadataFiles{
// 						EquipFiles: []string{"equipment_cluster.csv", "equipment_server.csv"},
// 					},
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {
// 				wantNodes := []*v1.Metadata{
// 					&v1.Metadata{
// 						Source: "equipment_cluster.csv",
// 						Attributes: []string{
// 							"cluster_name",
// 						},
// 					},
// 					&v1.Metadata{
// 						Source: "equipment_server.csv",
// 						Attributes: []string{
// 							"server_hostname",
// 							"server_processorsNumber",
// 							"server_coresNumber",
// 							"parent_hostname",
// 							"corefactor_oracle",
// 							"sag",
// 							"pvu",
// 						},
// 					},
// 				}
// 				q := `{
// 				Metadatas(func: eq(metadata.type,equipment),orderasc: metadata.source) {
// 				   Source:     metadata.source
// 				   Attributes: metadata.attributes
// 				}
// 			  }`

// 				resp, err := dgClient.NewTxn().Query(context.Background(), q)
// 				if err != nil {
// 					return errors.New("Metadata - cannot complete query")
// 				}
// 				type data struct {
// 					Metadatas []*v1.Metadata
// 				}
// 				metadata := data{}

// 				if err := json.Unmarshal(resp.GetJson(), &metadata); err != nil {
// 					return fmt.Errorf("Metadata - cannot unmarshal Json object")
// 				}
// 				compareMetadataAll(t, "metadatas", wantNodes, metadata.Metadatas)
// 				return nil
// 			},
// 		},
// 		{name: "load equipments - success",
// 			al: &AggregateLoader{
// 				config: &Config{
// 					BadgerDir:      "badger",
// 					Zero:           ":5080",
// 					Alpha:          []string{":9080"},
// 					CreateSchema:   true,
// 					LoadEquipments: true,
// 					SchemaFiles:    schemaFiles,
// 					Repository:     dgraph.NewLicenseRepository(dgClient),
// 					ScopeSkeleten:  "testdata/skeletonschema",
// 					MetadataFiles: &MetadataFiles{
// 						EquipFiles: []string{"equipment_server.csv"},
// 					},
// 					Scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					// add equipment files
// 					EquipmentFiles: []string{"equipment_server.csv"},
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				mu := &api.Mutation{
// 					CommitNow: true,
// 					Set: []*api.NQuad{

// 						&api.NQuad{
// 							Subject:     "_:data_source",
// 							Predicate:   "metadata.source",
// 							ObjectValue: stringObjectValue("equipment_server.csv"),
// 						},
// 					},
// 				}

// 				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
// 				if err != nil {
// 					return nil, errors.New("error in transaction of query " + err.Error())
// 				}
// 				sourceID, ok := assigned.Uids["data_source"]
// 				if !ok {
// 					return nil, errors.New("cannot find source id after mutation in setup")
// 				}

// 				eqType = &v1.EquipmentType{
// 					Type:       "Server",
// 					SourceID:   sourceID,
// 					SourceName: "equipment_server.csv",
// 					Attributes: []*v1.Attribute{
// 						&v1.Attribute{
// 							Name:         "server_hostname",
// 							Type:         v1.DataTypeString,
// 							IsSearchable: true,
// 							IsIdentifier: true,
// 							IsDisplayed:  true,
// 							MappedTo:     "server_hostname",
// 						},
// 						&v1.Attribute{
// 							Name:         "server_processorsNumber",
// 							Type:         v1.DataTypeInt,
// 							IsSearchable: true,
// 							IsDisplayed:  true,
// 							MappedTo:     "server_processorsNumber",
// 						},
// 						&v1.Attribute{
// 							Name:        "server_coresNumber",
// 							Type:        v1.DataTypeInt,
// 							IsDisplayed: true,
// 							MappedTo:    "server_coresNumber",
// 						},
// 						&v1.Attribute{
// 							Name:        "parent_hostname",
// 							Type:        v1.DataTypeString,
// 							IsDisplayed: true,
// 							MappedTo:    "parent_hostname",
// 						},
// 						&v1.Attribute{
// 							Name:     "corefactor_oracle",
// 							Type:     v1.DataTypeFloat,
// 							MappedTo: "corefactor_oracle",
// 						},
// 						&v1.Attribute{
// 							Name:     "sag",
// 							Type:     v1.DataTypeFloat,
// 							MappedTo: "sag",
// 						},
// 						&v1.Attribute{
// 							Name:     "pvu",
// 							Type:     v1.DataTypeFloat,
// 							MappedTo: "pvu",
// 						},
// 					},
// 				}

// 				repo := dgraph.NewLicenseRepository(dgClient)
// 				eqType, err = repo.CreateEquipmentType(context.Background(), eqType, []string{})
// 				if err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					fmt.Println("i m here")
// 					//return deleteNodes(sourceID, eqType.ID)
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {

// 				equipments, err := equipmentsJSONFromCSV("testdata/uploaded/scope1/equipment_server.csv", eqType, true)
// 				if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
// 					return err
// 				}

// 				tl, equips, err := dgraph.NewLicenseRepository(dgClient).Equipments(context.Background(), eqType, &v1.QueryEquipments{
// 					PageSize:  3,
// 					Offset:    0,
// 					SortBy:    "server_hostname",
// 					SortOrder: v1.SortASC,
// 				}, []string{"scope1"})

// 				want := int32(5)
// 				want1 := []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]")
// 				assert.Equalf(t, want, tl, "%s.TotalCount should be same")
// 				compareEquipments(t, equips, want1)
// 				return nil
// 			},
// 		},
// 		{name: "load static data - products",
// 			al: &AggregateLoader{
// 				loaders: []loader{
// 					{
// 						load:   loadProducts,
// 						files:  []string{"prod.csv", "products_equipments.csv"}, //product files
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 					{
// 						load:   loadApplications,
// 						files:  []string{"applications.csv", "applications_products.csv"}, //app files
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 				},
// 				config: &Config{
// 					BadgerDir:    "badger",
// 					Zero:         ":5080",
// 					Alpha:        []string{":9080"},
// 					CreateSchema: true,
// 					// add files
// 					SchemaFiles:    schemaFiles,
// 					LoadStaticData: true,
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {

// 				wantProducts := &v1.ProductInfo{
// 					NumOfRecords: []v1.TotalRecords{
// 						v1.TotalRecords{
// 							TotalCnt: 5,
// 						},
// 					},
// 					Products: []v1.ProductData{
// 						v1.ProductData{
// 							Name:              "Windows Instant Client",
// 							Version:           "9.2.0.8.0",
// 							Editor:            "Windows",
// 							Swidtag:           "WIN1",
// 							NumOfEquipments:   3,
// 							NumOfApplications: 1,
// 						},
// 						v1.ProductData{
// 							Name:              "Windows SGBD Noyau",
// 							Version:           "9.2.0.8.0",
// 							Editor:            "Windows",
// 							Swidtag:           "WIN2",
// 							NumOfEquipments:   4,
// 							NumOfApplications: 1,
// 						},
// 						v1.ProductData{
// 							Name:              "Windows Instant Client",
// 							Version:           "9.2.0.8.0",
// 							Editor:            "Windows",
// 							Swidtag:           "WIN3",
// 							NumOfEquipments:   2,
// 							NumOfApplications: 1,
// 						},
// 						v1.ProductData{
// 							Name:              "Windows SGBD Noyau",
// 							Version:           "9.2.0.8.0",
// 							Editor:            "Windows",
// 							Swidtag:           "WIN4",
// 							NumOfEquipments:   1,
// 							NumOfApplications: 1,
// 						},
// 						v1.ProductData{
// 							Name:              "Windows Database",
// 							Version:           "9.2.0.8.0",
// 							Editor:            "Windows",
// 							Swidtag:           "WIN5",
// 							NumOfEquipments:   1,
// 							NumOfApplications: 1,
// 						},
// 					},
// 				}

// 				got, err := dgraph.NewLicenseRepository(dgClient).GetProducts(context.Background(), &v1.QueryProducts{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "swidtag",
// 					SortOrder: "orderasc",
// 				}, []string{"scope1", "scope2"})
// 				if err != nil {
// 					return errors.New("couldnt fetch products from database - " + err.Error())
// 				}
// 				assert.Equalf(t, wantProducts.NumOfRecords[0].TotalCnt, got.NumOfRecords[0].TotalCnt, "%s.TotalCount should be same")
// 				compareProductsAll(t, "ProductData", wantProducts.Products, got.Products)
// 				return nil
// 			},
// 		},
// 		{name: "load static data - applications",
// 			al: &AggregateLoader{
// 				loaders: []loader{
// 					{
// 						load:   loadApplications,
// 						files:  []string{"applications.csv", "applications_products.csv"}, //app files
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 					{
// 						load:   loadProducts,
// 						files:  []string{"prod.csv", "products_equipments.csv"},
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 					{
// 						load:   loadInstances,
// 						files:  []string{"applications_instances.csv"},
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 				},
// 				config: &Config{
// 					BadgerDir:    "badger",
// 					Zero:         ":5080",
// 					Alpha:        []string{":9080"},
// 					CreateSchema: true,
// 					// add files
// 					SchemaFiles:    schemaFiles,
// 					LoadStaticData: true,
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {

// 				wantApplications := &v1.ApplicationInfo{
// 					NumOfRecords: []v1.TotalRecords{
// 						v1.TotalRecords{
// 							TotalCnt: 2,
// 						},
// 					},
// 					Applications: []v1.ApplicationData{
// 						v1.ApplicationData{
// 							Name:             "Acireales",
// 							ApplicationID:    "A01",
// 							ApplicationOwner: "Biogercorp",
// 							NumOfInstances:   1,
// 							NumOfProducts:    2,
// 						},
// 						v1.ApplicationData{
// 							Name:             "Afragusa",
// 							ApplicationID:    "A02",
// 							ApplicationOwner: "Pional",
// 							NumOfInstances:   2,
// 							NumOfProducts:    1,
// 						},
// 					},
// 				}
// 				got, err := dgraph.NewLicenseRepository(dgClient).GetApplications(context.Background(), &v1.QueryApplications{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "applicationId",
// 					SortOrder: "orderasc",
// 				}, []string{"scope2"})
// 				if err != nil {
// 					return errors.New("couldnt fetch applications from database - " + err.Error())
// 				}
// 				assert.Equalf(t, wantApplications.NumOfRecords[0].TotalCnt, got.NumOfRecords[0].TotalCnt, "%s.TotalCount should be same")
// 				compareApplicationsAll(t, "ApplicationData", wantApplications.Applications, got.Applications)
// 				return nil
// 			},
// 		},
// 		{name: "load static data - acquired rights",
// 			al: &AggregateLoader{
// 				loaders: []loader{
// 					{
// 						load:   loadAcquiredRights,
// 						files:  []string{"products_acqRights.csv"},
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 				},
// 				config: &Config{
// 					BadgerDir:      "badger",
// 					Zero:           ":5080",
// 					Alpha:          []string{":9080"},
// 					DropSchema:     true,
// 					CreateSchema:   true,
// 					SchemaFiles:    schemaFiles,
// 					LoadStaticData: true,
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {
// 				acquiredRights := []*v1.AcquiredRights{
// 					&v1.AcquiredRights{
// 						Entity:                         "",
// 						SKU:                            "WIN1PROC",
// 						SwidTag:                        "WIN1",
// 						ProductName:                    "Windows Client",
// 						Editor:                         "Windows",
// 						Metric:                         "Windows.processor.standard",
// 						AcquiredLicensesNumber:         1016,
// 						LicensesUnderMaintenanceNumber: 1008,
// 						AvgLicenesUnitPrice:            2042,
// 						AvgMaintenanceUnitPrice:        14294,
// 						TotalPurchaseCost:              2074672,
// 						TotalMaintenanceCost:           14408352,
// 						TotalCost:                      35155072,
// 					},
// 					&v1.AcquiredRights{
// 						Entity:                         "",
// 						SKU:                            "WIN2PROC",
// 						SwidTag:                        "WIN2",
// 						ProductName:                    "Windows XML Development Kit",
// 						Editor:                         "Windows",
// 						Metric:                         "Windows.processor.standard",
// 						AcquiredLicensesNumber:         181,
// 						LicensesUnderMaintenanceNumber: 181,
// 						AvgLicenesUnitPrice:            1759,
// 						AvgMaintenanceUnitPrice:        12313,
// 						TotalPurchaseCost:              318379,
// 						TotalMaintenanceCost:           2228653,
// 						TotalCost:                      5412443,
// 					},
// 				}
// 				tl, got, err := dgraph.NewLicenseRepository(dgClient).AcquiredRights(context.Background(), &v1.QueryAcquiredRights{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    v1.AcquiredRightsSortBySwidTag,
// 					SortOrder: v1.SortASC,
// 				}, []string{"scope1"})
// 				if err != nil {
// 					return errors.New("couldnt fetch acquired rights from database - " + err.Error())
// 				}
// 				assert.Equalf(t, int32(2), tl, "%s.TotalCount should be same")
// 				compareAcquiredRightsAll(t, "AcquiredRights", acquiredRights, got)
// 				return nil
// 			},
// 		},
// 		{name: "load static data - instances",
// 			al: &AggregateLoader{
// 				loaders: []loader{
// 					{
// 						load:   loadInstances,
// 						files:  []string{"applications_instances.csv", "instances_products.csv", "instances_equipments.csv"},
// 						scopes: []string{"testdata/uploaded/scope1", "testdata/uploaded/scope2"},
// 					},
// 				},
// 				config: &Config{
// 					BadgerDir:      "badger",
// 					Zero:           ":5080",
// 					Alpha:          []string{":9080"},
// 					DropSchema:     true,
// 					CreateSchema:   true,
// 					SchemaFiles:    schemaFiles,
// 					LoadStaticData: true,
// 				},
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			verify: func() error {
// 				q := `{
// 					Instances(func:has(instance.id),orderasc:instance.id){
// 						    ID:              instance.id
// 							Environment:     instance.environment
// 							NumOfEquipments: count(instance.equipment) 
// 							NumOfProducts:   count(instance.product)
								 
// 						  } 
//              	 }`
// 				resp, err := dgClient.NewTxn().Query(context.Background(), q)
// 				if err != nil {
// 					return fmt.Errorf("cannot complete query transaction")
// 				}

// 				wantInstances := []v1.InstancesForApplicationProductData{
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I01",
// 						Environment:     "Development",
// 						NumOfEquipments: 1,
// 						NumOfProducts:   2,
// 					},
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I02",
// 						Environment:     "Development",
// 						NumOfEquipments: 1,
// 						NumOfProducts:   1,
// 					},
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I03",
// 						Environment:     "Production",
// 						NumOfEquipments: 1,
// 						NumOfProducts:   1,
// 					},
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I1",
// 						Environment:     "Development",
// 						NumOfEquipments: 2,
// 						NumOfProducts:   2,
// 					},
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I2",
// 						Environment:     "Development",
// 						NumOfEquipments: 2,
// 						NumOfProducts:   1,
// 					},
// 					v1.InstancesForApplicationProductData{
// 						ID:              "I3",
// 						Environment:     "Production",
// 						NumOfEquipments: 2,
// 						NumOfProducts:   1,
// 					},
// 				}

// 				var decode struct {
// 					Instances []v1.InstancesForApplicationProductData
// 				}
// 				if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
// 					return fmt.Errorf("cannot unmarshal Json object")
// 				}
// 				compareInstancesAll(t, "Instances", wantInstances, decode.Instances)
// 				return nil
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cleanup, err := tt.setup()
// 			if !assert.Empty(t, err, "no error is expected from setup") {
// 				return
// 			}
// 			defer func() {
// 				assert.Empty(t, cleanup(), "error is not expected from cleanup")
// 			}()
// 			if err := tt.al.Load(); (err != nil) != tt.wantErr {
// 				t.Errorf("AggregateLoader.Load() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if !tt.wantErr {
// 				assert.Empty(t, tt.verify(), "error is not expected in verify")
// 			}
// 		})
// 	}
// }

// func querySchema(predicates ...string) ([]*SchemaNode, error) {
// 	if len(predicates) == 0 {
// 		return nil, nil
// 	}
// 	q := `schema (pred: [` + strings.Join(predicates, ",") + `]) {
// 		type
// 		index
// 		reverse
// 		tokenizer
// 		list
// 		count
// 		upsert
// 		lang
// 	  }
// 	`
// 	//	fmt.Println(q)
// 	resp, err := dgClient.NewTxn().Query(context.Background(), q)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return resp.Schema, nil
// }
// func compareSchemaNodeAll(t *testing.T, name string, exp []*SchemaNode, act []*SchemaNode) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of elements are: %d", len(exp)) {
// 		return
// 	}

// 	for i := range exp {
// 		actIdx := indexForPredicte(exp[i].Predicate, act)
// 		if assert.NotEqualf(t, -1, "%s.Predicate is not found in expected nodes", fmt.Sprintf("%s[%d]", name, i)) {

// 		}
// 		compareSchemaNode(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[actIdx])
// 	}
// }

// func indexForPredicte(predicate string, schemas []*SchemaNode) int {
// 	for i := range schemas {
// 		if schemas[i].Predicate == predicate {
// 			return i
// 		}
// 	}
// 	return -1
// }

// func compareSchemaNode(t *testing.T, name string, exp *SchemaNode, act *SchemaNode) {
// 	if exp == nil && act == nil {
// 		return
// 	}
// 	if exp == nil {
// 		assert.Nil(t, act, "attribute is expected to be nil")
// 	}

// 	assert.Equalf(t, exp.Predicate, act.Predicate, "%s.Predicate are not same", name)
// 	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
// 	assert.Equalf(t, exp.Index, act.Index, "%s.Index are not same", name)
// 	assert.ElementsMatchf(t, exp.Tokenizer, act.Tokenizer, "%s.Tokenizer are not same", name)
// 	assert.Equalf(t, exp.Reverse, act.Reverse, "%s.Reverse are not same", name)
// 	assert.Equalf(t, exp.Count, act.Count, "%s.Count are not same", name)
// 	assert.Equalf(t, exp.List, act.List, "%s.List are not same", name)
// 	assert.Equalf(t, exp.Upsert, act.Upsert, "%s.Upsert are not same", name)
// 	assert.Equalf(t, exp.Lang, act.Lang, "%s.Lang are not same", name)
// }

// func compareMetadataAll(t *testing.T, name string, exp []*v1.Metadata, act []*v1.Metadata) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of metdata is: %d", len(exp)) {
// 		return
// 	}

// 	for i := range exp {
// 		compareMetadata(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
// 	}
// }

// func compareMetadata(t *testing.T, name string, exp *v1.Metadata, act *v1.Metadata) {
// 	if exp == nil && act == nil {
// 		return
// 	}
// 	if exp == nil {
// 		assert.Nil(t, act, "metadata is expected to be nil")
// 	}

// 	if exp.ID != "" {
// 		assert.Emptyf(t, act.ID, "%s.ID is expected to be nil", name)
// 	}
// 	assert.Equalf(t, exp.Source, act.Source, "%s.Source should be same", name)
// 	assert.ElementsMatchf(t, exp.Attributes, act.Attributes, "%s.Attributes should be same", name)
// }

// // getAllFilesWithSuffixFullPath("../schema", ".schema")
// func getAllFilesWithSuffixFullPath(dir, suffix string) ([]string, error) {
// 	files, err := ioutil.ReadDir(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var fileNames []string
// 	for _, f := range files {
// 		name := filepath.Base(f.Name())
// 		fmt.Println(name, f.Name())
// 		if !f.IsDir() && strings.HasSuffix(name, suffix) {
// 			fileNames = append(fileNames, dir+"/"+f.Name())
// 		}
// 	}
// 	return fileNames, nil
// }

// func compareProductsAll(t *testing.T, name string, exp []v1.ProductData, act []v1.ProductData) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of metdata is: %d", len(exp)) {
// 		return
// 	}
// 	for i := range exp {
// 		compareProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
// 	}
// }

// func compareProduct(t *testing.T, name string, exp v1.ProductData, act v1.ProductData) {

// 	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
// 	assert.Equalf(t, exp.Version, act.Version, "%s.Version should be same", name)
// 	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor should be same", name)
// 	assert.Equalf(t, exp.Swidtag, act.Swidtag, "%s.Swidtag should be same", name)
// 	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments should be same", name)
// 	assert.Equalf(t, exp.NumOfApplications, act.NumOfApplications, "%s.NumOfApplications should be same", name)

// }

// func compareApplicationsAll(t *testing.T, name string, exp []v1.ApplicationData, act []v1.ApplicationData) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of metdata is: %d", len(exp)) {
// 		return
// 	}
// 	for i := range exp {
// 		compareApplication(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
// 	}
// }

// func compareApplication(t *testing.T, name string, exp v1.ApplicationData, act v1.ApplicationData) {

// 	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
// 	assert.Equalf(t, exp.ApplicationID, act.ApplicationID, "%s.ApplicationID should be same", name)
// 	assert.Equalf(t, exp.ApplicationOwner, act.ApplicationOwner, "%s.ApplicationOwner should be same", name)
// 	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances should be same", name)
// 	assert.Equalf(t, exp.NumOfProducts, act.NumOfProducts, "%s.NumOfProducts should be same", name)

// }
// func compareAcquiredRightsAll(t *testing.T, name string, exp []*v1.AcquiredRights, act []*v1.AcquiredRights) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
// 		return
// 	}

// 	for i := range exp {
// 		compareAcquiredRights(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
// 	}
// }

// func compareAcquiredRights(t *testing.T, name string, exp *v1.AcquiredRights, act *v1.AcquiredRights) {
// 	if exp == nil && act == nil {
// 		return
// 	}
// 	if exp == nil {
// 		assert.Nil(t, act, "attribute is expected to be nil")
// 	}

// 	// if exp.ID != "" {
// 	// 	assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
// 	// }
// 	assert.Equalf(t, exp.Entity, act.Entity, "%s.Entity are not same", name)
// 	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU are not same", name)
// 	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
// 	assert.Equalf(t, exp.ProductName, act.ProductName, "%s.ProductName are not same", name)
// 	assert.Equalf(t, exp.Editor, act.Editor, "%s.Type are not same", name)
// 	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
// 	assert.Equalf(t, exp.AcquiredLicensesNumber, act.AcquiredLicensesNumber, "%s.AcquiredLicensesNumber are not same", name)
// 	assert.Equalf(t, exp.LicensesUnderMaintenanceNumber, act.LicensesUnderMaintenanceNumber, "%s.LicensesUnderMaintenanceNumber are not same", name)
// 	assert.Equalf(t, exp.AvgLicenesUnitPrice, act.AvgLicenesUnitPrice, "%s.AvgLicenesUnitPrice are not same", name)
// 	assert.Equalf(t, exp.AvgMaintenanceUnitPrice, act.AvgMaintenanceUnitPrice, "%s.AvgMaintenanceUnitPrice are not same", name)
// 	assert.Equalf(t, exp.TotalPurchaseCost, act.TotalPurchaseCost, "%s.TotalPurchaseCost are not same", name)
// 	assert.Equalf(t, exp.TotalMaintenanceCost, act.TotalMaintenanceCost, "%s.TotalMaintenanceCost are not same", name)
// 	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost are not same", name)
// }

// func compareEquipments(t *testing.T, equips json.RawMessage, want []byte) {
// 	fields := strings.Split(string(equips), ",")

// 	idIndexes := []int{}
// 	for idx, field := range fields {
// 		if strings.Contains(field, `[{"ID"`) {
// 			if idx < len(fields)-1 {
// 				fields[idx+1] = "[{" + fields[idx+1]
// 			}
// 			idIndexes = append(idIndexes, idx)
// 			continue
// 		}
// 		if strings.Contains(field, `{"ID"`) {
// 			if idx < len(fields)-1 {
// 				fields[idx+1] = "{" + fields[idx+1]
// 			}
// 			idIndexes = append(idIndexes, idx)
// 		}
// 	}

// 	// remove indexes from fields
// 	idLessfields := make([]string, 0, len(fields)-len(idIndexes))
// 	count := 0
// 	for idx := range fields {
// 		if count < len(idIndexes) && idx == idIndexes[count] {
// 			count++
// 			continue
// 		}
// 		idLessfields = append(idLessfields, fields[idx])
// 	}

// 	assert.Equal(t, strings.Join(strings.Split(string(want), ","), ","), strings.Join(idLessfields, ","))

// }

// func equipmentsJSONFromCSV(filename string, eqType *v1.EquipmentType, ignoreDisplayed bool) ([]string, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	r := csv.NewReader(file)
// 	r.Comma = ';'
// 	records, err := r.ReadAll()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(records) == 0 {
// 		return nil, errors.New("no data in: " + filename)
// 	}

// 	headers := records[0]

// 	pkAttr, err := eqType.PrimaryKeyAttribute()
// 	if err != nil {
// 		return nil, err
// 	}

// 	records = records[1:]
// 	data := []string{}

// 	for _, rec := range records {
// 		recJSON := ""
// 		for idx, val := range rec {
// 			if headers[idx] == pkAttr.MappedTo {
// 				recJSON = fmt.Sprintf(`"%s":"%s",`, pkAttr.Name, val) + recJSON
// 				continue
// 			}
// 			i := attributeByMapping(headers[idx], eqType.Attributes)
// 			if i == -1 {
// 				// Continue log this
// 				continue
// 			}

// 			attr := eqType.Attributes[i]

// 			if attr.IsParentIdentifier {
// 				continue
// 			}

// 			if ignoreDisplayed {
// 				if !attr.IsDisplayed {
// 					continue
// 				}
// 			}

// 			switch attr.Type {
// 			case v1.DataTypeString:
// 				recJSON += fmt.Sprintf(`"%s":"%s",`, attr.Name, val)
// 			case v1.DataTypeInt:
// 				recJSON += fmt.Sprintf(`"%s":%s,`, attr.Name, val)
// 			case v1.DataTypeFloat:
// 				recJSON += fmt.Sprintf(`"%s":%s.000000,`, attr.Name, val)
// 			default:
// 				// TODO: unsupported data type log this
// 			}
// 		}
// 		recJSON = `{` + strings.TrimSuffix(recJSON, ",") + `}`
// 		data = append(data, recJSON)
// 	}

// 	return data, nil
// }
// func attributeByMapping(mappedTo string, attributes []*v1.Attribute) int {
// 	for idx := range attributes {
// 		if attributes[idx].MappedTo == mappedTo {
// 			return idx
// 		}
// 	}
// 	return -1
// }
// func deleteNode(id string) error {
// 	mu := &api.Mutation{
// 		CommitNow:  true,
// 		DeleteJson: []byte(`{"uid": "` + id + `"}`),
// 	}
// 	_, err := dgClient.NewTxn().Mutate(context.Background(), mu)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
// func compareInstancesAll(t *testing.T, name string, exp []v1.InstancesForApplicationProductData, act []v1.InstancesForApplicationProductData) {
// 	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
// 		return
// 	}

// 	for i := range exp {
// 		compareInstance(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
// 	}
// }

// func compareInstance(t *testing.T, name string, exp v1.InstancesForApplicationProductData, act v1.InstancesForApplicationProductData) {
// 	// if exp == nil && act == nil {
// 	// 	return
// 	// }
// 	// if exp == nil {
// 	// 	assert.Nil(t, act, "attribute is expected to be nil")
// 	// }

// 	assert.Equalf(t, exp.ID, act.ID, "%s.Id are not same", name)
// 	assert.Equalf(t, exp.Environment, act.Environment, "%s.Environment are not same", name)
// 	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
// 	assert.Equalf(t, exp.NumOfProducts, act.NumOfProducts, "%s.NumOfProducts are not same", name)
// }
// func deleteNodes(ids ...string) error {
// 	for _, id := range ids {
// 		if err := deleteNode(id); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
