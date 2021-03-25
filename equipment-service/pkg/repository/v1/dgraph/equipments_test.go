// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"strings"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

// var deleteAll = &api.Value{
// 	Val: &api.Value_DefaultVal{
// 		DefaultVal: "_STAR_ALL",
// 	},
// }

func TestEquipmentRepository_CreateEquipmentType(t *testing.T) {
	type args struct {
		ctx    context.Context
		eqType *v1.EquipmentType
		scopes []string
	}
	tests := []struct {
		name            string
		lr              *EquipmentRepository
		args            args
		setup           func() (*v1.EquipmentType, func() error, error)
		veryfy          func(repo *EquipmentRepository) (*v1.EquipmentType, error)
		wantSchemaNodes []*SchemaNode
		predicates      []string
		wantErr         bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (*v1.EquipmentType, func() error, error) {
				// TODO create two nodes for parent type and data source
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, nil, errors.New("cannot find source id after mutation in setup")
				}
				eqType := &v1.EquipmentType{
					Type:     "MyType",
					SourceID: sourceID,
					ParentID: parentID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsIdentifier: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_2",
						},
						&v1.Attribute{
							Name:     "attr2.1",
							Type:     v1.DataTypeInt,
							MappedTo: "mapping_2.1",
						},
						&v1.Attribute{
							Name:         "attr3",
							Type:         v1.DataTypeFloat,
							IsSearchable: true,
							MappedTo:     "mapping_3",
						},
						&v1.Attribute{
							Name:     "attr3.1",
							Type:     v1.DataTypeFloat,
							MappedTo: "mapping_3.1",
						},
						&v1.Attribute{
							Name:               "attr4",
							Type:               v1.DataTypeString,
							IsParentIdentifier: true,
							IsDisplayed:        true,
							MappedTo:           "mapping_4",
						},
						&v1.Attribute{
							Name:         "attr4.1",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4.1",
						},
						&v1.Attribute{
							Name:        "attr4.2",
							Type:        v1.DataTypeString,
							IsDisplayed: true,
							MappedTo:    "mapping_4.2",
						},
					},
				}
				return eqType, func() error {
					if err := deleteNode(parentID); err != nil {
						return err
					}
					if err := deleteNode(sourceID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			veryfy: func(repo *EquipmentRepository) (*v1.EquipmentType, error) {
				eqType, err := repo.equipmentTypeByType(context.Background(), "MyType", []string{"scope1"})
				if err != nil {
					return nil, err
				}
				return eqType, nil
			},
			wantSchemaNodes: []*SchemaNode{
				&SchemaNode{
					Predicate: "equipment.MyType.attr2",
					Type:      "int",
					Index:     true,
					Tokenizer: []string{"int"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr2.1",
					Type:      "int",
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr3",
					Type:      "float",
					Index:     true,
					Tokenizer: []string{"float"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr3.1",
					Type:      "float",
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr4.1",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr4.2",
					Type:      "string",
				},
			},
			predicates: []string{
				"equipment.MyType.attr2",
				"equipment.MyType.attr2.1",
				"equipment.MyType.attr3",
				"equipment.MyType.attr3.1",
				"equipment.MyType.attr4.1",
				"equipment.MyType.attr4.2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eqType, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expect in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expect in cleanup")
			}()
			got, err := tt.lr.CreateEquipmentType(tt.args.ctx, eqType, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.CreateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			defer func() {
				err := deleteNode(got.ID)
				assert.Empty(t, err, "error is not expect in deleteNode")
			}()

			want, err := tt.veryfy(tt.lr)
			if !assert.Empty(t, err, "error is not expect in verify") {
				return
			}

			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", want, got)
				sns, err := querySchema(tt.predicates...)
				if !assert.Emptyf(t, err, "error is not expect while quering schema for predicates: %v", tt.predicates) {
					return
				}
				compareSchemaNodeAll(t, "schemaNodes", tt.wantSchemaNodes, sns)
			}
		})
	}
}

