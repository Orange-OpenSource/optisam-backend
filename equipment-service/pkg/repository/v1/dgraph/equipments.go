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
// nolint: unused
func (r *EquipmentRepository) CreateEquipmentType(ctx context.Context, eqType *v1.EquipmentType, scopes []string) (retType *v1.EquipmentType, retErr error) {
	nquads := nquadsForEquipment(eqType)
	mu := &api.Mutation{
		Set: nquads,
		//	CommitNow: true,
	}

	logger.Log.Debug("eqTypes to be created ", zap.Any("EqType", eqType))

	txn := r.dg.NewTxn()
	r.mu.Lock()
	defer r.mu.Unlock()
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
	logger.Log.Debug("eqTypes created ", zap.Any("EqType", eqType))
	if err := r.dg.Alter(context.Background(), &api.Operation{
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

	return eqType, nil
}

// CreateAttributeNode implements Attribute node creation
func (r *EquipmentRepository) CreateAttributeNode(ctx context.Context, eqType *v1.EquipmentType, id string, scopes []string) (retType []*v1.Attribute, retErr error) {
	nquads := nquadsForAllAttributes(id, eqType.Attributes)
	mu := &api.Mutation{
		Set: nquads,
		//	CommitNow: true,
	}

	logger.Log.Debug("Attributes to be added ", zap.Any("Attribute", eqType.Attributes))

	txn := r.dg.NewTxn()
	r.mu.Lock()
	defer r.mu.Unlock()
	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateAttributeNode - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateAttributeNode - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateAttributeNode - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateAttributeNode - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("EquipmentType", eqType.Type),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", eqType.Attributes)...)
		logger.Log.Error("dgraph/CreateAttributeNode -Mutate ", fields...)
		return nil, fmt.Errorf("dgraph/CreateAttributeNode - cannot create equipment type :%s", eqType.Type)
	}

	assignIDsEquipmentAttributes(assigned.Uids, eqType.Type, eqType.Attributes)

	return eqType.Attributes, nil
}

// EquipmentTypes implements Licence EquipmentTypes function
func (r *EquipmentRepository) EquipmentTypes(ctx context.Context, scopes []string) ([]*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:has(metadata.equipment.type)) ` + agregateFilters(scopeFilters(scopes)) + `{
		  ` + eqTypeFields + `
		}
	}
	`
	resp, err := r.dg.NewTxn().Query(ctx, q)
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

// EquipmentTypeByType ...
func (r *EquipmentRepository) EquipmentTypeByType(ctx context.Context, typ string, scopes []string) (*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:eq(metadata.equipment.type,` + typ + `))` + agregateFilters(scopeFilters(scopes)) + `{
		 ` + eqTypeFields + `
		}
	}
	`
	resp, err := r.dg.NewTxn().Query(ctx, q)
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

