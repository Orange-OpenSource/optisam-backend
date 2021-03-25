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
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

//EquipmentRepository for Dgraph
type EquipmentRepository struct {
	dg *dgo.Dgraph
	mu sync.Mutex
}

//NewEquipmentRepository creates new Repository
func NewEquipmentRepository(dg *dgo.Dgraph) *EquipmentRepository {
	return &EquipmentRepository{
		dg: dg,
	}
}

// Equipments implements License Equipments function.
func (r *EquipmentRepository) Equipments(ctx context.Context, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))
	uids := []string{}
	querySlice := []string{}
	if params.ApplicationFilter != nil && len(params.ApplicationFilter.Filters) != 0 {
		uids = append(uids, "ID_App")
		querySlice = append(querySlice, aggAppEquipsQueryFromID("ID_App", params.ApplicationFilter))
	}
	if params.ProductFilter != nil && len(params.ProductFilter.Filters) != 0 {
		uids = append(uids, "ID_Pro")
		querySlice = append(querySlice, aggProEquipsQueryFromID("ID_Pro", params.ProductFilter))
	}
	if params.InstanceFilter != nil && len(params.InstanceFilter.Filters) != 0 {
		uids = append(uids, "ID_Ins")
		querySlice = append(querySlice, aggInsEquipsQueryFromID("ID_Ins", params.InstanceFilter))
	}
	q := `query Equips($tag:string,$pagesize:string,$offset:string) {
			EquipID as var(func: eq(equipment.type,` + eqType.Type + `)) ` + agregateFilters(scopeFilters(scopes), equipFilter(eqType, params.Filter)) + `{}
			` + strings.Join(querySlice, "\n") + `
			NumOfRecords(func:uid(EquipID)) ` + uidAndFilter(uids) + `{
				TotalCount:count(uid)
		  	}		
			Equipments(func: uid(EquipID), ` + string(sortOrder) + `:` + equipSortBy(params.SortBy, eqType) + `,first:$pagesize,offset:$offset)` + uidAndFilter(uids) + `{
				` + equipQueryFields(eqType) + `
			}
	} `
	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("Equipments - cannot complete query transaction")
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