func TestEquipmentRepository_EquipmentTypes(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		args    args
		setup   func(repo *EquipmentRepository) ([]*v1.EquipmentType, func() error, error)
		wantErr bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: []string{"scope1"},
			},
			setup: func(repo *EquipmentRepository) ([]*v1.EquipmentType, func() error, error) {
				// TODO create two nodes for parent type and data source
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, nil, errors.New("cannot find source id after mutation in setup")
				}

				eqTypes := []*v1.EquipmentType{
					&v1.EquipmentType{
						Type:     "MyType1",
						SourceID: sourceID,
						ParentID: parentID,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							&v1.Attribute{
								Name:         "attr1",
								Type:         v1.DataTypeString,
								IsSearchable: true,
								IsIdentifier: true,
								IsDisplayed:  true,
								MappedTo:     "mapping_1",
							},
							&v1.Attribute{
								Name:               "attr2",
								Type:               v1.DataTypeString,
								IsSearchable:       false,
								IsParentIdentifier: true,
								IsDisplayed:        false,
								MappedTo:           "mapping_2",
							},
						},
					},
					&v1.EquipmentType{
						Type:     "MyType2",
						SourceID: sourceID,
						ParentID: parentID,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							&v1.Attribute{
								Name:         "attr1",
								Type:         v1.DataTypeString,
								IsSearchable: true,
								IsIdentifier: true,
								IsDisplayed:  true,
								MappedTo:     "mapping_1",
							},
						},
					},
				}

				for _, eqType := range eqTypes {
					_, err := repo.CreateEquipmentType(context.Background(), eqType, eqType.Scopes)
					if err != nil {
						fmt.Print(err)
						return nil, nil, err
					}
				}

				return eqTypes, func() error {
					return deleteNodes(parentID, sourceID, eqTypes[0].ID, eqTypes[1].ID)
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, cleanup, err := tt.setup(tt.lr)
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			got, err := tt.lr.EquipmentTypes(tt.args.ctx, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.EquipmentTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareEquipmentTypeAll(t, "EquipmentTypes", want, got)
			}
		})
	}
}

