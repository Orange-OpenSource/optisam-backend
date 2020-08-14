// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"os"
	"strings"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func metadataBySourceName(name string, metadata []*v1.Metadata) int {
	for idx, m := range metadata {
		if m.Source == name {
			return idx
		}
	}
	return -1
}

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

	repo := NewEquipmentRepository(dgClient)
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

func equipmentsJSONFromCSV(filename string, eqType *v1.EquipmentType, ignoreDisplayed bool) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(file)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errors.New("no data in: " + filename)
	}

	headers := records[0]

	pkAttr, err := eqType.PrimaryKeyAttribute()
	if err != nil {
		return nil, err
	}

	records = records[1:]
	data := []string{}

	for _, rec := range records {
		recJSON := ""
		for idx, val := range rec {
			if headers[idx] == pkAttr.MappedTo {
				recJSON = fmt.Sprintf(`"%s":"%s",`, pkAttr.Name, val) + recJSON
				continue
			}
			i := attributeByMapping(headers[idx], eqType.Attributes)
			if i == -1 {
				// Continue log this
				continue
			}

			attr := eqType.Attributes[i]

			if attr.IsParentIdentifier {
				continue
			}

			if ignoreDisplayed {
				if !attr.IsDisplayed {
					continue
				}
			}

			switch attr.Type {
			case v1.DataTypeString:
				recJSON += fmt.Sprintf(`"%s":"%s",`, attr.Name, val)
			case v1.DataTypeInt:
				recJSON += fmt.Sprintf(`"%s":%s,`, attr.Name, val)
			case v1.DataTypeFloat:
				recJSON += fmt.Sprintf(`"%s":%s.000000,`, attr.Name, val)
			default:
				// TODO: unsupported data type log this
			}
		}
		recJSON = `{` + strings.TrimSuffix(recJSON, ",") + `}`
		data = append(data, recJSON)
	}

	return data, nil
}

func attributeByMapping(mappedTo string, attributes []*v1.Attribute) int {
	for idx := range attributes {
		if attributes[idx].MappedTo == mappedTo {
			return idx
		}
	}
	return -1
}

func getUIDForEquipmentXIDWithType(xid, eqType string) (string, error) {
	type id struct {
		ID string
	}
	type data struct {
		IDs []*id
	}

	resp, err := dgClient.NewTxn().Query(context.Background(), `{
	        IDs(func: eq(equipment.type,`+eqType+`))@filter(eq(equipment.id,`+xid+`)){
				ID:uid
			}
	}`)
	if err != nil {
		return "", err
	}

	var d data
	if err := json.Unmarshal(resp.Json, &d); err != nil {
		return "", err
	}
	if len(d.IDs) == 0 {
		return "", v1.ErrNoData
	}
	return d.IDs[0].ID, nil
}