// DeleteEquipments implements License DeleteEquipments function.
func (r *EquipmentRepository) DeleteEquipments(ctx context.Context, scope string) error {
	query := `query {
		equipmentType as var(func: type(Equipment)) @filter(eq(scopes,` + scope + `)){
			equipments as equipment.id
		}
		`
	delete := `
			uid(equipmentType) * * .
			uid(equipments) * * .
	`
	set := `
			uid(equipmentType) <Recycle> "true" .
			uid(equipments) <Recycle> "true" .
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
	if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DeleteEquipments - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf("DeleteEquipments - cannot complete query transaction")
	}
	return nil
}

// Equipment implements License Equipment function.
func (r *EquipmentRepository) Equipment(ctx context.Context, eqType *v1.EquipmentType, id string, scopes []string) (json.RawMessage, error) {
	eqTypeFilter := `eq(equipment.type,"` + eqType.Type + `")`
	q := `{
		  ID as var (func:eq(equipment.id,"` + id + `")){}
		  EqID as var(func: uid(ID))` + agregateFilters([]string{eqTypeFilter}, scopeFilters(scopes)) + `{}		
			Equipments(func: uid(EqID)){
				 ` + equipQueryFieldsAll(eqType) + `
			}
	} `
	//fmt.Println(q)
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("Equipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("Equipment - cannot complete query transaction")
	}

	type Data struct {
		Equipments []json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("Equipment - cannot unmarshal Json object")
	}

	if len(equipList.Equipments) == 0 {
		return nil, v1.ErrNoData
	}

	return equipList.Equipments[0], nil
}

// EquipmentParents implements License EquipmentParents function.
func (r *EquipmentRepository) EquipmentParents(ctx context.Context, eqType, parentEqType *v1.EquipmentType, id string, scopes []string) (int32, json.RawMessage, error) {
	eqTypeFilter := `eq(equipment.type,` + eqType.Type + `)`
	q := `{
			ID as var(func:uid(` + id + `))` + agregateFilters([]string{eqTypeFilter}, scopeFilters(scopes)) + `{
				equipment.parent @filter(eq(equipment.type,` + parentEqType.Type + `)) {
					parentID as uid
				}
			}
  
		   	Exists(func:uid(ID)){
					TotalCount: count(uid)
			}
			
		    NumOfRecords(func:uid(parentID)){
					TotalCount:count(uid)
			}		
			Equipments(func: uid(parentID)){
				 ` + equipQueryFields(parentEqType) + `
			}
	} `

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return 0, nil, fmt.Errorf("EquipmentParents - cannot complete query transaction")
	}

	type Data struct {
		Exists       []*totalRecords
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return 0, nil, fmt.Errorf("EquipmentParents - cannot unmarshal Json object")
	}

	if len(equipList.Exists) == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if equipList.Exists[0].TotalCount == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if len(equipList.NumOfRecords) == 0 {
		return 0, nil, v1.ErrNoData
	}
	if equipList.NumOfRecords[0].TotalCount == 0 {
		return 0, nil, v1.ErrNoData
	}

	return equipList.NumOfRecords[0].TotalCount, equipList.Equipments, nil
}

// EquipmentChildren implements License EquipmentChildren function.
func (r *EquipmentRepository) EquipmentChildren(ctx context.Context, eqType, childEqType *v1.EquipmentType, id string, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
	eqTypeFilter := `eq(equipment.type,` + eqType.Type + `)`
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	q := `query Equips($tag:string,$pagesize:string,$offset:string) {
				ID as var(func:uid(` + id + `))` + agregateFilters([]string{eqTypeFilter}, scopeFilters(scopes)) + ` {
					~equipment.parent` + agregateFilters(equipFilterWithType(childEqType, params.Filter)) + `{
						childID as uid
					}
				}

				Exists(func:uid(ID)){
					TotalCount: count(uid)
				}
				
				NumOfRecords(func:uid(childID)){
					TotalCount:count(uid)
				}		
				Equipments(func: uid(childID), ` + string(sortOrder) + `:` + equipSortBy(params.SortBy, childEqType) + `,first:$pagesize,offset:$offset){
					` + equipQueryFields(childEqType) + `
				}
    } `

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("Equipments - cannot complete query transaction")
	}

	type Data struct {
		Exists       []*totalRecords
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("Equipments - cannot unmarshal Json object")
	}

	if len(equipList.Exists) == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if equipList.Exists[0].TotalCount == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if len(equipList.NumOfRecords) == 0 {
		return 0, nil, v1.ErrNoData
	}

	if equipList.NumOfRecords[0].TotalCount == 0 {
		return 0, nil, v1.ErrNoData
	}

	return equipList.NumOfRecords[0].TotalCount, equipList.Equipments, nil
}

// ProductEquipments implements Licence ProductEquipments function
func (r *EquipmentRepository) ProductEquipments(ctx context.Context, swidTag string, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {

	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	q := `query Equips($tag:string,$pagesize:string,$offset:string) {
		  var(func: eq(product.swidtag,` + swidTag + `))` + agregateFilters(scopeFilters(scopes)) + `{
		  IID as product.equipment @filter(eq(equipment.type,` + eqType.Type + `))  {} }
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
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("Equipments - cannot complete query transaction")
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

//UpsertEquipment ...
func (r *EquipmentRepository) UpsertEquipment(ctx context.Context, scope string, eqType string, parentEqType string, eqData interface{}) error {
	v := reflect.ValueOf(eqData).Elem()
	var set string
	var ID, parentID string
	//Iterate over struct fields dynamically
	for i := 0; i < v.NumField(); i++ {
		//For Identifier
		switch v.Type().Field(i).Tag.Get("dbname") {
		case "equipment.id":
			ID = v.Field(i).String()
		case "equipment.parent":
			parentID = v.Field(i).String()
			continue
		}
		// switch v.Field(i).Interface(){
		// 	case int
		// }
		// fmt.Println()
		// fmt.Printf("FieldName:%v,FieldValue:%v,FieldTag:%v", v.Type().Field(i).Name, v.Field(i).Interface(), v.Type().Field(i).Tag.Get("dbname"))
		//type converison
		var val string
		switch v.Field(i).Kind() {
		case reflect.String:
			val = v.Field(i).String()
		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			val = strconv.FormatInt(v.Field(i).Int(), 10)
		case reflect.Float32, reflect.Float64:
			val = strconv.FormatFloat(v.Field(i).Float(), 'f', -1, 64)
		}
		set += `
		uid(equipment) <` + v.Type().Field(i).Tag.Get("dbname") + `> "` + val + `" .
		`
	}
	var mutations []*api.Mutation
	queries := []string{"query{"}
	query := `var(func: eq(equipment.id,"` + ID + `")) @filter(eq(type_name,"equipment") AND eq(scopes,"` + scope + `")){
			equipment as uid
		}`
	mutations = append(mutations, &api.Mutation{
		SetNquads: []byte(set),
	})
	log.Println("m1 ", string(mutations[0].SetNquads))
	mutations = append(mutations, &api.Mutation{
		Cond: "@if(eq(len(equipment),0))",
		SetNquads: []byte(`
		uid(equipment) <scopes> "` + scope + `" .
		uid(equipment) <equipment.id> "` + ID + `" .
		uid(equipment) <type_name> "equipment" .
		uid(equipment) <dgraph.type> "Equipment" .
		uid(equipment) <equipment.type> "` + eqType + `" .
		`),
	})
	log.Println("m2 ", string(mutations[1].SetNquads))
	//SCOPE BASED CHANGE
	var mut1, mut2 api.Mutation
	if parentID != "" {
		query += `
		var(func: eq(equipment.id,"` + parentID + `")) @filter(eq(type_name,"equipment") AND eq(scopes,"` + scope + `")){
			parent as uid
		}`

		mut1.Cond = "@if(eq(len(parent),0))"
		mut1.SetNquads = []byte(`
			uid(parent) <scopes> "` + scope + `" .
			uid(parent) <equipment.id> "` + parentID + `" .
			uid(parent) <equipment.type>  "` + parentEqType + `" .
			uid(parent) <type_name> "equipment" .
			uid(parent) <dgraph.type> "Equipment" .
			`)

		mut2.SetNquads = []byte(`
			uid(equipment) <equipment.parent>  uid(parent) .
			`)
		log.Println("m3 ", string(mut1.SetNquads))
		log.Println("m4 ", string(mut2.SetNquads))
	}

	mutations = append(mutations, &mut1)
	mutations = append(mutations, &mut2)

	queries = append(queries, query)
	queries = append(queries, "}")

	req := &api.Request{
		Query:     strings.Join(queries, "\n"),
		Mutations: mutations,
		CommitNow: true,
	}
	//Handling locking mechanism on Uspert txn as dgrpah doesnt provide it
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query))
		return errors.New("DBError")
	}
	return nil
}

func equipmentProductFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, filter := range filter.Filters {
		switch v1.EquipmentProductSearchKey(filter.Key()) {
		case v1.EquipmentProductSearchKeySwidTag:
			filters = append(filters, stringFilter(prodPredSwidTag.String(), filter))
		case v1.EquipmentProductSearchKeyName:
			filters = append(filters, stringFilter(prodPredName.String(), filter))
		case v1.EquipmentProductSearchKeyEditor:
			filters = append(filters, stringFilter(prodPredEditor.String(), filter))
		default:
			logger.Log.Error("equipmentProductFilter - unknown filter key", zap.String("filterKey", filter.Key()))
		}
	}
	return filters
}

func equipmentProductFilterSortBy(sortBy v1.EquipmentProductSortBy) string {
	switch sortBy {
	case v1.EquipmentProductSortBySwidTag:
		return prodPredSwidTag.String()
	case v1.EquipmentProductSortByName:
		return prodPredName.String()
	case v1.EquipmentProductSortByEditor:
		return prodPredEditor.String()
	case v1.EquipmentProductSortByVersion:
		return prodPredVersion.String()
	default:
		logger.Log.Error("equipmentProductFilterSortBy - unknown sortby field taking swidtag as sort by", zap.Uint8("sortBy", uint8(sortBy)))
		return prodPredSwidTag.String()
	}
}

func equipSortBy(name string, eqType *v1.EquipmentType) string {
	for _, attr := range eqType.Attributes {
		if attr.Name == name {
			if attr.IsIdentifier {
				return "equipment.id"
			}
			if !attr.IsDisplayed {
				// atribute is not displayed we cannot sort on this sort by id instead
				logger.Log.Error("equuipSortBy - invalid sort attribute attribute is not displayed", zap.String("attr_name", name))
				return "equipment.id"
			}

			if attr.IsParentIdentifier {
				// atribute is not displayed we cannot sort on this sort by id instead
				logger.Log.Error("equuipSortBy - invalid sort attribute attribute - parent identifier", zap.String("attr_name", name))
				return "equipment.id"
			}

			// TODO check if we need to get the parent_id
			// we can have that in equipment details

			return "equipment." + eqType.Type + "." + name
		}
	}
	logger.Log.Error("equipSoryBy - cannot find equip attribute sorting by identifier", zap.String("attribute_name", name))
	return "equipment.id"
}

func equipQueryFields(eqType *v1.EquipmentType) string {
	query := ""
	//query := ""
	eqName := "equipment." + eqType.Type + "."
	for _, attr := range eqType.Attributes {
		if !attr.IsDisplayed {
			continue
		}

		if attr.IsParentIdentifier {
			continue
		}

		if attr.IsIdentifier {
			query = attr.Name + " : equipment.id \n" + query
			continue
		}
		query += attr.Name + " : " + eqName + attr.Name + "\n"
	}
	return "ID: uid\n" + query
}

