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
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"strconv"
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
	Scopes: scopes
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
	Scopes     []string
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
		Scopes:     eq.Scopes,
	}
	if eq.Parent != nil {
		eqType.ParentType = eq.Parent.TypeName
		eqType.ParentID = eq.Parent.ID
	}
	eqType.SourceID = eq.DataSource.ID
	eqType.SourceName = eq.DataSource.Source

	return eqType
}

// CreateEquipmentType implements Licence CreateEquipmentType function
func (lr *EquipmentRepository) CreateEquipmentType(ctx context.Context, eqType *v1.EquipmentType, scopes []string) (retType *v1.EquipmentType, retErr error) {
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
func (lr *EquipmentRepository) EquipmentTypes(ctx context.Context, scopes []string) ([]*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:has(metadata.equipment.type)) ` + agregateFilters(scopeFilters(scopes)) + `{
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

//EquipmentTypeByType ...
func (lr *EquipmentRepository) EquipmentTypeByType(ctx context.Context, typ string, scopes []string) (*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:eq(metadata.equipment.type,` + typ + `))` + agregateFilters(scopeFilters(scopes)) + `{
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

func (lr *EquipmentRepository) equipmentTypeByType(ctx context.Context, typ string, scopes []string) (*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:eq(metadata.equipment.type,` + typ + `))` + agregateFilters(scopeFilters(scopes)) + `{
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

// metadata.equipment.type string @index(exact) .
// metadata.equipment.attribute uid .

// EquipmentWithID implements Licence EquipmentWithID function  TODO :EquipmentTypeByID
func (lr *EquipmentRepository) EquipmentWithID(ctx context.Context, id string, scopes []string) (*v1.EquipmentType, error) {
	q := `{
		Equipment(func: uid(` + id + `)) ` + agregateFilters(scopeFilters(scopes)) + `{
			` + eqTypeFields + `
		}
	  }`

	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/EquipmentWitID - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/EquipmentWithID - cannot complete query")
	}

	type equipment struct {
		Equipment []*equipmentType
	}

	data := equipment{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("dgraph/EquipmentWithId - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/EquipmentWithID - cannot unmarshal Json object")
	}
	if len(data.Equipment) == 0 {
		return nil, v1.ErrNoData
	}
	return convertEquipType(data.Equipment[0]), nil
}

// DeleteEquipmentType implements Equipment DeleteEquipmentType function
func (lr *EquipmentRepository) DeleteEquipmentType(ctx context.Context, eqType, scope string) error {
	query := `query {
		var(func: eq(metadata.equipment.type,` + eqType + `)) @filter(eq(scopes,` + scope + `)){
			equipType as uid
		}
		`
	delete := `
			uid(equipType) * * .
	`
	set := `
			uid(equipType) <Recycle> "true" .
	`
	query += `
	}`
	muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
	logger.Log.Info(query)
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{muDelete},
		CommitNow: true,
	}
	if _, err := lr.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DeleteEquipmentType - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf("DeleteEquipmentType - cannot complete query transaction")
	}
	return nil
}

// UpdateEquipmentType implements Licence UpdateEquipmentType function
func (lr *EquipmentRepository) UpdateEquipmentType(ctx context.Context, id string, typ string, req *v1.UpdateEquipmentRequest, scopes []string) (retType []*v1.Attribute, retErr error) {
	nquads := nquadsForAllAttributes(id, req.Attr)
	nquads = append(nquads, scopesNquad(scopes, id)...)
	if req.ParentID != "" {
		nquads = append(nquads, &api.NQuad{
			Subject:   id,
			Predicate: "metadata.equipment.parent",
			ObjectId:  req.ParentID,
		})
	}
	mu := &api.Mutation{
		Set: nquads,
	}
	txn := lr.dg.NewTxn()

	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/UpdateEquipmentType - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/UpdateEquipmentType - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/UpdateEquipmentType - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/UpdateEquipmentType - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("EquipmentType", typ),
			zap.String("ID", id),
			zap.String("ParentID", req.ParentID),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", req.Attr)...)
		logger.Log.Error("dgraph/UpdateEquipmentType -Mutate ", fields...)
		return nil, fmt.Errorf("dgraph/UpdateEquipmentType - cannot create equipment type :%s", typ)
	}

	assignIDsEquipmentAttributes(assigned.Uids, typ, req.Attr)
	schema := schemaForEquipmentType(typ, req.Attr)
	if schema == "" {
		return req.Attr, nil
	}
	if err := lr.dg.Alter(context.Background(), &api.Operation{
		Schema: schema,
	}); err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("EquipmentType", typ),
			zap.String("ID", id),
			zap.String("ParentID", req.ParentID),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", req.Attr)...)
		logger.Log.Error("dgraph/UpdateEquipmentType - Alter ", fields...)
		return nil, fmt.Errorf("dgraph/UpdateEquipmentType - cannot create schema for equipment type type :%s", typ)
	}

	return req.Attr, nil

}

