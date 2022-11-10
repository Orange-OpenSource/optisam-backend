package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

const EQUIPMENTID = "equipment.id"

// EquipmentRepository for Dgraph
type EquipmentRepository struct {
	dg *dgo.Dgraph
	mu sync.Mutex
}

// NewEquipmentRepository creates new Repository
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
		return 0, nil, fmt.Errorf("equipments - cannot complete query transaction")
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

// DropMetaData deletes metadata
func (r *EquipmentRepository) DropMetaData(ctx context.Context, scope string) error {
	query := `query {
		var(func: type(Metadata)) @filter(eq(scopes,` + scope + `)){
			metadataId as  uid
		}
		var(func: type(MetadataEquipment)) @filter(eq(scopes,` + scope + `)){
			eqTypeId as  uid
		}

		`
	delete := `
			uid(metadataId) * * .
			uid(eqTypeId) * * .
	`
	set := `
			uid(metadataId) <Recycle> "true" .
			uid(eqTypeId) <Recycle> "true" .
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
		logger.Log.Error("DropMetaData - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf(" dropMetaData - cannot complete query transaction")
	}
	return nil
}

// DeleteEquipments implements License DeleteEquipments function.
func (r *EquipmentRepository) DeleteEquipments(ctx context.Context, scope string) error {
	batchsize, err := strconv.Atoi(os.Getenv("DEL_BATCH_SIZE"))
	if batchsize == 0 || err != nil {
		batchsize = 25000
	}
	query := `query {
		var(func: type(Equipment), first: ` + strconv.Itoa(batchsize) + `) @filter(eq(scopes,` + scope + `)){
			equipments as uid
		}
	}`
	delete := `
			uid(equipments) * * .
	`
	set := `
			uid(equipments) <Recycle> "true" .
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
	totalEquipments, err := r.getEquipmentCount(ctx, scope)
	if err != nil {
		if errors.Is(err, v1.ErrNoData) {
			logger.Log.Info("deleteEquipments - No equipment")
			return nil
		}
		return fmt.Errorf("deleteEquipments - getEquipmentCount - can not delete equipments")
	}
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
	defer cancel()
	for i := 0; i < int(math.Ceil((float64(totalEquipments) / float64(batchsize)))); i++ {
		retryCount := 0
	retry:
		if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
			if err != dgo.ErrAborted {
				logger.Log.Error("deleteEquipments - ", zap.String("reason", err.Error()), zap.String("query", query))
				return fmt.Errorf("deleteEquipments - cannot complete query transaction")
			}
			time.Sleep(1 * time.Second)
			logger.Log.Info("deleteEquipments - Tansaction aborted error - batch retry", zap.Int("Batch number:", i+1))
			if retryCount < 2 {
				retryCount++
				goto retry
			} else {
				logger.Log.Info("deleteEquipments - Tansaction aborted error - batch failure", zap.Int("Batch number:", i+1))
				break
			}
		}
		logger.Log.Info("deleteEquipments - batch completed", zap.Int("Batch number:", i+1))
		time.Sleep(1 * time.Millisecond)
	}
	return nil
}

