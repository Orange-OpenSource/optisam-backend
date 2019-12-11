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
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type prodPred string

func (p prodPred) String() string {
	return string(p)
}

const (
	prodPredName     prodPred = "product.name"
	prodPredSwidTag  prodPred = "product.swidtag"
	prodPredVersion  prodPred = "product.version"
	prodPredEditor   prodPred = "product.editor"
	prodPredNumOfApp prodPred = "val(numOfApplications)"
	prodPredNumOfEqp prodPred = "val(numOfEquipments)"
)

const (
	offset   string = "$offset"
	pagesize string = "$pagesize"
)

//GetProducts is the implementation
func (r *LicenseRepository) GetProducts(ctx context.Context, params *v1.QueryProducts, scopes []string) (*v1.ProductInfo, error) {

	variables := make(map[string]string)
	variables["$offset"] = strconv.Itoa(int(params.Offset))
	variables["$pagesize"] = strconv.Itoa(int(params.PageSize))

	sortBy, err := keyToPredForProduct(params.SortBy)
	if err != nil {
		sortBy = prodPredName
	}

	uids := []string{}
	aggQuery := ""
	if params.AggFilter != nil && len(params.AcqFilter.Filters) != 0 {
		uids = append(uids, "ID_AGG")
		aggQuery = aggQueryFromFilterWithID("ID", "ID_AGG", params.AggFilter)
	}

	q := `query OrderProducts($pagesize:string,$offset:string) {

		ID as var(func: eq(type, "product"))@cascade` + agregateFilters(scopeFilters(scopes), productFilter(params.Filter)) + `{
			numOfApplications as count(~application.product)
			numOfEquipments as count(product.equipment)
			` + acqFilter(params.AcqFilter) + `
		}

		` + aggQuery + `

		NumOfRecords(func:uid(ID))` + uidNotFilter(uids) + `{
			TotalCnt : count(uid)
		}      
		

		Products(func:uid(ID), ` + params.SortOrder + `:` + string(sortBy) + `,first:$pagesize,offset:$offset) ` + uidNotFilter(uids) + `{
		   Name :              product.name
		   Version  :          product.version
		   Category :          product.category
           Editor :            product.editor
		   Swidtag :           product.swidtag
		   NumOfEquipments :   val(numOfEquipments)
           NumOfApplications : val(numOfApplications)	
		  }
		}`

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetProducts - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProducts - cannot complete query transaction")
	}

	var ProdList v1.ProductInfo

	if err := json.Unmarshal(resp.GetJson(), &ProdList); err != nil {
		logger.Log.Error("GetProducts - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProducts - cannot unmarshal Json object")
	}
	return &ProdList, nil
}

// GetProductInformation ...
func (r *LicenseRepository) GetProductInformation(ctx context.Context, swidtag string, scopes []string) (*v1.ProductAdditionalInfo, error) {

	variables := make(map[string]string)
	variables["$tag"] = swidtag

	q := `query OrderProducts($tag:string) {
		Products(func: eq(product.swidtag, $tag))` + agregateFilters(scopeFilters(scopes)) + ` {    
			Swidtag : 		   product.swidtag
			Name :    		   product.name
			Version : 		   product.version
			Editor :  		   product.editor
			NumofEquipments:   count(product.equipment)
	    NumOfApplications: count(~application.product)
	    NumofOptions:      count(product.child)
			Child:            product.child{
															SwidTag :	product.swidtag
															Name :	    product.name
															Editor :    product.editor
															Version :   product.version
					                }
		}

	}`

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetProductInformation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProductInformation - cannot complete query transaction")
	}

	var ProductDetails v1.ProductAdditionalInfo

	if err := json.Unmarshal(resp.GetJson(), &ProductDetails); err != nil {
		logger.Log.Error("GetProductInformation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProductInformation - cannot unmarshal Json object")
	}
	return &ProductDetails, nil
}

// GetApplicationsForProduct ...
func (r *LicenseRepository) GetApplicationsForProduct(ctx context.Context, params *v1.QueryApplicationsForProduct, scopes []string) (*v1.ApplicationsForProduct, error) {

	variables := make(map[string]string)

	variables["$tag"] = params.SwidTag
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	sortBy, err := keyToPredForGetApplicationsForProduct(params.SortBy)
	if err != nil {
		logger.Log.Error("GetApplicationsForProduct - ", zap.String("reason", err.Error()))
		// By default we are using product name for sorting
		sortBy = appPredName
	}

	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		sortOrder = sortASC
	}

	q := `query ApplicationsForProduct($tag:string,$pagesize:string,$offset:string) {
		var(func: eq(product.swidtag, $tag))` + agregateFilters(scopeFilters(scopes)) + `{
			ID as ~application.product ` + agregateFilters(applicationFilterForGetApplicationsForProduct(params.Filter)) + ` {
				numOfInstances as count(application.instance) 
					 application.instance{
				  	 eqp as count(instance.equipment)
				   }
				   numofEquipments as sum(val(eqp)) 	
	     	  }
	  	}
		  NumOfRecords(func:uid(ID)){
		    TotalCnt:count(uid)
			}
					
			Applications(func: uid(ID), ` + string(sortOrder) + `:` + string(sortBy) + `,first:$pagesize,offset:$offset){
				 ApplicationID:   application.id
				 Name:            application.name
				 Owner:	          application.owner
			     NumOfEquipments: val(numofEquipments) 
				 NumOfInstances:  val(numOfInstances)
			}
	} `

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetApplicationsForProduct - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplicationsForProduct - cannot complete query transaction")
	}

	var appList v1.ApplicationsForProduct

	if err := json.Unmarshal(resp.GetJson(), &appList); err != nil {
		logger.Log.Error("GetApplicationsForProduct - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplicationsForProduct - cannot unmarshal Json object")
	}
	return &appList, nil

}