func TestEquipmentRepository_UpdateEquipmentType(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		typ    string
		req    *v1.UpdateEquipmentRequest
		scopes []string
	}
	tests := []struct {
		name   string
		lr     *EquipmentRepository
		args   args
		setup  func() (*v1.EquipmentType, string, func() error, error)
		veryfy func(repo *EquipmentRepository) (*v1.EquipmentType, error)
		//wantRetType     []*v1.Attribute
		wantSchemaNodes []*SchemaNode
		predicates      []string
		wantErr         bool
	}{
		{name: "SUCCESS - no change in parent",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateEquipmentRequest{
					Attr: []*v1.Attribute{
						&v1.Attribute{
							Name:               "attr4",
							Type:               1,
							IsIdentifier:       false,
							IsDisplayed:        true,
							IsSearchable:       true,
							IsParentIdentifier: false,
							MappedTo:           "mapping_4",
						},
						&v1.Attribute{
							Name:               "attr5",
							Type:               2,
							IsIdentifier:       false,
							IsDisplayed:        true,
							IsSearchable:       false,
							IsParentIdentifier: false,
							MappedTo:           "mapping_5",
						},
						&v1.Attribute{
							Name:         "attr6",
							Type:         v1.DataTypeFloat,
							IsSearchable: true,
							MappedTo:     "mapping_6",
						},
					},
				},
				scopes: []string{"scope1"},
			},
			setup: func() (*v1.EquipmentType, string, func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, "", nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return nil, "", nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, "", nil, errors.New("cannot find source id after mutation in setup")
				}

				eqType := &v1.EquipmentType{
					Type:     "MyType",
					SourceID: sourceID,
					ParentID: parentID,
					Scopes:   []string{"scope1"},
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:               "attr2",
							Type:               v1.DataTypeString,
							IsParentIdentifier: true,
							IsDisplayed:        true,
							MappedTo:           "mapping_2",
						},
						&v1.Attribute{
							Name:         "attr3",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_3",
						},
					},
				}
				repo := NewEquipmentRepository(dgClient)
				retEqp, err := repo.CreateEquipmentType(context.Background(), eqType, eqType.Scopes)
				if err != nil {
					return nil, "", nil, errors.New("cannot create equipment in setup")
				}
				return retEqp, "", func() error {
					return deleteNodes(parentID, sourceID, retEqp.ID)
				}, nil
			},
			veryfy: func(repo *EquipmentRepository) (*v1.EquipmentType, error) {
				eqType, err := repo.equipmentTypeByType(context.Background(), "MyType", []string{"scope1"})
				if err != nil {
					return nil, err
				}
				return eqType, nil
			},
			wantSchemaNodes: []*SchemaNode{
				&SchemaNode{
					Predicate: "equipment.MyType.attr1",
					Type:      "int",
					Index:     true,
					Tokenizer: []string{"int"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr3",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr4",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr5",
					Type:      "int",
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr6",
					Type:      "float",
					Index:     true,
					Tokenizer: []string{"float"},
				},
			},
			predicates: []string{
				"equipment.MyType.attr1",
				"equipment.MyType.attr3",
				"equipment.MyType.attr4",
				"equipment.MyType.attr5",
				"equipment.MyType.attr6",
			},
			wantErr: false,
		},
		{name: "SUCCESS - parent created ",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateEquipmentRequest{
					Attr: []*v1.Attribute{
						&v1.Attribute{
							Name:               "attr3",
							Type:               v1.DataTypeString,
							IsParentIdentifier: true,
							IsDisplayed:        true,
							MappedTo:           "mapping_3",
						},
						&v1.Attribute{
							Name:               "attr4",
							Type:               v1.DataTypeInt,
							IsIdentifier:       false,
							IsDisplayed:        true,
							IsSearchable:       false,
							IsParentIdentifier: false,
							MappedTo:           "mapping_4",
						},
						&v1.Attribute{
							Name:         "attr5",
							Type:         v1.DataTypeFloat,
							IsSearchable: true,
							MappedTo:     "mapping_5",
						},
					},
				},
			},
			setup: func() (*v1.EquipmentType, string, func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						// &api.NQuad{
						// 	Subject:     blankID("parent"),
						// 	Predicate:   "metadata_parent",
						// 	ObjectValue: stringObjectValue("eq_type_1"),
						// },
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, "", nil, err
				}

				// parentID, ok := assigned.Uids["parent"]
				// if !ok {
				// 	return nil, "", nil, errors.New("cannot find parent id after mutation in setup")
				// }

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, "", nil, errors.New("cannot find source id after mutation in setup")
				}
				repo := NewEquipmentRepository(dgClient)
				eqType1 := &v1.EquipmentType{
					Type:     "MyType2",
					SourceID: sourceID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_2",
						},
					},
					Scopes: []string{"scope1"},
				}
				equip1, err := repo.CreateEquipmentType(context.Background(), eqType1, eqType1.Scopes)
				if err != nil {
					return nil, "", nil, errors.New("cannot create equipment in setup")
				}
				eqType2 := &v1.EquipmentType{
					Type:     "MyType",
					SourceID: sourceID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_2",
						},
					},
					Scopes: []string{"scope1"},
				}
				retEqp, err := repo.CreateEquipmentType(context.Background(), eqType2, eqType2.Scopes)
				if err != nil {
					return nil, "", nil, errors.New("cannot create equipment in setup")
				}
				return retEqp, equip1.ID, func() error {
					return deleteNodes(sourceID, equip1.ID, retEqp.ID)
				}, nil
			},
			veryfy: func(repo *EquipmentRepository) (*v1.EquipmentType, error) {
				eqType, err := repo.equipmentTypeByType(context.Background(), "MyType", []string{"scope1"})
				if err != nil {
					return nil, err
				}
				return eqType, nil
			},
			wantSchemaNodes: []*SchemaNode{
				&SchemaNode{
					Predicate: "equipment.MyType.attr1",
					Type:      "int",
					Index:     true,
					Tokenizer: []string{"int"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr2",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr4",
					Type:      "int",
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr5",
					Type:      "float",
					Index:     true,
					Tokenizer: []string{"float"},
				},
			},
			predicates: []string{
				"equipment.MyType.attr1",
				"equipment.MyType.attr2",
				"equipment.MyType.attr4",
				"equipment.MyType.attr5",
			},
			wantErr: false,
		},
		{name: "SUCCESS - parent updated ",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateEquipmentRequest{
					Attr: []*v1.Attribute{
						&v1.Attribute{
							Name:               "attr3",
							Type:               v1.DataTypeString,
							IsParentIdentifier: true,
							IsDisplayed:        true,
							MappedTo:           "mapping_3",
						},
						&v1.Attribute{
							Name:               "attr4",
							Type:               v1.DataTypeInt,
							IsIdentifier:       false,
							IsDisplayed:        true,
							IsSearchable:       false,
							IsParentIdentifier: false,
							MappedTo:           "mapping_4",
						},
						&v1.Attribute{
							Name:         "attr5",
							Type:         v1.DataTypeFloat,
							IsSearchable: true,
							MappedTo:     "mapping_5",
						},
					},
				},
			},
			setup: func() (*v1.EquipmentType, string, func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, "", nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return nil, "", nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, "", nil, errors.New("cannot find source id after mutation in setup")
				}
				repo := NewEquipmentRepository(dgClient)
				eqType1 := &v1.EquipmentType{
					Type:     "MyType2",
					SourceID: sourceID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_2",
						},
					},
					Scopes: []string{"scope1"},
				}
				equip1, err := repo.CreateEquipmentType(context.Background(), eqType1, eqType1.Scopes)
				if err != nil {
					return nil, "", nil, errors.New("cannot create equipment in setup")
				}
				eqType2 := &v1.EquipmentType{
					Type:     "MyType",
					SourceID: sourceID,
					ParentID: parentID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_2",
						},
					},
					Scopes: []string{"scope1"},
				}
				retEqp, err := repo.CreateEquipmentType(context.Background(), eqType2, eqType2.Scopes)
				if err != nil {
					return nil, "", nil, errors.New("cannot create equipment in setup")
				}
				return retEqp, equip1.ID, func() error {
					return deleteNodes(parentID, sourceID, equip1.ID, retEqp.ID)
				}, nil
			},
			veryfy: func(repo *EquipmentRepository) (*v1.EquipmentType, error) {
				eqType, err := repo.equipmentTypeByType(context.Background(), "MyType", []string{"scope1"})
				if err != nil {
					return nil, err
				}
				return eqType, nil
			},
			wantSchemaNodes: []*SchemaNode{
				&SchemaNode{
					Predicate: "equipment.MyType.attr1",
					Type:      "int",
					Index:     true,
					Tokenizer: []string{"int"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr2",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram"},
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr4",
					Type:      "int",
				},
				&SchemaNode{
					Predicate: "equipment.MyType.attr5",
					Type:      "float",
					Index:     true,
					Tokenizer: []string{"float"},
				},
			},
			predicates: []string{
				"equipment.MyType.attr1",
				"equipment.MyType.attr2",
				"equipment.MyType.attr4",
				"equipment.MyType.attr5",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, parID, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expect in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expect in cleanup")
			}()
			tt.args.req.ParentID = parID
			gotRetType, err := tt.lr.UpdateEquipmentType(tt.args.ctx, got.ID, got.Type, tt.args.req, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.UpdateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			defer func() {
				err := deleteNode(got.ID)
				assert.Empty(t, err, "error is not expect in deleteNode")
			}()

			want, err := tt.veryfy(tt.lr)
			if !assert.Empty(t, err, "error is not expect in verify") {
				return
			}

			if !tt.wantErr {
				got.Attributes = append(got.Attributes, gotRetType...)
				if parID != "" {
					got.ParentID = parID
				}
				compareEquipmentType(t, "EquipmentType", want, got)
				sns, err := querySchema(tt.predicates...)
				if !assert.Emptyf(t, err, "error is not expect while quering schema for predicates: %v", tt.predicates) {
					return
				}
				compareSchemaNodeAll(t, "schemaNodes", tt.wantSchemaNodes, sns)
			}
		})
	}
}