// EquipmentTypeChildren  implements Equipment EquipmentTypeChildren function
func (lr *EquipmentRepository) EquipmentTypeChildren(ctx context.Context, eqTypeID string, depth int, scopes []string) ([]*v1.EquipmentType, error) {
	q := `
	{
		var (func: uid(` + eqTypeID + `))` + agregateFilters(scopeFilters(scopes)) + ` @recurse(depth: ` + strconv.Itoa(depth) + `, loop: false){
			childID as ~metadata.equipment.parent
		}
		EqTypes(func: uid(childID)){
		 ` + eqTypeFields + `
		}
	}
	`
	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/EquipmentTypeChildren - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/EquipmentTypeChildren - cannot complete query")
	}

	type eqTypes struct {
		EqTypes []*equipmentType
	}

	data := eqTypes{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("dgraph/EquipmentTypeChildren - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/EquipmentTypeChildren - cannot unmarshal Json object")
	}
	if len(data.EqTypes) == 0 {
		return nil, v1.ErrNoData
	}
	return convertEquipTypeAll(data.EqTypes), nil
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

	nquads = append(nquads, scopesNquad(eqType.Scopes, equipID)...)
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

//ListEquipmentsForProductAggregation ...
func (lr *EquipmentRepository) ListEquipmentsForProductAggregation(ctx context.Context, proAggName string, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables["$name"] = proAggName
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	q := `query EquipsForProductAggregation($name:string,$pagesize:string,$offset:string) {
		var(func: eq(product_aggregation.name, $name))` + agregateFilters(scopeFilters(scopes)) + `{
			ID_PRODUCTS as product_aggregation.products
	  	}
		var(func: uid(ID_PRODUCTS))` + agregateFilters(scopeFilters(scopes)) + `{
		  	IID as product.equipment @filter(eq(equipment.type,` + eqType.Type + `))
		}
		ID as var(func: uid(IID)) ` + agregateFilters(equipFilter(eqType, params.Filter)) + `{}
		NumOfRecords(func:uid(ID)){
			TotalCount:count(uid)
		}
		Equipments(func: uid(ID), ` + string(sortOrder) + `:` + equipSortBy(params.SortBy, eqType) + `,first:$pagesize,offset:$offset){
			 ` + equipQueryFields(eqType) + `
		}
	} `

	resp, err := lr.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("ListEquipmentsForProductAggregation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("ListEquipmentsForProductAggregation - cannot complete query transaction")
	}

	type Data struct {
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("Equipments - cannot unmarshal Json object")
	}

	if len(equipList.NumOfRecords) == 0 {
		return 0, nil, v1.ErrNoData
	}

	return equipList.NumOfRecords[0].TotalCount, equipList.Equipments, nil
}

func scopesNquad(scp []string, blankID string) []*api.NQuad {
	nquads := []*api.NQuad{}
	for _, sID := range scp {
		nquads = append(nquads, scopeNquad(sID, blankID)...)
	}
	return nquads
}

func scopeNquad(scope, uid string) []*api.NQuad {
	return []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
	}
}
