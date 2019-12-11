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

type predAcqRights string

// String implements stringer
func (p predAcqRights) String() string {
	return string(p)
}

const (
	predAcqRightsEntity                         predAcqRights = "acqRights.entity"
	predAcqRightsSKU                            predAcqRights = "acqRights.SKU"
	predAcqRightsSwidTag                        predAcqRights = "acqRights.swidtag"
	predAcqRightsProductName                    predAcqRights = "acqRights.productName"
	predAcqRightsEditor                         predAcqRights = "acqRights.editor"
	predAcqRightsMetric                         predAcqRights = "acqRights.metric"
	predAcqRightsAcquiredLicensesNumber         predAcqRights = "acqRights.numOfAcqLicences"
	predAcqRightsLicensesUnderMaintenanceNumber predAcqRights = "acqRights.numOfLicencesUnderMaintenance"
	predAcqRightsAvgLicenesUnitPrice            predAcqRights = "acqRights.averageUnitPrice"
	predAcqRightsAvgMaintenanceUnitPrice        predAcqRights = "acqRights.averageMaintenantUnitPrice"
	predAcqRightsTotalPurchaseCost              predAcqRights = "acqRights.totalPurchaseCost"
	predAcqRightsTotalMaintenanceCost           predAcqRights = "acqRights.totalMaintenanceCost"
	predAcqRightsTotalCost                      predAcqRights = "acqRights.totalCost"
)

type totalRecords struct {
	TotalCount int32
}

// AcquiredRights implements License interface AcquiredRights method
func (lr *LicenseRepository) AcquiredRights(ctx context.Context, params *v1.QueryAcquiredRights, scopes []string) (int32, []*v1.AcquiredRights, error) {
	variables := make(map[string]string)
	variables["$offset"] = strconv.Itoa(int(params.Offset))
	variables["$pagesize"] = strconv.Itoa(int(params.PageSize))
	sortOrder, err := sortOrderForDgraph(params.SortOrder)
	if err != nil {
		// As we get some sort order which is undefined we are taking sororder ASC
		logger.Log.Error("dgraph/AcquiredRights - unknown sort order:", zap.String("reason", err.Error()))
		sortOrder = sortASC
	}

	sortBy, err := acquiredRightsSortBy(params.SortBy)
	if err != nil {
		logger.Log.Error("dgraph/AcquiredRights - unknown sort by:", zap.String("reason", err.Error()))
		// if there is a problem in getting sort by we sort by sku as default.
		sortBy = predAcqRightsSKU
	}
	q := `
	  query AcquiredRights($pagesize:string, $offset:string){
		ID  as  var(func:eq(type,"acqRights")) ` + agregateFilters(scopeFilters(scopes), acquiredRightsFilter(params.Filter)) + ` {}
		TotalRecords (func:uid(ID)){
			TotalCount: count(uid)
		  }
		AcquiredRights(func:uid(ID),` + sortOrder.String() + `:` + sortBy.String() + `, first:$pagesize,offset:$offset){
		  Entity                        :  acqRights.entity
		  SKU                           :  acqRights.SKU
		  SwidTag                       :  acqRights.swidtag
		  ProductName                   :  acqRights.productName
		  Editor                        :  acqRights.editor
		  Metric                        :  acqRights.metric
		  AcquiredLicensesNumber        :  acqRights.numOfAcqLicences
		  LicensesUnderMaintenanceNumber:  acqRights.numOfLicencesUnderMaintenance
		  AvgLicenesUnitPrice           :  acqRights.averageUnitPrice
		  AvgMaintenanceUnitPrice       :  acqRights.averageMaintenantUnitPrice
		  TotalPurchaseCost             :  acqRights.totalPurchaseCost
		  TotalMaintenanceCost          :  acqRights.totalMaintenanceCost
		  TotalCost                     :  acqRights.totalCost
		  }
	  }
	`

	resp, err := lr.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("AcquiredRights - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("AcquiredRights - cannot complete query transaction")
	}

	type dataTemp struct {
		TotalRecords   []*totalRecords
		AcquiredRights []*v1.AcquiredRights
	}

	data := dataTemp{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("AcquiredRights - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return 0, nil, fmt.Errorf("AcquiredRights - cannot unmarshal Json object")

	}
	if len(data.TotalRecords) == 0 {
		logger.Log.Error("AcquiredRights - ", zap.String("reason", " total records lenght is zero"))
		return 0, nil, errors.New("AcquiredRights - length of total count cannot be zero")
	}
	return data.TotalRecords[0].TotalCount, data.AcquiredRights, nil
}

func scopeFilters(scopes []string) []string {
	return []string{
		fmt.Sprintf("eq(scopes,[%s])", strings.Join(scopes, ",")),
	}
}

func agregateFilters(filters ...[]string) string {
	var aggFilters []string
	for _, filter := range filters {
		aggFilters = append(aggFilters, filter...)
	}
	return "@filter( " + strings.Join(aggFilters, " AND ") + " )"
}

func acquiredRightsFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := acquiredRightsPredForFilteringKey(v1.AcquiredRightsSearchKey(f.Key()))
		if err != nil {
			logger.Log.Error("dgraph - acquiredRightsFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case predAcqRightsSKU, predAcqRightsSwidTag, predAcqRightsProductName, predAcqRightsEditor, predAcqRightsMetric:
			filters = append(filters, fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, f.Value()))
		}
	}
	return filters
}