func TestEquipmentRepository_EquipmentWithID(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		scopes []string
	}
	tests := []struct {
		name            string
		lr              *EquipmentRepository
		args            args
		setup           func() (*v1.EquipmentType, func() error, error)
		wantSchemaNodes []*SchemaNode
		wantErr         bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (*v1.EquipmentType, func() error, error) {
				// TODO create two nodes for parent type and data source
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return nil, nil, errors.New("cannot find source id after mutation in setup")
				}

				eqType := &v1.EquipmentType{
					Type:     "MyType",
					SourceID: sourceID,
					ParentID: parentID,
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:         "attr1",
							Type:         v1.DataTypeString,
							IsSearchable: true,
							IsIdentifier: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_1",
						},
						&v1.Attribute{
							Name:         "attr2",
							Type:         v1.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_2",
						},
						&v1.Attribute{
							Name:               "attr3",
							Type:               v1.DataTypeString,
							IsParentIdentifier: true,
							IsDisplayed:        true,
							MappedTo:           "mapping_3",
						},
						&v1.Attribute{
							Name:        "attr4",
							Type:        v1.DataTypeString,
							IsDisplayed: true,
							MappedTo:    "mapping_4",
						},
					},
				}
				repo := NewEquipmentRepository(dgClient)
				retEqp, err := repo.CreateEquipmentType(context.Background(), eqType, []string{})
				if err != nil {
					return nil, nil, errors.New("cannot create equipment in setup")
				}
				return retEqp, func() error {
					if err := deleteNode(parentID); err != nil {
						return err
					}
					if err := deleteNode(sourceID); err != nil {
						return err
					}
					return nil
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expect in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expect in cleanup")
			}()

			defer func() {
				err := deleteNode(got.ID)
				assert.Empty(t, err, "error is not expect in deleteNode")
			}()

			want, err := tt.lr.EquipmentWithID(tt.args.ctx, got.ID, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.EquipmentWithID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", want, got)
			}

		})
	}
}