// nolint: unused
func (r *EquipmentRepository) equipmentTypeByType(ctx context.Context, typ string, scopes []string) (*v1.EquipmentType, error) {
	q := `
	{
		EqTypes(func:eq(metadata.equipment.type,` + typ + `))` + agregateFilters(scopeFilters(scopes)) + `{
		 ` + eqTypeFields + `
		}
	}
	`
	resp, err := r.dg.NewTxn().Query(ctx, q)
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
func (r *EquipmentRepository) EquipmentWithID(ctx context.Context, id string, scopes []string) (*v1.EquipmentType, error) {
	q := `{
		Equipment(func: uid(` + id + `)) ` + agregateFilters(scopeFilters(scopes)) + `{
			` + eqTypeFields + `
		}
	  }`

	resp, err := r.dg.NewTxn().Query(ctx, q)
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
func (r *EquipmentRepository) DeleteEquipmentType(ctx context.Context, eqType, scope string) error {
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
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DeleteEquipmentType - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf("deleteEquipmentType - cannot complete query transaction")
	}
	return nil
}

// UpdateEquipmentType implements Licence UpdateEquipmentType function
func (r *EquipmentRepository) UpdateEquipmentType(ctx context.Context, id string, typ string, parentID string, req *v1.UpdateEquipmentRequest, scopes []string) (retType []*v1.Attribute, retErr error) {
	nquads := nquadsForAllAttributes(id, req.AddAttr)
	nquads = append(nquads, scopesNquad(scopes, id)...)
	r.mu.Lock()
	defer r.mu.Unlock()
	if req.ParentID != "" && req.ParentID != parentID {
		nquads = append(nquads, &api.NQuad{
			Subject:   id,
			Predicate: "metadata.equipment.parent",
			ObjectId:  req.ParentID,
		})
		delQuery := `query{
			  v as q(func: eq(metadata.equipment.type, "` + typ + `"))  ` + agregateFilters(scopeFilters(scopes)) + `
			}
		`
		delete := `
			uid(v) <metadata.equipment.parent> * .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete)}
		delreq := &api.Request{
			Query:     delQuery,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := r.dg.NewTxn().Do(ctx, delreq); err != nil {
			logger.Log.Error("dgraph/UpdateEquipmentType - unable to delete child node for parent update - ", zap.String("reason", err.Error()), zap.String("query", delQuery))
			return nil, fmt.Errorf("dgraph/UpdateEquipmentType - cannot complete query transaction")
		}
	}
	mu := &api.Mutation{
		Set: nquads,
	}
	txn := r.dg.NewTxn()
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
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", req.AddAttr)...)
		logger.Log.Error("dgraph/UpdateEquipmentType -Mutate ", fields...)
		return nil, fmt.Errorf("dgraph/UpdateEquipmentType - cannot create equipment type :%s", typ)
	}

	if len(req.UpdateAttr) != 0 {
		for _, attr := range req.UpdateAttr {
			q := `query {
				var(func: uid(` + attr.ID + `))  {
					ID as uid
				}
				}
				`
			set := `
			uid(ID) <attribute.schema_name> "` + attr.SchemaName + `" .
			uid(ID) <attribute.searchable> "` + strconv.FormatBool(attr.IsSearchable) + `" .
			uid(ID) <attribute.displayed> "` + strconv.FormatBool(attr.IsDisplayed) + `" .
			`
			req := &api.Request{
				Query: q,
				Mutations: []*api.Mutation{
					{
						SetNquads: []byte(set),
					},
				},
				CommitNow: true,
			}
			if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
				logger.Log.Error("dgraph/UpdateAttr - failed to mutate", zap.Error(err), zap.String("query", req.Query))
				return nil, errors.New("dgraph/UpdateAttr - failed to mutuate")
			}
		}
	}

	assignIDsEquipmentAttributes(assigned.Uids, typ, req.AddAttr)
	schema := schemaForEquipmentType(typ, req.AddAttr)
	if schema == "" {
		return req.AddAttr, nil
	}
	if err := r.dg.Alter(context.Background(), &api.Operation{
		Schema: schema,
	}); err != nil {
		fields := []zap.Field{
			zap.String("reason", err.Error()),
			zap.String("EquipmentType", typ),
			zap.String("ID", id),
			zap.String("ParentID", req.ParentID),
		}
		fields = append(fields, attributesZapFields("EquipmentType.Attributes", req.AddAttr)...)
		logger.Log.Error("dgraph/UpdateEquipmentType - Alter ", fields...)
		return nil, fmt.Errorf("dgraph/UpdateEquipmentType - cannot create schema for equipment type type :%s", typ)
	}

	return req.AddAttr, nil

}

// EquipmentTypeChildren  implements Equipment EquipmentTypeChildren function
func (r *EquipmentRepository) EquipmentTypeChildren(ctx context.Context, eqTypeID string, depth int, scopes []string) ([]*v1.EquipmentType, error) {
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
	resp, err := r.dg.NewTxn().Query(ctx, q)
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
	var nquads []*api.NQuad // nolint: prealloc
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
		{
			Subject:     blankID,
			Predicate:   "attribute.name",
			ObjectValue: stringObjectValue(attr.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.type",
			ObjectValue: intObjectValue(int64(attr.Type)),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.schema_name",
			ObjectValue: stringObjectValue(attr.Name),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.searchable",
			ObjectValue: boolObjectValue(attr.IsSearchable),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.displayed",
			ObjectValue: boolObjectValue(attr.IsDisplayed),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.identifier",
			ObjectValue: boolObjectValue(attr.IsIdentifier),
		},
		{
			Subject:     blankID,
			Predicate:   "attribute.parentIdentifier",
			ObjectValue: boolObjectValue(attr.IsParentIdentifier),
		},
		{
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

func attributesZapFields(name string, attrs []*v1.Attribute) []zap.Field { //nolint:unparam
	var fields []zap.Field
	for idx, attr := range attrs {
		fields = append(fields, attributeZapFields(fmt.Sprintf("%v[%d]", name, idx), attr)...)
	}
	return fields
}

func schemaForEquipmentType(typ string, attrb []*v1.Attribute) string {
	// typ := eqType.Type
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
	return strings.Replace(strings.TrimSpace(mappedTo), " ", "_", -1) // nolint: gocritic
}

func schemaForAttribute(name string, attr *v1.Attribute) string {

	// TODO Change this to attr.schema_name
	name += "." + replaceSpaces(attr.Name) + ":"
	switch attr.Type {
	case v1.DataTypeString:
		name += " string "
		if attr.IsSearchable {
			// check data Type
			name += " @index(trigram,exact) "
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

// ListEquipmentsForProductAggregation ...
func (r *EquipmentRepository) ListEquipmentsForProductAggregation(ctx context.Context, proAggName string, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
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
		var(func: eq(aggregation.name, $name))` + agregateFilters(scopeFilters(scopes)) + `{
			ID_PRODUCTS as aggregation.products
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

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("ListEquipmentsForProductAggregation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("listEquipmentsForProductAggregation - cannot complete query transaction")
	}

	type Data struct {
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("equipments - cannot unmarshal Json object")
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
		{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
	}
}

// ParentsHirerachyForEquipment ...
func (r *EquipmentRepository) ParentsHirerachyForEquipment(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, scopes ...string) ([]*v1.EquipmentInfo, error) {

	s1 := stringFilterSingle(v1.EqFilter, "equipment.type", equipType)

	q := `{
		ParentsHirerachy(func: uid(` + equipID + `)) @recurse(depth: ` + strconv.Itoa(int(hirearchyLevel)) + `) ` + agregateFilters(scopeFilters(scopes), []string{s1}) + ` {
			ID: uid
		 	EquipID: equipment.id
			Type: equipment.type
			Parent:equipment.parent
		}
	}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/parentsHirerachyForEquipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/parentsHirerachyForEquipment - cannot complete query transaction")
	}

	type data struct {
		ParentsHirerachy []*v1.EquipmentInfo
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/parentsHirerachyForEquipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/parentsHirerachyForEquipment - cannot unmarshal Json object")
	}
	return d.ParentsHirerachy, nil
}

func (r *EquipmentRepository) GetAllEquipmentsInHierarchy(ctx context.Context, equipType, endEquID string, scopes ...string) (*v1.EquipmentHierarchy, error) {
	query := BuildQueryForEquipmentParentHierarchy([]string{equipType}, scopes, endEquID)

	resp, err := r.dg.NewTxn().Query(ctx, query)
	if err != nil {
		logger.Log.Error("dgraph/getAllEquipmentsInHierarchy - ", zap.String("reason", err.Error()), zap.String("query", query))
		return nil, fmt.Errorf("dgraph/getAllEquipmentsInHierarchy -  cannot complete query transaction")
	}

	d := &v1.EquipmentHierarchy{}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/getAllEquipmentsInHierarchy -  ", zap.String("reason", err.Error()), zap.String("query", query))
		return nil, fmt.Errorf("dgraph/getAllEquipmentsInHierarchy - cannot unmarshal Json object")
	}

	return d, nil
}