func (r *EquipmentRepository) getEquipmentCount(ctx context.Context, scope string) (int32, error) {
	q := `query {
		NumOfRecords(func: type(Equipment)) @filter(eq(scopes,` + scope + `)){
			TotalCount:count(uid)
		}	
	} `
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("getEquipmentCount - ", zap.String("reason", err.Error()), zap.String("query", q))
		return 0, fmt.Errorf("getEquipmentCount - cannot complete query transaction")
	}
	type Data struct {
		NumOfRecords []*totalRecords
	}
	var equipList Data
	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("getEquipmentCount - ", zap.String("reason", err.Error()), zap.String("query", q))
		return 0, fmt.Errorf("getEquipmentCount - cannot unmarshal Json object")
	}

	if len(equipList.NumOfRecords) == 0 {
		return 0, v1.ErrNoData
	}
	return equipList.NumOfRecords[0].TotalCount, nil
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
	// fmt.Println(q)
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("Equipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipment - cannot complete query transaction")
	}

	type Data struct {
		Equipments []json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipment - cannot unmarshal Json object")
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
		return 0, nil, fmt.Errorf("equipmentParents - cannot complete query transaction")
	}

	type Data struct {
		Exists       []*totalRecords
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return 0, nil, fmt.Errorf("equipmentParents - cannot unmarshal Json object")
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
		return 0, nil, fmt.Errorf("equipments - cannot complete query transaction")
	}

	type Data struct {
		Exists       []*totalRecords
		NumOfRecords []*totalRecords
		Equipments   json.RawMessage
	}

	var equipList Data

	if err := json.Unmarshal(resp.GetJson(), &equipList); err != nil {
		logger.Log.Error("Equipments - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("equipments - cannot unmarshal Json object")
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
		  var(func: eq(product.swidtag,"` + swidTag + `"))` + agregateFilters(scopeFilters(scopes), typeFilters("type_name", "product")) + `{
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
		return 0, nil, fmt.Errorf("equipments - cannot complete query transaction")
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

// UpsertEquipment ...
func (r *EquipmentRepository) UpsertEquipment(ctx context.Context, scope string, eqType string, parentEqType string, eqData interface{}) error {
	v := reflect.ValueOf(eqData).Elem()
	var set string
	var ID, parentID string
	// Iterate over struct fields dynamically
	for i := 0; i < v.NumField(); i++ {
		// For Identifier
		switch v.Type().Field(i).Tag.Get("dbname") {
		case EQUIPMENTID:
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
		// type conversion
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

	mutations = append(mutations, &api.Mutation{
		// We do not have upsert for now, it will either be insert or delete
		//Cond: "@if(eq(len(equipment),0))",
		SetNquads: []byte(`
		uid(equipment) <scopes> "` + scope + `" .
		uid(equipment) <equipment.id> "` + ID + `" .
		uid(equipment) <type_name> "equipment" .
		uid(equipment) <dgraph.type> "Equipment" .
		uid(equipment) <equipment.type> "` + eqType + `" .
		`),
	})

	// SCOPE BASED CHANGE
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
	logger.Log.Info("EquipmentService - dgraph/UpsertEquipment", zap.Any("api.request", req))
	//logger.Log.Sugar().Info("Equipment Services", "appendSet", set)
	// Handling locking mechanism on Uspert txn as dgrpah doesnt provide it
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query))
		return errors.New("DBError")
	}
	return nil
}

func equipSortBy(name string, eqType *v1.EquipmentType) string {
	for _, attr := range eqType.Attributes {
		if attr.Name == name {
			if attr.IsIdentifier {
				return EQUIPMENTID
			}
			if !attr.IsDisplayed {
				// attribute is not displayed we cannot sort on this sort by id instead
				logger.Log.Error("equuipSortBy - invalid sort attribute attribute is not displayed", zap.String("attr_name", name))
				return EQUIPMENTID
			}

			if attr.IsParentIdentifier {
				// attribute is not displayed we cannot sort on this sort by id instead
				logger.Log.Error("equuipSortBy - invalid sort attribute attribute - parent identifier", zap.String("attr_name", name))
				return EQUIPMENTID
			}

			// TODO check if we need to get the parent_id
			// we can have that in equipment details

			return "equipment." + eqType.Type + "." + name
		}
	}
	logger.Log.Error("equipSoryBy - cannot find equip attribute sorting by identifier", zap.String("attribute_name", name))
	return EQUIPMENTID
}

func equipQueryFields(eqType *v1.EquipmentType) string {
	query := ""
	// query := ""
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
	// query := ""
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
				pred = EQUIPMENTID
			}
			dgFilters = append(dgFilters, stringFilter(pred, f))
		case v1.DataTypeInt, v1.DataTypeFloat:
			pred := equipName + f.Key()
			dgFilters = append(dgFilters, fmt.Sprintf("(eq(%v,%v))", pred, f.Value()))
		default:
			logger.Log.Error("dgraph - equipFilter - datatype is not supported ",
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
				pred = EQUIPMENTID
			}
			dgFilters = append(dgFilters, stringFilter(pred, f))
		case v1.DataTypeInt, v1.DataTypeFloat:
			pred := equipName + f.Key()
			dgFilters = append(dgFilters, fmt.Sprintf("(ge(%v,%v))", pred, f.Value()))
		default:
			logger.Log.Error("dgraph - equipFilter - datatype is not supported ",
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
	return ` var(func: eq(application.id,"` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `")) @cascade{
		application.instance{
		` + id + ` as instance.equipment
		}
	  }`
}

func aggProEquipsQueryFromID(id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return ` var(func: eq(product.swidtag,"` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `")) @cascade{
		` + id + ` as product.equipment
	  }`
}

func aggInsEquipsQueryFromID(id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return ` var(func: eq(instance.id,"` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `")) @cascade{
		` + id + ` as instance.equipment
	  }`
}