func compareEquipmentTypeAll(t *testing.T, name string, exp []*v1.EquipmentType, act []*v1.EquipmentType) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareEquipmentType(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareEquipmentType(t *testing.T, name string, exp *v1.EquipmentType, act *v1.EquipmentType) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "equipment Type is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	if exp.ParentID != "" {
		assert.Equalf(t, exp.ParentID, act.ParentID, "%s.ParentID are not same", name)
	}

	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
	assert.Equalf(t, exp.SourceID, act.SourceID, "%s.SourceID are not same", name)
}

func compareAttributeAll(t *testing.T, name string, exp []*v1.Attribute, act []*v1.Attribute) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAttribute(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareAttribute(t *testing.T, name string, exp *v1.Attribute, act *v1.Attribute) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.IsIdentifier, act.IsIdentifier, "%s.IsIdentifier are not same", name)
	assert.Equalf(t, exp.IsDisplayed, act.IsDisplayed, "%s.IsDisplayed are not same", name)
	assert.Equalf(t, exp.IsSearchable, act.IsSearchable, "%s.Type are not same", name)
	assert.Equalf(t, exp.IsParentIdentifier, act.IsParentIdentifier, "%s.IsParentIdentifier are not same", name)
	assert.Equalf(t, exp.MappedTo, act.MappedTo, "%s.Type are not same", name)
}

func compareSchemaNodeAll(t *testing.T, name string, exp []*SchemaNode, act []*SchemaNode) {
	if !assert.Lenf(t, act, len(exp), "expected number of elements are: %d", len(exp)) {
		return
	}

	for i := range exp {
		actIdx := indexForPredicte(exp[i].Predicate, act)
		if assert.NotEqualf(t, -1, "%s.Predicate is not found in expected nodes", fmt.Sprintf("%s[%d]", name, i)) {

		}
		compareSchemaNode(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[actIdx])
	}
}

func indexForPredicte(predicate string, schemas []*SchemaNode) int {
	for i := range schemas {
		if schemas[i].Predicate == predicate {
			return i
		}
	}
	return -1
}

func compareSchemaNode(t *testing.T, name string, exp *SchemaNode, act *SchemaNode) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Predicate, act.Predicate, "%s.Predicate are not same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
	assert.Equalf(t, exp.Index, act.Index, "%s.Index are not same", name)
	assert.ElementsMatchf(t, exp.Tokenizer, act.Tokenizer, "%s.Tokenizer are not same", name)
	assert.Equalf(t, exp.Reverse, act.Reverse, "%s.Reverse are not same", name)
	assert.Equalf(t, exp.Count, act.Count, "%s.Count are not same", name)
	assert.Equalf(t, exp.List, act.List, "%s.List are not same", name)
	assert.Equalf(t, exp.Upsert, act.Upsert, "%s.Upsert are not same", name)
	assert.Equalf(t, exp.Lang, act.Lang, "%s.Lang are not same", name)
}