func (r *EquipmentRepository) GetAllEquipmentForSpecifiedProduct(ctx context.Context, swidTag string, scopes ...string) (*v1.DeployedProducts, error) {
	query := BuildQueryForAllDeployedProducts([]string{swidTag}, scopes)

	resp, err := r.dg.NewTxn().Query(ctx, query)
	if err != nil {
		logger.Log.Error("dgraph/getAllEquipmentForSpecifiedProduct - ", zap.String("reason", err.Error()), zap.String("query", query))
		return nil, fmt.Errorf("dgraph/getAllEquipmentForSpecifiedProduct -  cannot complete query transaction")
	}

	fmt.Println(string(resp.Json))

	d := &v1.DeployedProducts{}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/getAllEquipmentForSpecifiedProduct - ", zap.String("reason", err.Error()), zap.String("query", query))
		return nil, fmt.Errorf("dgraph/getAllEquipmentForSpecifiedProduct -  cannot unmarshal Json object")
	}

	return d, nil
}

func BuildQueryForEquipmentParentHierarchy(endEquipmentType []string, scopes []string, id ...string) string {
	q := ``
	if endEquipmentType[0] == "vcenter" {

		q += `{
		VcenterEquipments(func:uid($ID))@filter(eq(scopes,[$Scopes])){ 
			id as uid
			equipment.id
	  		equipment.type
		}
		
		ClusterEquipments(func:uid(id)){
			~equipment.parent @filter(eq(equipment.type, cluster)){
	  			cid as uid
	  			equipment.id
	  			equipment.type
			}
 		}`
	} else {
		q += `{
		ClusterEquipments(func:uid($ID))@filter(eq(scopes,[$Scopes])){
				 cid as uid
				 equipment.id
				 equipment.type
		}`
	}
	q += `
		ServerEquipments(func:uid(cid)){
	  		~equipment.parent @filter(eq(equipment.type, server)){
				sid as uid
				equipment.id
				equipment.type
   			}
  		}

		SoftPartitionEquipments(func:uid(sid)){
	  		~equipment.parent @filter(eq(equipment.type, softpartition)){
				uid
				equipment.id
				equipment.type
  			}
  		}  
	}
	`

	return replacer(q, map[string]string{
		"$ID":        strings.Join(id, ","),
		"$Scopes":    strings.Join(scopes, ","),
		"$EndEqType": strings.Join(endEquipmentType, ","),
	})

}