// GetInstancesForApplicationsProduct implements Licence GetInstancesForApplicationsProduct function
func (r *LicenseRepository) GetInstancesForApplicationsProduct(ctx context.Context, params *v1.QueryInstancesForApplicationProduct, scopes []string) (*v1.InstancesForApplicationProduct, error) {

	variables := make(map[string]string)
	variables["$swidTag"] = params.SwidTag
	variables["$appId"] = params.AppID
	variables[offset] = strconv.Itoa(int(params.Offset))
	variables[pagesize] = strconv.Itoa(int(params.PageSize))

	sortBy, err := keyToPredForGetInstancesForApplicationsProduct(params.SortBy)
	if err != nil {
		sortBy = insPredName
	}

	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		sortOrder = sortASC
	}

	q := `query InstancesForApplicationProduct($swidTag:string, $appId:string ,$pagesize:string, $offset:string){
	       	 var(func:eq(product.swidtag, $swidTag))` + agregateFilters(scopeFilters(scopes)) + `{
							~application.product @filter(eq(application.id, $appId)) {
							  ID as application.instance{
										numOfEquipments as count(instance.equipment)
										numOfProducts as count(instance.product) 
									}
								}  
					  } 
																								 
				 NumOfRecords(func:uid(ID)){
						TotalCnt:count(uid)
					}
																								 

		     Instances(func:uid(ID), ` + string(sortOrder) + `:` + string(sortBy) + `,first:$pagesize,offset:$offset){
						ID:              instance.id
						Environment:     instance.environment
						NumOfEquipments: val(numOfEquipments)
						NumOfProducts:   val(numOfProducts)  
		      }
	 
 }`

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetInstancesForApplicationsProduct - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetInstancesForApplicationsProduct - cannot complete query transaction")
	}

	var instanceList v1.InstancesForApplicationProduct

	if err := json.Unmarshal(resp.GetJson(), &instanceList); err != nil {
		logger.Log.Error("GetInstancesForApplicationsProduct - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetInstancesForApplicationsProduct - cannot unmarshal Json object")
	}
	return &instanceList, nil

}

// ProductAcquiredRights implements Licence ProductAcquiredRights function
func (r *LicenseRepository) ProductAcquiredRights(ctx context.Context, swidTag string, scopes []string) (string, []*v1.ProductAcquiredRight, error) {
	q := `
	{
		Products(func: eq(product.swidtag,` + swidTag + `))` + agregateFilters(scopeFilters(scopes)) + `{
		  ID: uid
		  AcquiredRights: product.acqRights{
		  SKU: acqRights.SKU
		  Metric: acqRights.metric
		  AcqLicenses: acqRights.numOfAcqLicences
		  TotalCost: acqRights.totalCost
		  AvgUnitPrice: acqRights.averageUnitPrice
		}
		}
	  }
	`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ProductAcquiredRights - query failed", zap.Error(err), zap.String("query", q))
		return "", nil, errors.New("dgraph/ProductAcquiredRights -  failed to fetch acquired rights")
	}

	type product struct {
		ID             string
		AcquiredRights []*v1.ProductAcquiredRight
	}

	type products struct {
		Products []*product
	}

	data := &products{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/ProductAcquiredRights - unmarshal failed", zap.Error(err), zap.String("query", q), zap.String("response", string(resp.Json)))
		return "", nil, errors.New("dgraph/ProductAcquiredRights -  failed unmarshal response")
	}

	if len(data.Products) == 0 {
		return "", nil, v1.ErrNodeNotFound
	}

	return data.Products[0].ID, data.Products[0].AcquiredRights, nil
}

// ProductEquipments implements Licence ProductEquipments function
func (r *LicenseRepository) ProductEquipments(ctx context.Context, swidTag string, eqType *v1.EquipmentType, params *v1.QueryEquipments, scopes []string) (int32, json.RawMessage, error) {

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

func productFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := keyToPredForProduct(f.Key())
		if err != nil {
			logger.Log.Error("dgraph - productFilter - ", zap.String("reason", err.Error()))
			continue
		}

		switch pred {
		case prodPredSwidTag, prodPredName, prodPredEditor:
			filters = append(filters, fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, f.Value()))
		}
	}
	return filters
}

func keyToPredForProduct(key string) (prodPred, error) {
	switch key {
	case "name":
		return prodPredName, nil
	case "swidtag":
		return prodPredSwidTag, nil
	case "version":
		return prodPredVersion, nil
	case "editor":
		return prodPredEditor, nil
	case "numOfApplications":
		return prodPredNumOfApp, nil
	case "numofEquipments":
		return prodPredNumOfEqp, nil
	default:
		return "", fmt.Errorf("keyToPredForProduct - cannot find dgraph predicate for key: %s", key)
	}
}

func acqFilter(filter *v1.AggregateFilter) string {
	filters := acquiredRightsFilter(filter)
	if len(filters) == 0 {
		return ""
	}
	return `product.acqRights @filter(` + strings.Join(filters, ",") + `)`
}

func aggFilter(filter *v1.AggregateFilter) string {
	if filter == nil || len(filter.Filters) == 0 {
		return ""
	}
	return ` @filter(eq(metric.name,` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `))`
}