type SchemaNode struct {
	Predicate string   `json:"predicate,omitempty"`
	Type      string   `json:"type,omitempty"`
	Index     bool     `json:"index,omitempty"`
	Tokenizer []string `json:"tokenizer,omitempty"`
	Reverse   bool     `json:"reverse,omitempty"`
	Count     bool     `json:"count,omitempty"`
	List      bool     `json:"list,omitempty"`
	Upsert    bool     `json:"upsert,omitempty"`
	Lang      bool     `json:"lang,omitempty"`
}

func querySchema(predicates ...string) ([]*SchemaNode, error) {
	if len(predicates) == 0 {
		return nil, nil
	}
	q := `
		schema (pred: [` + strings.Join(predicates, ",") + `]) {
		type
		index
		reverse
		tokenizer
		list
		count
		upsert
		lang
	  }
	`
	//	fmt.Println(q)
	resp, err := dgClient.NewTxn().Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	type data struct {
		Schema []*SchemaNode
	}
	d := &data{}
	if err := json.Unmarshal(resp.Json, d); err != nil {
		return nil, err
	}

	return d.Schema, nil
}

func deleteNodes(ids ...string) error {

	for _, id := range ids {
		if err := deleteNode(id); err != nil {
			return err
		}
	}

	return nil
}

func deleteNode(id string) error {
	mu := &api.Mutation{
		CommitNow:  true,
		DeleteJson: []byte(`{"uid": "` + id + `"}`),
		// Del: []*api.NQuad{
		// 	&api.NQuad{
		// 		Subject:     id,
		// 		Predicate:   "*",
		// 		ObjectValue: deleteAll,
		// 	},
	}

	// delete all the data
	_, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return err
	}

	return nil
}

func attributeIndex(expAttr *v1.Attribute, actAttr []*v1.Attribute) int {
	for i := range actAttr {
		if expAttr.Name == actAttr[i].Name {
			return i
		}
	}
	return -1
}

// func TestEquipmentRepository_EquipmentWithID(t *testing.T) {
// 	type args struct {
// 		ctx    context.Context
// 		id     string
// 		scopes []string
// 	}
// 	tests := []struct {
// 		name            string
// 		lr              *EquipmentRepository
// 		args            args
// 		setup           func() (*v1.EquipmentType, func() error, error)
// 		wantSchemaNodes []*api.SchemaNode
// 		wantErr         bool
// 	}{
// 		{name: "success",
// 			lr: NewEquipmentRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 			},
// 			setup: func() (*v1.EquipmentType, func() error, error) {
// 				// TODO create two nodes for parent type and data source
// 				mu := &api.Mutation{
// 					CommitNow: true,
// 					Set: []*api.NQuad{
// 						&api.NQuad{
// 							Subject:     blankID("parent"),
// 							Predicate:   "metadata_parent",
// 							ObjectValue: stringObjectValue("eq_type_1"),
// 						},
// 						&api.NQuad{
// 							Subject:     blankID("data_source"),
// 							Predicate:   "metadata_source",
// 							ObjectValue: stringObjectValue("eq_type_1"),
// 						},
// 					},
// 				}

// 				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
// 				if err != nil {
// 					return nil, nil, err
// 				}

// 				parentID, ok := assigned.Uids["parent"]
// 				if !ok {
// 					return nil, nil, errors.New("cannot find parent id after mutation in setup")
// 				}

// 				sourceID, ok := assigned.Uids["data_source"]
// 				if !ok {
// 					return nil, nil, errors.New("cannot find source id after mutation in setup")
// 				}

