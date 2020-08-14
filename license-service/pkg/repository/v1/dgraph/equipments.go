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
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
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

// CreateEquipmentType implements Licence CreateEquipmentType function
func (lr *LicenseRepository) CreateEquipmentType(ctx context.Context, eqType *v1.EquipmentType, scopes []string) (retType *v1.EquipmentType, retErr error) {
	nquads := nquadsForEquipment(eqType)
	mu := &api.Mutation{
		Set: nquads,
		//	CommitNow: true,
	}
	fmt.Printf("eqtype: %+v", eqType)
	txn := lr.dg.NewTxn()

	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateEquipmentType - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateEquipmentType - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateEquipmentType - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateEquipmentType - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("EquipmentType", eqType.Type),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", eqType.Attributes)...)
		logger.Log.Error("dgraph/CreateEquipmentType -Mutate ", fields...)
		return nil, fmt.Errorf("dgraph/CreateEquipmentType - cannot create equipment type :%s", eqType.Type)
	}

	assignIDsEquipemntType(assigned.Uids, eqType)

	schema := schemaForEquipmentType(eqType.Type, eqType.Attributes)
	if schema == "" {
		return eqType, nil
	}
	fmt.Printf("eqtype1: %+v", eqType)
	if err := lr.dg.Alter(context.Background(), &api.Operation{
		Schema: schema,
	}); err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("Schema", schema),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", eqType.Attributes)...)
		logger.Log.Error("dgraph/CreateEquipmentType - Alter ", fields...)
		return nil, fmt.Errorf("dgraph/CreateEquipmentType - cannot create schema for equipment type type :%s", eqType.Type)
	}

	fmt.Printf("eqtype: %+v", eqType)
	return eqType, nil
}

// EquipmentTypes implements Licence EquipmentTypes function
func (lr *LicenseRepository) EquipmentTypes(ctx context.Context, scopes []string) ([]*v1.EquipmentType, error) {
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

func (lr *LicenseRepository) equipmentTypeByType(ctx context.Context, typ string) (*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:eq(metadata.equipment.type,` + typ + `)){
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
	if len(data.EqTypes) == 0 {
		return nil, v1.ErrNoData
	}
	return convertEquipType(data.EqTypes[0]), nil
}

func assignIDsEquipemntType(ids map[string]string, eqType *v1.EquipmentType) {
	if len(ids) == 0 {
		logger.Log.Error("dgraph/assignIDsEquipemntType - ", zap.String("reason", "cannot assign id to equipmentType id map is empty"),
			zap.String("EquipmentType", eqType.Type), zap.String("idMap", fmt.Sprintf("%+v", ids)))
		return
	}
	eqTypeID, ok := ids[eqType.Type]
	if !ok {
		logger.Log.Error("dgraph/assignIDsEquipemntType - ", zap.String("reason", "cannot assign id to equipmentType"),
			zap.String("EquipmentType", eqType.Type), zap.String("idMap", fmt.Sprintf("%+v", ids)))
	} else {
		eqType.ID = eqTypeID
	}

	assignIDsEquipmentAttributes(ids, eqType.Type, eqType.Attributes)

}

func assignIDsEquipmentAttributes(ids map[string]string, typ string, attrb []*v1.Attribute) {
	for i, attr := range attrb {
		attrID, ok := ids[attr.Name]
		if !ok {
			fields := []zap.Field{
				zap.String("reason", "cannot assign id to attribute"),
				zap.String("EquipmentType", typ),
				zap.String("idMap", fmt.Sprintf("%+v", ids)),
			}
			fields = append(fields, attributeZapFields(fmt.Sprintf("EquipmentType.Attributes[%d]", i), attrb[i])...)
			logger.Log.Error("dgraph/assignIDsEquipemntType - ", fields...)
			continue
		}
		attr.ID = attrID
	}
}

func nquadsForEquipment(eqType *v1.EquipmentType) []*api.NQuad {
	equipID := blankID(eqType.Type)
	var nquads []*api.NQuad
	// assign predicate for equipment type
	nquads = append(nquads,
		&api.NQuad{
			Subject:     equipID,
			Predicate:   "metadata.equipment.type",
			ObjectValue: stringObjectValue(eqType.Type),
		},
		&api.NQuad{
			Subject:     equipID,
			Predicate:   "dgraph.type",
			ObjectValue: stringObjectValue("MetadataEquipment"),
		},
	)

	if eqType.ParentID != "" {
		nquads = append(nquads, &api.NQuad{
			Subject:   equipID,
			Predicate: "metadata.equipment.parent",
			ObjectId:  eqType.ParentID,
		})
	}

	// assign predicate for source type
	nquads = append(nquads, &api.NQuad{
		Subject:   equipID,
		Predicate: "metadata.equipment.source",
		ObjectId:  eqType.SourceID,
	})

	nquads = append(nquads, nquadsForAllAttributes(equipID, eqType.Attributes)...)
	return nquads

}

// attribute.name string @index(exact) .
// attribute.searchable bool .
// attribute.identifier bool .
// attribute.displayed bool .
// attibute.mapped_to string .

func nquadsForAllAttributes(equipID string, attributes []*v1.Attribute) []*api.NQuad {

	var nquads []*api.NQuad
	for _, attr := range attributes {
		attrBlankID, nqs := nquadsForAttributes(attr)
		nquads = append(nquads, &api.NQuad{
			Subject:   equipID,
			Predicate: "metadata.equipment.attribute",
			ObjectId:  attrBlankID,
		})
		nquads = append(nquads, nqs...)
	}

	return nquads
}

func nquadsForAttributes(attr *v1.Attribute) (string, []*api.NQuad) {
	blankID := blankID(attr.Name)
	fmt.Println(blankID)
	return blankID, []*api.NQuad{
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.name",
			ObjectValue: stringObjectValue(attr.Name),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.type",
			ObjectValue: intObjectValue(int64(attr.Type)),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.schema_name",
			ObjectValue: stringObjectValue(attr.Name),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.searchable",
			ObjectValue: boolObjectValue(attr.IsSearchable),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.displayed",
			ObjectValue: boolObjectValue(attr.IsDisplayed),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.identifier",
			ObjectValue: boolObjectValue(attr.IsIdentifier),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.parentIdentifier",
			ObjectValue: boolObjectValue(attr.IsParentIdentifier),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "attribute.mapped_to",
			ObjectValue: stringObjectValue(attr.MappedTo),
		},
	}
}