func acquiredRightsPredForFilteringKey(key v1.AcquiredRightsSearchKey) (predAcqRights, error) {
	switch key {
	case v1.AcquiredRightsSearchKeySKU:
		return predAcqRightsSKU, nil
	case v1.AcquiredRightsSearchKeySwidTag:
		return predAcqRightsSwidTag, nil
	case v1.AcquiredRightsSearchKeyProductName:
		return predAcqRightsProductName, nil
	case v1.AcquiredRightsSearchKeyEditor:
		return predAcqRightsEditor, nil
	case v1.AcquiredRightsSearchKeyMetric:
		return predAcqRightsMetric, nil
	default:
		return "", fmt.Errorf("acquiredRightsPredForFilteringKey - unknown acquired key")
	}
}

func acquiredRightsSortBy(sortBy v1.AcquiredRightsSortBy) (predAcqRights, error) {
	switch sortBy {
	case v1.AcquiredRightsSortByEntity:
		return predAcqRightsEntity, nil
	case v1.AcquiredRightsSortBySKU:
		return predAcqRightsSKU, nil
	case v1.AcquiredRightsSortBySwidTag:
		return predAcqRightsSwidTag, nil
	case v1.AcquiredRightsSortByProductName:
		return predAcqRightsProductName, nil
	case v1.AcquiredRightsSortByEditor:
		return predAcqRightsEditor, nil
	case v1.AcquiredRightsSortByMetric:
		return predAcqRightsMetric, nil
	case v1.AcquiredRightsSortByAcquiredLicensesNumber:
		return predAcqRightsAcquiredLicensesNumber, nil
	case v1.AcquiredRightsSortByLicensesUnderMaintenanceNumber:
		return predAcqRightsLicensesUnderMaintenanceNumber, nil
	case v1.AcquiredRightsSortByAvgLicenseUnitPrice:
		return predAcqRightsAvgLicenesUnitPrice, nil
	case v1.AcquiredRightsSortByAvgMaintenanceUnitPrice:
		return predAcqRightsAvgMaintenanceUnitPrice, nil
	case v1.AcquiredRightsSortByTotalPurchaseCost:
		return predAcqRightsTotalPurchaseCost, nil
	case v1.AcquiredRightsSortByTotalMaintenanceCost:
		return predAcqRightsTotalMaintenanceCost, nil
	case v1.AcquiredRightsSortByTotalCost:
		return predAcqRightsTotalCost, nil
	default:
		return "", fmt.Errorf("acquiredRightsSortOrder - unknown sortby: %v", sortBy)
	}
}