func TestEquipmentRepository_Equipments(t *testing.T) {

	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}
	defer func() {
		assert.Empty(t, cleanup(), "error  not expected from clean up")
	}()

	//	return
	eqType := eqTypes[0]

	equipments, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_3.csv", eqType, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}
	equipmentsNew, err := equipmentsJSONFromCSV("testdata/scope3/v1/equip_3.csv", eqType, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}
	eqType1 := eqTypes[1]

	equipmentsPar, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_4.csv", eqType1, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	type args struct {
		ctx    context.Context
		eqType *v1.EquipmentType
		params *v1.QueryEquipments
		scopes []string
	}
	tests := []struct {
		name    string
		r       *EquipmentRepository
		args    args
		want    int32
		want1   json.RawMessage
		wantErr bool
	}{
		{name: "success : some sorting - product filter",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType1,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
					ProductFilter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "swidtag",
								FilterValue: "ORAC001",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipmentsPar[0], equipmentsPar[1], equipmentsPar[2]}, ",") + "]"),
		},
		{name: "success : no sort by choose default,page size 2 offset 1",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  2,
					Offset:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sort by non displayable attribute",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr4",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sort by unknown attribute",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr4.111",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sorting, searching by multiple params",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "attr1",
								FilterValue: "equip3",
							},
							&v1.Filter{
								FilterKey:   "attr4.1",
								FilterValue: "mmmmmm34_12",
							},
							&v1.Filter{
								FilterKey:   "attr2",
								FilterValue: 333333322,
							},
							&v1.Filter{
								FilterKey:   "attr3",
								FilterValue: 333333332,
							},
							&v1.Filter{
								FilterKey:   "attr3.xxx",
								FilterValue: 333333332,
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  1,
			want1: []byte("[" + strings.Join([]string{equipments[1]}, ",") + "]"),
		},
		{name: "success : sorting on non-displayable attribute",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr3.1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sorting on parent key",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType1,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "p_attr",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  7,
			want1: []byte("[" + strings.Join([]string{equipmentsPar[0], equipmentsPar[1], equipmentsPar[2]}, ",") + "]"),
		},
		{name: "success : some sorting - scope3",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope3"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipmentsNew[0], equipmentsNew[1], equipmentsNew[2]}, ",") + "]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.r.Equipments(tt.args.ctx, tt.args.eqType, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.Equipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EquipmentRepository.Equipments() got = %v, want %v", got, tt.want)
			}

			fields := strings.Split(string(got1), ",")

			idIndexes := []int{}
			for idx, field := range fields {
				if strings.Contains(field, `[{"ID"`) {
					if idx < len(fields)-1 {
						fields[idx+1] = "[{" + fields[idx+1]
					}
					idIndexes = append(idIndexes, idx)
					continue
				}
				if strings.Contains(field, `{"ID"`) {
					if idx < len(fields)-1 {
						fields[idx+1] = "{" + fields[idx+1]
					}
					idIndexes = append(idIndexes, idx)
				}
			}

			// remove indexes from fields
			idLessfields := make([]string, 0, len(fields)-len(idIndexes))
			count := 0
			for idx := range fields {
				if count < len(idIndexes) && idx == idIndexes[count] {
					count++
					continue
				}
				idLessfields = append(idLessfields, fields[idx])
			}

			assert.Equal(t, strings.Join(strings.Split(string(tt.want1), ","), ","), strings.Join(idLessfields, ","))
		})
	}
}

func TestEquipmentRepository_Equipment(t *testing.T) {
	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}
	defer func() {
		assert.Empty(t, cleanup(), "error  not expected from clean up")
	}()

	//	return
	eqType := eqTypes[0]
	equipments, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_3.csv", eqType, false)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}
	equipmentsNew, err := equipmentsJSONFromCSV("testdata/scope3/v1/equip_3.csv", eqType, false)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	// uid, err := getUIDForEquipmentXIDWithType("equip3_1", "MyType1")
	// if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
	// 	return
	// }
	// uidNew, err := getUIDForEquipmentXIDWithType("equip33_1", "MyType1")
	// if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
	// 	return
	// }

	type args struct {
		ctx    context.Context
		eqType *v1.EquipmentType
		id     string
		scopes []string
	}
	tests := []struct {
		name    string
		r       *EquipmentRepository
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{name: "success ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				id:     "equip3_1",
				scopes: []string{"scope1", "scope2"},
			},
			want: []byte(equipments[0]),
		},
		{name: "no node exists ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				id:     "",
				scopes: []string{"scope1", "scope2"},
			},
			wantErr: true,
		},
		{name: "success - scope 3 ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				eqType: eqType,
				id:     "equip33_1",
				scopes: []string{"scope3"},
			},
			want: []byte(equipmentsNew[0]),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Equipment(tt.args.ctx, tt.args.eqType, tt.args.id, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.Equipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				fields := strings.Split(string(got), ",")

				idIndexes := []int{}
				for idx, field := range fields {
					if strings.Contains(field, `[{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "[{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
						continue
					}
					if strings.Contains(field, `{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
					}
				}

				// remove indexes from fields
				idLessfields := make([]string, 0, len(fields)-len(idIndexes))
				count := 0
				for idx := range fields {
					if count < len(idIndexes) && idx == idIndexes[count] {
						count++
						continue
					}
					idLessfields = append(idLessfields, fields[idx])
				}

				assert.Equal(t, strings.Join(strings.Split(string(tt.want), ","), ","), strings.Join(idLessfields, ","))
			}
		})
	}
}