func stringObjectValue(val string) *api.Value {
	return &api.Value{
		Val: &api.Value_StrVal{
			StrVal: val,
		},
	}
}

func boolObjectValue(val bool) *api.Value {
	return &api.Value{
		Val: &api.Value_BoolVal{
			BoolVal: val,
		},
	}
}

func intObjectValue(val int64) *api.Value {
	return &api.Value{
		Val: &api.Value_IntVal{
			IntVal: val,
		},
	}
}

func blankID(id string) string {
	return "_:" + id
}

func attributeZapFields(name string, attr *v1.Attribute) []zap.Field {
	return []zap.Field{
		zap.String(name+".name", attr.Name),
		zap.String(name+".dataType", attr.Type.String()),
		zap.Bool(name+".IsIdentifier", attr.IsIdentifier),
		zap.Bool(name+".IsDisplayed", attr.IsDisplayed),
		zap.Bool(name+".IsSearchable", attr.IsSearchable),
		zap.String(name+".MappedTo", attr.MappedTo),
	}
}

func attributesZapFields(name string, attrs []*v1.Attribute) []zap.Field {
	var fields []zap.Field
	for idx, attr := range attrs {
		fields = append(fields, attributeZapFields(fmt.Sprintf("%v[%d]", name, idx), attr)...)
	}
	return fields
}

func schemaForEquipmentType(typ string, attrb []*v1.Attribute) string {
	//typ := eqType.Type
	equipType := "type Equipment" + typ + " { \n"
	equipTypeFields := []string{
		"type_name",
		"scopes",
		"updated",
		"created",
		"equipment.users",
	}
	typeName := "equipment." + typ
	schema := ""
	for _, attr := range attrb {
		if attr.IsIdentifier {
			equipTypeFields = append(equipTypeFields, "equipment.id")
			// we always map identifier to equipment.id predicate
			// so we can skip this here as schema for that is already
			// created
			continue
		}
		if attr.IsParentIdentifier {
			equipTypeFields = append(equipTypeFields, "equipment.parent")
			// we always map identifier to equipment.id predicate
			// so we can skip this here as schema for that is already
			// created
			continue
		}
		equipTypeFields = append(equipTypeFields, typeName+"."+replaceSpaces(attr.Name))
		schema += schemaForAttribute(typeName, attr) + "\n"
	}
	return schema + "\n\n" + equipType + strings.Join(equipTypeFields, "\n") + "\n }\n"
}

// some of the special characters are not allowed in
func replaceSpaces(mappedTo string) string {
	return strings.Replace(strings.TrimSpace(mappedTo), " ", "_", -1)
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