func BuildQueryForAllDeployedProducts(swidTag []string, scopes []string) string {

	q := `{
			Products(func: eq(product.swidtag, $SWIDTAG))@filter( eq(scopes,[$Scopes]) AND eq(type_name,"product")) {    
	   			Swidtag : 		   product.swidtag
	   			Name :    		   product.name
	   			Version : 		   product.version
	   			Editor :  		   product.editor
 				product.equipment {
		   				uid
			   			equipment.id
	   					equipment.uid

 				}
			}
		}
	`

	return replacer(q, map[string]string{
		"$SWIDTAG": strings.Join(swidTag, ","),
		"$Scopes":  strings.Join(scopes, ","),
	})

}

func replacer(q string, params map[string]string) string {
	for key, val := range params {
		q = strings.Replace(q, key, val, -1) // nolint: gocritic
	}
	return q
}

func (r *EquipmentRepository) UpsertAllocateMetricInEquipmentHierarchy(ctx context.Context, mat *v1.MetricAllocationRequest, scope string) error {

	q := `query {
		var(func: eq(equipment.id,"` + mat.EquipmentID + `")) @filter(eq(scopes,[` + scope + `]) AND eq(product.swidtag, ` + mat.Swidtag + `) AND eq(type_name,"metricallocation")){
			ID as uid
		}
	}`
	set := `
	    uid(ID) <allocation.metric> "` + mat.AllocationMetric + `" .
		uid(ID) <product.swidtag> "` + mat.Swidtag + `" .
		uid(ID) <equipment.id> "` + mat.EquipmentID + `" .
		uid(ID) <scopes> "` + scope + `" .
		uid(ID) <type_name> "metricallocation" .
		uid(ID) <dgraph.type> "MetricAllocation" .
	`

	req := &api.Request{
		Query: q,
		Mutations: []*api.Mutation{
			{
				SetNquads: []byte(set),
			},
		},
		CommitNow: true,
	}
	_, err := r.dg.NewTxn().Do(ctx, req)
	if err != nil {
		logger.Log.Error("dgraph/upsertAllocateMetricInEquipmentHierarchy - failed to create Allocation Matric", zap.Error(err), zap.String("query", req.Query))
		return err
	}
	return err

}
func (r *EquipmentRepository) GetAllocatedMetricByEquipment(ctx context.Context, swidtag string, equipmentID string, scopes ...string) (*v1.AllocatedMetric, error) {

	q := ` {
		allocatedMetricList(func: eq(type_name,"metricallocation")) @filter(eq(scopes,` + scopes[0] + `) AND eq(product.swidtag, ` + swidtag + `) AND eq(equipment.id, ` + equipmentID + `)){
			uid
      		expand(_all_) 
		}
	}
	`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetAllocatedMetricByEquipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetAllocatedMetricByEquipment - cannot complete query transaction")
	}
	d := &v1.AllocatedMetrics{}

	if err = json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/GetAllocatedMetricByEquipment -", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetAllocatedMetricByEquipment - cannot unmarshal Json object")
	}

	if len(d.AllocatedMetricList) == 0 {
		return nil, nil
	}

	return d.AllocatedMetricList[0], err
}