func TestEquipmentRepository_EquipmentParent(t *testing.T) {
	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}

	defer func() {
		assert.Empty(t, cleanup(), "error  not expected from clean up")
	}()

	//	return

	eqTypeParent := eqTypes[0]
	eqType := eqTypes[1]

	equipments, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_3.csv", eqTypeParent, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	uid, err := getUIDForEquipmentXIDWithType("equip4_1", "MyType2")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	equipmentsNew, err := equipmentsJSONFromCSV("testdata/scope3/v1/equip_3.csv", eqTypeParent, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	uidNew, err := getUIDForEquipmentXIDWithType("equip44_1", "MyType2")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	uid1, err := getUIDForEquipmentXIDWithType("equip4_7", "MyType2")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	parID, err := getUIDForEquipmentXIDWithType("equip3_3", "MyType1")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	// SETUP
	if err := deleteNode(parID); err != nil {
		t.Log(err)
	}

	type args struct {
		ctx          context.Context
		eqType       *v1.EquipmentType
		parentEqType *v1.EquipmentType
		id           string
		scopes       []string
	}
	tests := []struct {
		name        string
		r           *EquipmentRepository
		args        args
		wantRecords int32
		want        json.RawMessage
		wantErr     bool
	}{
		{name: "success ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:          context.Background(),
				eqType:       eqType,
				parentEqType: eqTypeParent,
				id:           uid,
				scopes:       []string{"scope1", "scope2"},
			},
			wantRecords: 1,
			want:        []byte("[" + equipments[0] + "]"),
		},
		{name: "no node exists ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:          context.Background(),
				eqType:       eqType,
				parentEqType: eqTypeParent,
				id:           "0x5678",
				scopes:       []string{"scope1", "scope2"},
			},
			wantErr: true,
		},
		{name: "node exists but no data ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:          context.Background(),
				eqType:       eqType,
				parentEqType: eqTypeParent,
				id:           uid1,
				scopes:       []string{"scope1", "scope2"},
			},
			wantErr: true,
		},
		{name: "success - scope3 ",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:          context.Background(),
				eqType:       eqType,
				parentEqType: eqTypeParent,
				id:           uidNew,
				scopes:       []string{"scope3"},
			},
			wantRecords: 1,
			want:        []byte("[" + equipmentsNew[0] + "]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numOfRecords, got, err := tt.r.EquipmentParents(tt.args.ctx, tt.args.eqType, tt.args.parentEqType, tt.args.id, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.Equipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !assert.Equal(t, tt.wantRecords, numOfRecords, "number of records should be equal") {
					return
				}
				fields := strings.Split(string(got), ",")

				idIndexes := []int{}
				for idx, field := range fields {
					if strings.Contains(field, `[{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "[{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
						continue
					}
					if strings.Contains(field, `{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
					}
				}

				// remove indexes from fields
				idLessfields := make([]string, 0, len(fields)-len(idIndexes))
				count := 0
				for idx := range fields {
					if count < len(idIndexes) && idx == idIndexes[count] {
						count++
						continue
					}
					idLessfields = append(idLessfields, fields[idx])
				}

				assert.Equal(t, strings.Join(strings.Split(string(tt.want), ","), ","), strings.Join(idLessfields, ","))
			}
		})
	}
}

