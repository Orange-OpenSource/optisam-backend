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

type appPred string

const (
	// For application request
	appPredName      appPred = "application.name"
	appPredID        appPred = "application.id"
	appPredOwner     appPred = "application.owner"
	appPredNumOfProd appPred = "val(numOfProducts)"
	appPredNumOfEqp  appPred = "val(numofEquipments)"

	// for product application only
	appPredNumOfIns appPred = "val(numOfInstances)"
)

//GetApplications ...
func (r *LicenseRepository) GetApplications(ctx context.Context, params *v1.QueryApplications, scopes []string) (*v1.ApplicationInfo, error) {

	variables := make(map[string]string)
	variables["$offset"] = strconv.Itoa(int(params.Offset))
	variables["$pagenum"] = strconv.Itoa(int(params.PageSize))
	sortBy, err := keyToPredForApplication(params.SortBy)
	if err != nil {
		sortBy = appPredName
	}

	q := `query OrderApplications($pagenum:int,$offset:int) {

		ID as var(func: eq(type, "application"))` + agregateFilters(scopeFilters(scopes), applicationFilter(params.Filter)) + `  { 
			numOfInstances as count(application.instance)
			numOfProducts as count(application.product)
		}

		NumOfRecords(func:uid(ID)){
			TotalCnt:count(uid)
		}  

		Applications(func:uid(ID), ` + params.SortOrder + `:` + string(sortBy) + `,first:$pagenum,offset:$offset){
			Name :             application.name
			ApplicationID :    application.id
			ApplicationOwner : application.owner
			NumOfInstances :   val(numOfInstances)
			NumOfProducts :    val(numOfProducts)		
		}
	
	}`
	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)

	if err != nil {
		logger.Log.Error("GetApplications - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplications - cannot complete query transaction")
	}

	var AppList v1.ApplicationInfo

	if err = json.Unmarshal(resp.GetJson(), &AppList); err != nil {
		logger.Log.Error("GetApplications - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplications - cannot unmarshal Json object")
	}
	return &AppList, nil
}

// GetApplication ...
func (r *LicenseRepository) GetApplication(ctx context.Context, appID string, scopes []string) (*v1.ApplicationDetails, error) {
	variables := make(map[string]string)
	variables["$id"] = appID

	q := `query ApplicationByID($id:string) {

	    Application(func:eq(application.id,$id))` + agregateFilters(scopeFilters(scopes)) + `{
		Name :             application.name
		ApplicationID :    application.id
		ApplicationOwner : application.owner
		NumOfInstances :   count(application.instance)
		NumOfProducts :    count(application.product)		
	   }
	 }  	
	`
	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)

	if err != nil {
		logger.Log.Error("GetApplication - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplication - cannot complete query transaction")
	}

	var application struct {
		Application []*v1.ApplicationDetails
	}

	if err = json.Unmarshal(resp.GetJson(), &application); err != nil {
		logger.Log.Error("GetApplication - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetApplication - cannot unmarshal Json object")
	}
	if len(application.Application) == 0 {
		return nil, v1.ErrNoData
	}
	return application.Application[0], nil
}

//GetProductsForApplication ...
func (r *LicenseRepository) GetProductsForApplication(ctx context.Context, id string, scopes []string) (*v1.ProductsForApplication, error) {

	variables := make(map[string]string)
	variables["$id"] = id
	q := `query ProductsforApplication($id:string) {	
			    var(func:eq(application.id,$id))` + agregateFilters(scopeFilters(scopes)) + `{
					prodID as application.product
					application.instance {
					ins_equip as  instance.equipment
				 }
		    }

				NumOfRecords(func:uid(prodID)){
					TotalCnt:count(uid)
				}

				Products(func:uid(prodID)){
					SwidTag:	     product.swidtag
					Name:            product.name
					Version:	     product.version
					Editor:          product.editor
					NumOfEquipments: count(product.equipment)
					NumOfInstances:  count(product.equipment @filter(uid(ins_equip)))		
				}  
		  } `

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetProductsForApplications - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProductsForApplications - cannot complete query transaction")
	}

	var prodList v1.ProductsForApplication

	if err := json.Unmarshal(resp.GetJson(), &prodList); err != nil {
		logger.Log.Error("GetProductsForApplications - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("GetProductsForApplications - cannot unmarshal Json object")
	}
	return &prodList, nil

}

func applicationFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := keyToPredForApplication(f.Key())
		if err != nil {
			logger.Log.Error("applicationFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case appPredName, appPredOwner:
			filters = append(filters, fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, f.Value()))
		}
	}
	return filters
}

func applicationFilterForGetApplicationsForProduct(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := keyToPredForApplication(f.Key())
		if err != nil {
			logger.Log.Error("applicationFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case appPredName, appPredOwner:
			filters = append(filters, fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, f.Value()))
		}
	}
	return filters
}

func keyToPredForApplication(key string) (appPred, error) {
	switch key {
	case "name":
		return appPredName, nil
	case "applicationId":
		return appPredID, nil
	case "application_owner":
		return appPredOwner, nil
	case "numOfInstances":
		return appPredNumOfIns, nil
	case "numofProducts":
		return appPredNumOfProd, nil
	case "numofEquipments":
		return appPredNumOfEqp, nil
	default:
		return "", fmt.Errorf("keyToPredForApplication - cannot find dgraph predicate for key: %s", key)
	}
}

func keyToPredForGetApplicationsForProduct(key string) (appPred, error) {
	switch key {
	case "name":
		return appPredName, nil
	case "applicationId":
		return appPredID, nil
	case "application_owner":
		return appPredOwner, nil
	case "numOfInstances":
		return appPredNumOfIns, nil
	case "numofProducts":
		return appPredNumOfProd, nil
	case "numofEquipments":
		return appPredNumOfEqp, nil
	default:
		return "", fmt.Errorf("keyToPredForApplication - cannot find dgraph predicate for key: %s", key)
	}
}
