// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
)

func equipmentSetup(t *testing.T) (eq []*v1.EquipmentType, cleanup func() error, retErr error) {
	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			&api.NQuad{
				Subject:     blankID("parent"),
				Predicate:   "parent",
				ObjectValue: stringObjectValue("parent_equip"),
			},
			&api.NQuad{
				Subject:     blankID("data_source"),
				Predicate:   "metadata.source",
				ObjectValue: stringObjectValue("equip_3.csv"),
			},
			&api.NQuad{
				Subject:     blankID("data_source1"),
				Predicate:   "metadata.source",
				ObjectValue: stringObjectValue("equip_4.csv"),
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

	defer func() {
		if retErr != nil {
			if err := deleteNode(parentID); err != nil {
				t.Log(err)
			}
		}
	}()

	sourceID, ok := assigned.Uids["data_source"]
	if !ok {
		return nil, nil, errors.New("cannot find source id after mutation in setup")
	}

	sourceID1, ok := assigned.Uids["data_source1"]
	if !ok {
		return nil, nil, errors.New("cannot find source id for equip_3.csv after mutation in setup")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(sourceID); err != nil {
				t.Log(err)
			}
		}
	}()

	eqType := &v1.EquipmentType{
		Type:       "MyType1",
		SourceID:   sourceID,
		SourceName: "equip_3.csv",
		ParentID:   parentID,
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
				IsDisplayed:  true,
				MappedTo:     "mapping_2",
			},
			&v1.Attribute{
				Name:        "attr2.1",
				Type:        v1.DataTypeInt,
				IsDisplayed: true,
				MappedTo:    "mapping_2.1",
			},
			&v1.Attribute{
				Name:         "attr3",
				Type:         v1.DataTypeFloat,
				IsSearchable: true,
				IsDisplayed:  true,
				MappedTo:     "mapping_3",
			},
			&v1.Attribute{
				Name:     "attr3.1",
				Type:     v1.DataTypeFloat,
				MappedTo: "mapping_3.1",
			},
			&v1.Attribute{
				Name:        "attr4",
				Type:        v1.DataTypeString,
				IsDisplayed: true,
				MappedTo:    "mapping_4",
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

	repo := NewLicenseRepository(dgClient)
	eqType, err = repo.CreateEquipmentType(context.Background(), eqType, []string{})
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(eqType.ID); err != nil {
				t.Log(err)
			}
		}
	}()

	eqType1 := &v1.EquipmentType{
		Type:       "MyType2",
		SourceID:   sourceID1,
		SourceName: "equip_4.csv",
		ParentID:   eqType.ID,
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
				IsDisplayed:  true,
				MappedTo:     "mapping_2",
			},
			&v1.Attribute{
				Name:        "attr2.1",
				Type:        v1.DataTypeInt,
				IsDisplayed: true,
				MappedTo:    "mapping_2.1",
			},
			&v1.Attribute{
				Name:         "attr3",
				Type:         v1.DataTypeFloat,
				IsSearchable: true,
				IsDisplayed:  true,
				MappedTo:     "mapping_3",
			},
			&v1.Attribute{
				Name:     "attr3.1",
				Type:     v1.DataTypeFloat,
				MappedTo: "mapping_3.1",
			},
			&v1.Attribute{
				Name:        "attr4",
				Type:        v1.DataTypeString,
				IsDisplayed: true,
				MappedTo:    "mapping_4",
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
			&v1.Attribute{
				Name:               "p_attr",
				Type:               v1.DataTypeString,
				IsParentIdentifier: true,
				IsDisplayed:        true,
				MappedTo:           "p_mapping",
			},
		},
	}

	eqType1, err = repo.CreateEquipmentType(context.Background(), eqType1, []string{})
	if err != nil {
		return nil, nil, err
	}

	return []*v1.EquipmentType{
			eqType,
			eqType1,
		}, func() error {
			if err := deleteNode(parentID); err != nil {
				return err
			}
			if err := deleteNode(sourceID); err != nil {
				return err
			}
			if err := deleteNode(sourceID1); err != nil {
				return err
			}
			if err := deleteNode(eqType.ID); err != nil {
				return err
			}
			if err := deleteNode(eqType1.ID); err != nil {
				return err
			}
			return nil
		}, nil
}