func TestEquipmentRepository_EquipmentChild(t *testing.T) {
	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}

	//	return
	defer func() {
		assert.Empty(t, cleanup(), "error not expected from clean up")
	}()

	//	return

	eqType := eqTypes[0]
	eqTypeChild := eqTypes[1]

	equipments, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_4.csv", eqTypeChild, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	uid, err := getUIDForEquipmentXIDWithType("equip3_1", "MyType1")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	equipmentsNew, err := equipmentsJSONFromCSV("testdata/scope3/v1/equip_4.csv", eqTypeChild, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	uidNew, err := getUIDForEquipmentXIDWithType("equip33_1", "MyType1")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}
	uid1, err := getUIDForEquipmentXIDWithType("equip3_3", "MyType1")
	if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
		return
	}

	type args struct {
		ctx         context.Context
		eqType      *v1.EquipmentType
		childEqType *v1.EquipmentType
		id          string
		params      *v1.QueryEquipments
		scopes      []string
	}
	tests := []struct {
		name    string
		r       *EquipmentRepository
		args    args
		setup   func() error
		want    int32
		want1   json.RawMessage
		wantErr bool
	}{
		{name: "success : some sorting",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[3], equipments[4]}, ",") + "]"),
		},
		{name: "success : some sorting not primary key",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr2",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[3], equipments[4]}, ",") + "]"),
		},
		{name: "success : no sort by choose default,page size 2 offset 1",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  2,
					Offset:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipments[3], equipments[4]}, ",") + "]"),
		},
		{name: "success : sort by non displayable attribute",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  5,
					Offset:    0,
					SortBy:    "attr4",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[3], equipments[4], equipments[5]}, ",") + "]"),
		},
		{name: "success : sort by unknown attribute",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr4.111",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[3], equipments[4]}, ",") + "]"),
		},
		{name: "success : sorting, searching by multiple params",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "attr1",
								FilterValue: "equip4",
							},
							&v1.Filter{
								FilterKey:   "attr4.1",
								FilterValue: "mmmmmm44_1",
							},
							&v1.Filter{
								FilterKey:   "attr2",
								FilterValue: 333333424,
							},
							&v1.Filter{
								FilterKey:   "attr4",
								FilterValue: 333333434,
							},
							&v1.Filter{
								FilterKey:   "attr3.xxx",
								FilterValue: 333333332,
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[3], equipments[4], equipments[5]}, ",") + "]"),
		},
		{name: "no node exists",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          "0x6677",
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			wantErr: true,
		},
		{name: "node exists - but no data",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uid1,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			setup: func() error {
				childID1, err := getUIDForEquipmentXIDWithType("equip4_7", "MyType2")
				if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
					return err
				}
				childID2, err := getUIDForEquipmentXIDWithType("equip4_3", "MyType2")
				if !assert.Empty(t, err, "error not expected from getUIDForEquipmentXIDWithType") {
					return err
				}
				if err := deleteNode(childID1); err != nil {
					return err
				}
				if err := deleteNode(childID2); err != nil {
					return err
				}
				return nil
			},
			wantErr: true,
		},
		{name: "success : some sorting - scope3",
			r: NewEquipmentRepository(dgClient),
			args: args{
				ctx:         context.Background(),
				id:          uidNew,
				eqType:      eqType,
				childEqType: eqTypeChild,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope3"},
			},
			want:  4,
			want1: []byte("[" + strings.Join([]string{equipmentsNew[0], equipmentsNew[3], equipmentsNew[4]}, ",") + "]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if !assert.Empty(t, tt.setup(), "error not expected from getUIDForEquipmentXIDWithType") {
					return
				}
			}
			numOfRecords, got, err := tt.r.EquipmentChildren(tt.args.ctx, tt.args.eqType, tt.args.childEqType, tt.args.id, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.Equipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !assert.Equal(t, tt.want, numOfRecords, "number of records should be equal") {
					return
				}
				fields := strings.Split(string(got), ",")

				idIndexes := []int{}
				for idx, field := range fields {
					if strings.Contains(field, `[{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "[{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
						continue
					}
					if strings.Contains(field, `{"ID"`) {
						if idx < len(fields)-1 {
							fields[idx+1] = "{" + fields[idx+1]
						}
						idIndexes = append(idIndexes, idx)
					}
				}

				// remove indexes from fields
				idLessfields := make([]string, 0, len(fields)-len(idIndexes))
				count := 0
				for idx := range fields {
					if count < len(idIndexes) && idx == idIndexes[count] {
						count++
						continue
					}
					idLessfields = append(idLessfields, fields[idx])
				}

				assert.Equal(t, strings.Join(strings.Split(string(tt.want1), ","), ","), strings.Join(idLessfields, ","))
			}
		})
	}
}

func compareEquipmentProductAll(t *testing.T, name string, exp []*v1.EquipmentProduct, act []*v1.EquipmentProduct) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareEquipmentProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareEquipmentProduct(t *testing.T, name string, exp *v1.EquipmentProduct, act *v1.EquipmentProduct) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)

	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)

	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)
}