// 				eqType := &v1.EquipmentType{
// 					Type:     "MyType",
// 					SourceID: sourceID,
// 					ParentID: parentID,
// 					Attributes: []*v1.Attribute{
// 						&v1.Attribute{
// 							Name:         "attr1",
// 							Type:         v1.DataTypeString,
// 							IsSearchable: true,
// 							IsIdentifier: true,
// 							IsDisplayed:  true,
// 							MappedTo:     "mapping_1",
// 						},
// 						&v1.Attribute{
// 							Name:         "attr2",
// 							Type:         v1.DataTypeInt,
// 							IsSearchable: true,
// 							MappedTo:     "mapping_2",
// 						},
// 						&v1.Attribute{
// 							Name:               "attr3",
// 							Type:               v1.DataTypeString,
// 							IsParentIdentifier: true,
// 							IsDisplayed:        true,
// 							MappedTo:           "mapping_3",
// 						},
// 						&v1.Attribute{
// 							Name:        "attr4",
// 							Type:        v1.DataTypeString,
// 							IsDisplayed: true,
// 							MappedTo:    "mapping_4",
// 						},
// 					},
// 				}
// 				repo := NewEquipmentRepository(dgClient)
// 				retEqp, err := repo.CreateEquipmentType(context.Background(), eqType, []string{})
// 				if err != nil {
// 					return nil, nil, errors.New("cannot create equipment in setup")
// 				}
// 				return retEqp, func() error {
// 					if err := deleteNode(parentID); err != nil {
// 						return err
// 					}
// 					if err := deleteNode(sourceID); err != nil {
// 						return err
// 					}
// 					return nil
// 				}, nil
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			got, cleanup, err := tt.setup()
// 			if !assert.Empty(t, err, "error is not expect in setup") {
// 				return
// 			}
// 			defer func() {
// 				err := cleanup()
// 				assert.Empty(t, err, "error is not expect in cleanup")
// 			}()

// 			defer func() {
// 				err := deleteNode(got.ID)
// 				assert.Empty(t, err, "error is not expect in deleteNode")
// 			}()

// 			want, err := tt.lr.EquipmentWithID(tt.args.ctx, got.ID, tt.args.scopes)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("EquipmentRepository.EquipmentWithID() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if !tt.wantErr {
// 				compareEquipmentType(t, "EquipmentType", want, got)
// 			}

// 		})
// 	}
// }

func TestEquipmentRepository_EquipmentTypeChildren(t *testing.T) {
	type args struct {
		ctx      context.Context
		eqTypeID string
		depth    int
		scopes   []string
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		args    args
		setup   func(repo *EquipmentRepository) (string, []*v1.EquipmentType, func() error, error)
		wantErr bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				depth:  2,
				scopes: []string{"scope1"},
			},
			setup: func(repo *EquipmentRepository) (string, []*v1.EquipmentType, func() error, error) {
				// TODO create two nodes for parent type and data source
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
						&api.NQuad{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("eq_type_1"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return "", nil, nil, err
				}

				parentID, ok := assigned.Uids["parent"]
				if !ok {
					return "", nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID, ok := assigned.Uids["data_source"]
				if !ok {
					return "", nil, nil, errors.New("cannot find source id after mutation in setup")
				}

				eqTypes := []*v1.EquipmentType{
					&v1.EquipmentType{
						Type:     "MyType1",
						SourceID: sourceID,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							&v1.Attribute{
								Name:         "attr1",
								Type:         v1.DataTypeString,
								IsSearchable: true,
								IsIdentifier: true,
								IsDisplayed:  true,
								MappedTo:     "mapping_1",
							},
							&v1.Attribute{
								Name:         "attr2",
								Type:         v1.DataTypeString,
								IsSearchable: false,
								// IsParentIdentifier: true,
								IsDisplayed: false,
								MappedTo:    "mapping_2",
							},
						},
					},
					&v1.EquipmentType{
						Type:     "MyType2",
						SourceID: sourceID,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							&v1.Attribute{
								Name:               "attr1",
								Type:               v1.DataTypeString,
								IsSearchable:       true,
								IsIdentifier:       true,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_1",
							},
						},
					},
				}
				eqType1, err := repo.CreateEquipmentType(context.Background(), eqTypes[0], eqTypes[0].Scopes)
				if err != nil {
					return "", nil, nil, err
				}
				eqTypes[1].ParentID = eqType1.ID
				eqType2, err := repo.CreateEquipmentType(context.Background(), eqTypes[1], eqTypes[1].Scopes)
				if err != nil {
					return "", nil, nil, err
				}

				return eqType1.ID, []*v1.EquipmentType{eqTypes[1]}, func() error {
					return deleteNodes(parentID, sourceID, eqType1.ID, eqType2.ID)
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eqTypeID, want, cleanup, err := tt.setup(tt.lr)
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			got, err := tt.lr.EquipmentTypeChildren(tt.args.ctx, eqTypeID, tt.args.depth, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.EquipmentTypeChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentTypeAll(t, "EquipmentRepository.EquipmentTypeChildren()", want, got)
			}
		})
	}
}