// DeleteAllocateMetricInEquipment will delete allocation metric from equipment
func (r *EquipmentRepository) DeleteAllocateMetricInEquipment(ctx context.Context, uid string) error {
	query := `query {
		var(func: uid("` + uid + `")){
			ID as uid
		}
	}`

	delete := `
			uid(ID) * * .
	`
	set := `
			uid(ID) <Recycle> "true" .
	`
	muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
	logger.Log.Info(query)
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{muDelete},
		CommitNow: true,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DeleteAllocateMetricInEquipment - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf("DeleteAllocateMetricInEquipment - cannot complete query transaction")
	}
	return nil
}

func (r *EquipmentRepository) DeleteAllocationMetricInProduct(ctx context.Context, uid string, swidTag string, scope string) error {
	// Delete from product allocation
	queryP := `query {
		var(func: eq(product.swidtag,"` + swidTag + `")) @filter(eq(scopes,"` + scope + `") AND eq(type_name,"product")){
			ID as uid
		}
	}`

	deleteP := `
		uid(ID) <product.allocation> <` + uid + `> .
	`
	muDeleteP := &api.Mutation{DelNquads: []byte(deleteP)}
	reqP := &api.Request{
		Query:     queryP,
		Mutations: []*api.Mutation{muDeleteP},
		CommitNow: true,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.dg.NewTxn().Do(ctx, reqP); err != nil {
		logger.Log.Error("DeleteAllocateMetricInEquipment - ", zap.String("reason", err.Error()), zap.String("query", queryP))
		return fmt.Errorf("DeleteAllocateMetricInEquipment - cannot complete query transaction")
	}

	return nil
}

func (r *EquipmentRepository) GetMetricAlloc(ctx context.Context, equipmentId string, swidtag string, scope string) (*v1.MetricAllocated, error) {
	q := `{
		MetricAllocated(func: eq(type_name,"metricallocation")) @filter(eq(scopes,"` + scope + `") AND eq(equipment.id,"` + equipmentId + `") AND eq(product.swidtag,"` + swidtag + `")){
			AllocatedMetric: allocation.metric 
		}
	}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetMetricAlloc - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetMetricAlloc - cannot complete query transaction")
	}

	type data struct {
		MetricAllocated []*v1.MetricAllocated
	}
	d := &data{}

	if err = json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/GetMetricAlloc -", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetMetricAlloc - cannot unmarshal Json object")
	}

	if len(d.MetricAllocated) == 0 {
		logger.Log.Error("dgraph/GetMetricAlloc -", zap.String("reason", "Allocated metric not found"), zap.String("query", q))
		return nil, nil
	}

	return d.MetricAllocated[0], err
}

func (r *EquipmentRepository) GetEquipmentInfo(ctx context.Context, equipmentId string, scope string) (*v1.EquipInfo, error) {
	q := `{
		EquipInfo(func: eq(type_name,"equipment")) @filter(eq(scopes,"` + scope + `") AND eq(equipment.id,"` + equipmentId + `")){
			UID: uid
			Type: equipment.type 
		}
	}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetEquipmentInfo - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetEquipmentInfo - cannot complete query transaction")
	}
	fmt.Printf("resp= %s", resp)
	type data struct {
		EquipInfo []*v1.EquipInfo
	}
	d := &data{}

	if err = json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/GetEquipmentInfo -", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetEquipmentInfo - cannot unmarshal Json object")
	}
	fmt.Printf("d= %s", d)

	return d.EquipInfo[0], err
}