func equipQueryFieldsAll(eqType *v1.EquipmentType) string {
	query := ""
	//query := ""
	eqName := "equipment." + eqType.Type + "."
	for _, attr := range eqType.Attributes {

		if attr.IsParentIdentifier {
			continue
		}

		if attr.IsIdentifier {
			query = attr.Name + " : equipment.id \n" + query
			continue
		}
		query += attr.Name + " : " + eqName + attr.Name + "\n"
	}
	return "ID: uid\n" + query
}

func equipFilter(eqType *v1.EquipmentType, filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	equipName := "equipment." + eqType.Type + "."
	sort.Sort(filter)
	dgFilters := []string{}
	for _, f := range filter.Filters {
		i := attributeIndexByName(f.Key(), eqType.Attributes)
		if i == -1 {
			logger.Log.Error("dgraph - equipFilter - ", zap.String("reason", "attribute "+f.Key()+" not found in attributes"))
			continue
		}
		switch eqType.Attributes[i].Type {
		case v1.DataTypeString:
			pred := equipName + f.Key()
			if eqType.Attributes[i].IsIdentifier {
				pred = "equipment.id"
			}
			dgFilters = append(dgFilters, stringFilter(pred, f))
		case v1.DataTypeInt, v1.DataTypeFloat:
			pred := equipName + f.Key()
			dgFilters = append(dgFilters, fmt.Sprintf("(eq(%v,%v))", pred, f.Value()))
		default:
			logger.Log.Error("dgraph - equipFilter - datatype is not suppoted ",
				zap.String("dataType", eqType.Attributes[i].Type.String()), zap.String("predicate", f.Key()))
		}
	}

	if len(dgFilters) == 0 {
		return nil
	}
	return dgFilters
}

// TODO: we need refactoring here
func equipFilterWithType(eqType *v1.EquipmentType, filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	equipName := "equipment." + eqType.Type + "."
	sort.Sort(filter)
	dgFilters := []string{
		`(eq(equipment.type,` + eqType.Type + `))`,
	}
	for idx, f := range filter.Filters {
		i := attributeIndexByName(f.Key(), eqType.Attributes)
		if i == -1 {
			logger.Log.Error("dgraph - equipFilter - ", zap.String("reason", "attribute "+f.Key()+" not found in attributes"))
			continue
		}
		if !eqType.Attributes[i].IsSearchable {
			logger.Log.Error("dgraph - equipFilter - attribute is not searchable", zap.String("attr_name", eqType.Attributes[i].Name), zap.String("attr_mapping", eqType.Attributes[i].MappedTo))
			continue
		}

		if !eqType.Attributes[i].IsDisplayed {
			logger.Log.Error("dgraph - equipFilter - attribute is not searchable as it is not displayable", zap.String("attr_name", eqType.Attributes[i].Name), zap.String("attr_mapping", eqType.Attributes[i].MappedTo))
			continue
		}

		switch eqType.Attributes[i].Type {
		case v1.DataTypeString:
			pred := equipName + f.Key()
			if eqType.Attributes[i].IsIdentifier {
				pred = "equipment.id"
			}
			dgFilters = append(dgFilters, stringFilter(pred, f))
		case v1.DataTypeInt, v1.DataTypeFloat:
			pred := equipName + f.Key()
			dgFilters = append(dgFilters, fmt.Sprintf("(ge(%v,%v))", pred, f.Value()))
		default:
			logger.Log.Error("dgraph - equipFilter - datatype is not suppoted ",
				zap.String("dataType", eqType.Attributes[idx].Type.String()), zap.String("predicate", f.Key()))
		}
	}

	if len(dgFilters) == 0 {
		return nil
	}
	return dgFilters
}

func attributeIndexByName(name string, attrs []*v1.Attribute) int {
	for i := range attrs {
		if attrs[i].Name == name {
			return i
		}
	}
	return -1
}

func uidAndFilter(uids []string) string {
	if len(uids) == 0 {
		return ""
	}
	filters := make([]string, len(uids))
	for i := range uids {
		filters[i] = " uid( " + uids[i] + ")"
	}
	return "@filter( " + strings.Join(filters, " AND ") + " )"
}

func aggAppEquipsQueryFromID(id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return ` var(func: eq(application.id,` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `)) @cascade{
		application.instance{
		` + id + ` as instance.equipment
		}
	  }`
}

func aggProEquipsQueryFromID(id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return ` var(func: eq(product.swidtag,` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `)) @cascade{
		` + id + ` as product.equipment
	  }`
}

func aggInsEquipsQueryFromID(id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return ` var(func: eq(instance.id,` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `)) @cascade{
		` + id + ` as instance.equipment
	  }`
}
