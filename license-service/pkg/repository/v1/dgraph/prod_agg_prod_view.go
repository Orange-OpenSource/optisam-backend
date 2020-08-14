// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"sort"

	"go.uber.org/zap"
)

type predProductAgg string

const (
	predProductAggName        predProductAgg = "product_aggregation.name"
	predProductAggEditor      predProductAgg = "product_aggregation.editor"
	predProductAggProductName predProductAgg = "product_aggregation.product_name"
	predProductAggNumApps     predProductAgg = "val(tapps)"
	predProductAggNumEquips   predProductAgg = "val(tequips)"
)

func (p predProductAgg) String() string {
	return string(p)
}

// ProductAggregationDetails ...
func (lr *LicenseRepository) ProductAggregationDetails(ctx context.Context, name string, params *v1.QueryProductAggregations, scopes []string) (*v1.ProductAggregation, error) {
	q := `
	{
		ID as var(func:eq(product_aggregation.name,"` + name + `"))@cascade` + agregateFilters(scopeFilters(scopes)) + `{
			product_aggregation.metric {
					mn as metric.name
			}
			mna as sum(val(mn))
			Products: product_aggregation.products {
				product.acqRights{
			   }
		    }
		}

		var(func:uid(ID)){
			product_aggregation.products{
				product.acqRights {
					p_ct as acqRights.totalCost
				}
			p_ctp as sum(val(p_ct))
			}
		}

		var(func:uid(ID))@cascade{
			product_aggregation.products{
				apps as count(~application.product)
				equips as count(product.equipment)
				product.acqRights  @filter(eq(acqRights.metric,val(mna))){
					ct as acqRights.totalCost
				}
				ctp as sum(val(ct))
			}
		    tc as sum(val(ctp))
		    tapps as sum(val(apps))
		    tequips as sum(val(equips))
		}
		
		ProductAggregations(func:uid(ID)){
			ID  : uid
			Name:        product_aggregation.name
			Product: product_aggregation.product_name
			Editor:      product_aggregation.editor
			Metric:      product_aggregation.metric{
				Name: metric.name
			}
			Products:    product_aggregation.products {
			   Name :              product.name
			   Version  :          product.version
			   Category :          product.category
			   Editor :            product.editor
			   Swidtag :           product.swidtag
			   NumOfEquipments :   val(equips)
			   NumOfApplications : val(apps)
			   AcqRights: product.acqRights @filter(eq(acqRights.metric,val(mna))) {
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
				TotalCost: val(p_ctp)
			}
			NumOfApplications:   val(tapps)
			NumOfEquipments:     val(tequips)
			TotalCost:   val(tc)
			}
		}
	`

	fmt.Println(q)
	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductAggregationDetails - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductAggregationDetails - cannot complete query ")
	}

	type metric struct {
		Name string
	}

	type product struct {
		Name              string
		Version           string
		Category          string
		Editor            string
		Swidtag           string
		NumOfEquipments   int32
		NumOfApplications int32
		TotalCost         float64
		AcqRights         []*v1.AcquiredRights
	}

	type productAggregation struct {
		ID                string
		Name              string
		Editor            string
		Product           string
		Metric            *metric
		Products          []*product
		NumOfApplications int
		NumOfEquipments   int
		TotalCost         float64
	}

	type dataTemp struct {
		ProductAggregations []*productAggregation
	}

	data := dataTemp{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("ListProductAggregationsProductView - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ListProductAggregationsProductView - cannot unmarshal Json object")
	}

	productAggs := make([]*v1.ProductAggregation, len(data.ProductAggregations))
	for i := range data.ProductAggregations {
		pa := data.ProductAggregations[i]
		var metric string
		metric = pa.Metric.Name
		products := make([]string, len(pa.Products))
		productsFull := make([]*v1.ProductData, len(pa.Products))
		var acqRightRithtsFull []*v1.AcquiredRights
		var acqRights []string
		for i := range pa.Products {
			products[i] = pa.Products[i].Swidtag
			productsFull[i] = &v1.ProductData{
				Name:              pa.Products[i].Name,
				Version:           pa.Products[i].Version,
				Category:          pa.Products[i].Category,
				Editor:            pa.Products[i].Editor,
				Swidtag:           pa.Products[i].Swidtag,
				NumOfEquipments:   pa.Products[i].NumOfEquipments,
				NumOfApplications: pa.Products[i].NumOfApplications,
				TotalCost:         pa.Products[i].TotalCost,
			}
			for _, acqR := range pa.Products[i].AcqRights {
				if metric != acqR.Metric {
					// we only want acquired rights for aggregation metric
					continue
				}
				acqRights = append(acqRights, acqR.SKU)
				acqRightRithtsFull = append(acqRightRithtsFull, acqR)
			}
		}
		productAggs[i] = &v1.ProductAggregation{
			ID:                pa.ID,
			Name:              pa.Name,
			Product:           pa.Product,
			Editor:            pa.Editor,
			Metric:            metric,
			NumOfApplications: pa.NumOfApplications,
			NumOfEquipments:   pa.NumOfEquipments,
			TotalCost:         pa.TotalCost,
			Products:          products,
			ProductsFull:      productsFull,
			AcqRights:         acqRights,
			AcqRightsFull:     acqRightRithtsFull,
		}
	}

	if len(productAggs) == 0 {
		return nil, v1.ErrNodeNotFound
	}

	return productAggs[0], nil
}

func productAggregationsSortBy(sortBy v1.ProductAggSortBy) (predProductAgg, error) {
	switch sortBy {
	case v1.ProductAggSortByName:
		return predProductAggName, nil
	case v1.ProductAggSortByEditor:
		return predProductAggEditor, nil
	case v1.ProductAggSortByNumApp:
		return predProductAggNumApps, nil
	case v1.ProductAggSortByNumEquips:
		return predProductAggNumEquips, nil
	case v1.ProductAggSortByProductName:
		return predProductAggName, nil
	default:
		return "", fmt.Errorf("productAggregationsSortBy - unknown sortBy: %v", sortBy)
	}
}

func productAggPredForFilteringKey(key v1.ProductAggSearchKey) (predProductAgg, error) {
	switch key {
	case v1.ProductAggSearchKeyName:
		return predProductAggName, nil
	case v1.ProductAggSearchKeyEditor:
		return predProductAggEditor, nil
	case v1.ProductAggSearchKeyProductName:
		return predProductAggProductName, nil
	default:
		return "", fmt.Errorf("productAggPredForFilteringKey - unknown filtering key: %v", key)
	}
}

func productAggFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := productAggPredForFilteringKey(v1.ProductAggSearchKey(f.Key()))
		if err != nil {
			logger.Log.Error("dgraph - productAggFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case predProductAggName, predProductAggEditor, predProductAggProductName:
			filters = append(filters, stringFilter(pred.String(), f))
		}
	}
	return filters
}

func metricFilter(filter *v1.AggregateFilter) []string {
	if filter == nil {
		return nil
	}
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := predMetricForSearchKey(v1.MetricSearchKey(f.Key()))
		if err != nil {
			logger.Log.Error("dgraph - productAggFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case predMetricName:
			filters = append(filters, stringFilter(pred.String(), f))
		}
	}
	return filters
}

func predMetricForSearchKey(key v1.MetricSearchKey) (predMetric, error) {
	switch key {
	case v1.MetricSearchKeyName:
		return predMetricName, nil
	default:
		return "", fmt.Errorf("search is not supported on metric - %v", key)
	}
}