func TestEquipmentRepository_DeleteEquipmentType(t *testing.T) {
	type args struct {
		ctx    context.Context
		eqType string
		scope  string
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		setup   func(repo *EquipmentRepository) ([]*v1.EquipmentType, func() error, error)
		verify  func(repo *EquipmentRepository) ([]*v1.EquipmentType, error)
		args    args
		wantErr bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: "MyType1",
				scope:  "scope1",
			},
			setup: func(repo *EquipmentRepository) ([]*v1.EquipmentType, func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("parent1"),
						},
						{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("data_source1"),
						},
						{
							Subject:     blankID("parent"),
							Predicate:   "metadata_parent",
							ObjectValue: stringObjectValue("parent2"),
						},
						{
							Subject:     blankID("data_source"),
							Predicate:   "metadata_source",
							ObjectValue: stringObjectValue("data_source2"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				parentID1, ok := assigned.Uids["parent1"]
				if !ok {
					return nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID1, ok := assigned.Uids["data_source1"]
				if !ok {
					return nil, nil, errors.New("cannot find source id after mutation in setup")
				}
				parentID2, ok := assigned.Uids["parent2"]
				if !ok {
					return nil, nil, errors.New("cannot find parent id after mutation in setup")
				}

				sourceID2, ok := assigned.Uids["data_source2"]
				if !ok {
					return nil, nil, errors.New("cannot find source id after mutation in setup")
				}

				eqTypes := []*v1.EquipmentType{
					{
						Type:     "MyType1",
						SourceID: sourceID1,
						ParentID: parentID1,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							{
								Name:         "attr1",
								Type:         v1.DataTypeString,
								IsSearchable: true,
								IsIdentifier: true,
								IsDisplayed:  true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               v1.DataTypeString,
								IsSearchable:       false,
								IsParentIdentifier: true,
								IsDisplayed:        false,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						Type:     "MyType2",
						SourceID: sourceID2,
						ParentID: parentID2,
						Scopes:   []string{"scope1"},
						Attributes: []*v1.Attribute{
							{
								Name:         "attr1",
								Type:         v1.DataTypeString,
								IsSearchable: true,
								IsIdentifier: true,
								IsDisplayed:  true,
								MappedTo:     "mapping_1",
							},
						},
					},
				}

				for _, eqType := range eqTypes {
					_, err := repo.CreateEquipmentType(context.Background(), eqType, eqType.Scopes)
					if err != nil {
						fmt.Print(err)
						return nil, nil, err
					}
				}

				return eqTypes[1:], func() error {
					return deleteNodes(parentID1, sourceID1, parentID2, sourceID2, eqTypes[0].ID, eqTypes[1].ID)
				}, nil
			},
			verify: func(repo *EquipmentRepository) ([]*v1.EquipmentType, error) {
				eqTypes, err := repo.EquipmentTypes(context.Background(), []string{"scope1"})
				if err != nil {
					return nil, err
				}
				return eqTypes, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantEqTypes, cleanup, err := tt.setup(tt.lr)
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			if err := tt.lr.DeleteEquipmentType(tt.args.ctx, tt.args.eqType, tt.args.scope); (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.DeleteEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				actEqTypes, err := tt.verify(tt.lr)
				assert.Empty(t, err, "error is not expected in verify")
				compareEquipmentTypeAll(t, "DeleteEquipmentType", wantEqTypes, actEqTypes)
			}
		})
	}
}
