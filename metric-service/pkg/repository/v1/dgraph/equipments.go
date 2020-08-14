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
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"

	"strings"
)

const (
	eqTypeFields = ` 
	ID:         uid
	Type:       metadata.equipment.type
	DataSource: metadata.equipment.source {
		ID:     uid
    Source: metadata.source
	}
	Parent:     metadata.equipment.parent{
		ID:       uid
    TypeName: metadata.equipment.type
	}
	Attributes: metadata.equipment.attribute{
		ID:			            uid
		Name:               attribute.name
		Type:               attribute.type
		IsDisplayed:        attribute.displayed
		IsIdentifier:       attribute.identifier
		IsParentIdentifier: attribute.parentIdentifier
		IsSearchable:       attribute.searchable
		MappedTo:           attribute.mapped_to
		SchemaName:         attribute.schema_name
		}`
)

type eqTypeParent struct {
	ID       string
	TypeName string
}

type eqTypeDataSource struct {
	ID     string
	Source string
}

type equipmentType struct {
	ID         string
	Type       string
	DataSource *eqTypeDataSource
	Parent     *eqTypeParent
	Attributes []*v1.Attribute
}

func convertEquipTypeAll(eqTypes []*equipmentType) []*v1.EquipmentType {
	types := make([]*v1.EquipmentType, len(eqTypes))
	for i := range eqTypes {
		types[i] = convertEquipType(eqTypes[i])
	}
	return types
}

func convertEquipType(eq *equipmentType) *v1.EquipmentType {
	eqType := &v1.EquipmentType{
		ID:         eq.ID,
		Type:       eq.Type,
		Attributes: eq.Attributes,
	}
	if eq.DataSource != nil {
		eqType.SourceID = eq.DataSource.ID
		eqType.SourceName = eq.DataSource.Source
	}

	if eq.Parent != nil {
		eqType.ParentID = eq.Parent.ID
		eqType.ParentType = eq.Parent.TypeName
	}

	return eqType
}

// EquipmentTypes implements Licence EquipmentTypes function
func (lr *MetricRepository) EquipmentTypes(ctx context.Context, scopes []string) ([]*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:has(metadata.equipment.type)){
		  ` + eqTypeFields + `
		}
	}
	`
	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/EquipmentTypes - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/EquipmentTypes - cannot complete query")
	}

	type eqTypes struct {
		EqTypes []*equipmentType
	}

	data := eqTypes{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("dgraph/EquipmentTypes - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/EquipmentTypes - cannot unmarshal Json object")
	}
	return convertEquipTypeAll(data.EqTypes), nil
}

func schemaForAttribute(name string, attr *v1.Attribute) string {

	// TODO Change this to attr.schema_name
	name += "." + replaceSpaces(attr.Name) + ":"
	switch attr.Type {
	case v1.DataTypeString:
		name += " string "
		if attr.IsSearchable {
			// check data Type
			name += " @index(trigram) "
		}
	case v1.DataTypeInt:
		name += " int "
		if attr.IsSearchable {
			// check data Type
			name += " @index(int) "
		}
	case v1.DataTypeFloat:
		name += " float "
		if attr.IsSearchable {
			// check data Type
			name += " @index(float) "
		}
	default:
		name += " string "
	}

	name += "."
	return name
}

func replaceSpaces(mappedTo string) string {
	return strings.Replace(strings.TrimSpace(mappedTo), " ", "_", -1)
}
