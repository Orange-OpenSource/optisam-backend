// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"sort"
	"strconv"

	"go.uber.org/zap"
)

// Equipments implements License Equipments function.
func (r *LicenseRepository) Equipments(ctx context.Context, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	q := `query Equips($tag:string,$pagesize:string,$offset:string) {
		  ID as var(func: eq(equipment.type,` + eqType.Type + `)) ` + agregateFilters(scopeFilters(scopes), equipFilter(eqType, params.Filter)) + `{}
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

// Equipment implements License Equipment function.
func (r *LicenseRepository) Equipment(ctx context.Context, eqType *v1.EquipmentType, id string, scopes []string) (json.RawMessage, error) {
	eqTypeFilter := `eq(equipment.type,` + eqType.Type + `)`
	q := `{
		  ID as var(func: uid(` + id + `))` + agregateFilters([]string{eqTypeFilter}, scopeFilters(scopes)) + `{}		
			Equipments(func: uid(ID)){
				 ` + equipQueryFieldsAll(eqType) + `
			}
	} `

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
func (r *LicenseRepository) EquipmentParents(ctx context.Context, eqType, parentEqType *v1.EquipmentType, id string, scopes []string) (int32, json.RawMessage, error) {
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
func (r *LicenseRepository) EquipmentChildren(ctx context.Context, eqType, childEqType *v1.EquipmentType, id string, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {
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

// EquipmentProducts implements License EquipmentProducts function.
func (r *LicenseRepository) EquipmentProducts(ctx context.Context, eqType *v1.EquipmentType, id string, params *v1.QueryEquipmentProduct, scopes []string) (int32, []*v1.EquipmentProduct, error) {
	eqTypeFilter := `eq(equipment.type,` + eqType.Type + `)`
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// TODO: log error
		sortOrder = sortASC
	}

	variables := make(map[string]string)
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))
	q := `
	query Products($tag:string,$pagesize:string,$offset:string) {
			ID as vars(func:uid(` + id + `))` + agregateFilters([]string{eqTypeFilter}, scopeFilters(scopes)) + `{
					~product.equipment ` + agregateFilters(equipmentProductFilter(params.Filter)) + `{
						productID as uid
					}
			}
			
			Exists(func:uid(ID)){
				TotalCount: count(uid)
			}
			
			NumOfRecords(func:uid(productID)){
				TotalCount:count(uid)
			}

			Products(func:uid(productID), ` + string(sortOrder) + `:` + equipmentProductFilterSortBy(params.SortBy) + `,first:$pagesize,offset:$offset){
				SwidTag: product.swidtag
				Name:    product.name
				Editor:  product.editor
				Version: product.version
			} 
		} 
		 `

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("EquipmentChildren - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("EquipmentChildren - cannot complete query transaction")
	}

	type Data struct {
		Exists       []*totalRecords
		NumOfRecords []*totalRecords
		Products     []*v1.EquipmentProduct
	}

	var prodList Data

	if err := json.Unmarshal(resp.GetJson(), &prodList); err != nil {
		logger.Log.Error("EquipmentChildren - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("EquipmentChildren - cannot unmarshal Json object")
	}

	if len(prodList.Exists) == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if prodList.Exists[0].TotalCount == 0 {
		return 0, nil, v1.ErrNodeNotFound
	}

	if len(prodList.NumOfRecords) == 0 {
		return 0, nil, v1.ErrNoData
	}

	if prodList.NumOfRecords[0].TotalCount == 0 {
		return 0, nil, v1.ErrNoData
	}

	return prodList.NumOfRecords[0].TotalCount, prodList.Products, nil
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
			filters = append(filters, fmt.Sprintf("(regexp(%v,/^%v/i))", prodPredSwidTag.String(), filter.Value()))
		case v1.EquipmentProductSearchKeyName:
			filters = append(filters, fmt.Sprintf("(regexp(%v,/^%v/i))", prodPredName.String(), filter.Value()))
		case v1.EquipmentProductSearchKeyEditor:
			filters = append(filters, fmt.Sprintf("(regexp(%v,/^%v/i))", prodPredName.String(), filter.Value()))
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
			dgFilters = append(dgFilters, fmt.Sprintf("(regexp(%v,/^%s/i))", pred, f.Value()))
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
			dgFilters = append(dgFilters, fmt.Sprintf("(regexp(%v,/^%v/i))", pred, f.Value()))
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