func (r *EquipmentRepository) GetAllocatedMetrics(ctx context.Context, swidtag string, scopes ...string) (*v1.AllocatedMetrics, error) {

	q := BuildQueryToGetAllocatedMetrics([]string{swidtag}, scopes)

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/getAllocatedMetrics - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/getAllocatedMetrics - cannot complete query transaction")
	}

	//fmt.Println(string(resp.Json))

	d := &v1.AllocatedMetrics{}

	if err = json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/getAllocatedMetrics -", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/getAllocatedMetrics - cannot unmarshal Json object")
	}

	return d, err
}

func BuildQueryToGetAllocatedMetrics(swidTag []string, scopes []string) string {

	q := ` {
		allocatedMetricList(func: eq(type_name,"metricallocation")) @filter(eq(scopes,$SCOPES) AND eq(product.swidtag, $SWIDTAG)){
			uid
      		expand(_all_) 
		}
	}
	`

	return replacer(q, map[string]string{
		"$SWIDTAG": strings.Join(swidTag, ","),
		"$SCOPES":  strings.Join(scopes, ","),
	})

}

func (r *EquipmentRepository) UpsertAllocateMetricInProduct(ctx context.Context, swidTag string, metUids []string, scope string) error {

	q := `query {
		var(func: eq(type_name,"product")) @filter(eq(scopes,[` + scope + `]) AND eq(product.swidtag, ` + swidTag + `)){
			ID as uid
		}
	}`
	var set string
	for _, metuid := range metUids {
		set += `
				uid(ID) <product.allocation> <` + metuid + `> .
		`
	}

	req := &api.Request{
		Query: q,
		Mutations: []*api.Mutation{

			{
				SetNquads: []byte(set),
			},
		},
		CommitNow: true,
	}

	_, err := r.dg.NewTxn().Do(ctx, req)
	if err != nil {
		logger.Log.Error("dgraph/upsertAllocateMetricInProduct - failed to create Allocation Matric", zap.Error(err), zap.String("query", req.Query))
		return err
	}
	return err

}

func (r *EquipmentRepository) GetEquipmentInfoByID(ctx context.Context, uid string) (*v1.EquipmentInfo, error) {
	q := ` {
		EquipmentInfo(func: uid(` + uid + `)){
			ID : uid
			EquipID : equipment.id
			Type : equipment.type
		}
	}
	`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/GetEquipmentInfoByID - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetEquipmentInfoByID - cannot complete query transaction")
	}
	type data struct {
		EquipmentInfo []*v1.EquipmentInfo
	}

	d := &data{}

	if err = json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("dgraph/GetEquipmentInfoByID -", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("dgraph/GetEquipmentInfoByID - cannot unmarshal Json object")
	}

	return d.EquipmentInfo[0], nil
}

func (r *EquipmentRepository) UpdateEquipmentUser(ctx context.Context, equipUserID string, scope string, equipmentUser int32) error {
	q := ` query{
		var(func: eq(users.id,` + equipUserID + `)) @filter(eq(scopes,` + scope + `) AND eq(type_name, "instance_users")){
			ID as uid
		}
	}
	`
	set := `
		uid(ID) <users.count> "` + strconv.FormatUint(uint64(equipmentUser), 10) + `" .
	`
	req := &api.Request{
		Query: q,
		Mutations: []*api.Mutation{
			{
				SetNquads: []byte(set),
			},
		},
		CommitNow: true,
	}
	_, err := r.dg.NewTxn().Do(ctx, req)
	if err != nil {
		logger.Log.Error("dgraph/UpdateEquipmentUser - failed to update equipment user", zap.Error(err), zap.String("query", req.Query))
		return err
	}

	return nil
}
