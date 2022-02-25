package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
)

// ProductAggregationDetails ...
// nolint: funlen
func (l *LicenseRepository) GetAggregationDetails(ctx context.Context, name string, scopes ...string) (*v1.AggregationInfo, error) {
	q := `{
			var(func:eq(aggregation.name,"` + name + `")) ` + agregateFilters(scopeFilters(scopes)) + ` {
				AggUID as uid
				aggregation.products{
					apps as count(~application.product)
					equips as count(product.equipment)
				}
				tapps as sum(val(apps))
				tequips as sum(val(equips))
			}
			AggregatedRight(func: uid(AggUID)){
				ID						:aggregation.id                           
				Name					:aggregation.name                         
				SKU						:aggregation.SKU 
				ProductNames			:aggregation.product_names
				Swidtags				:aggregation.swidtags
				Editor					:aggregation.editor         
				Metric					:aggregation.metric                                           
				Licenses				:aggregation.numOfAcqLicences             
				MaintenanceLicenses		:aggregation.numOfLicencesUnderMaintenance
				UnitPrice				:aggregation.averageUnitPrice             
				MaintenanceUnitPrice	:aggregation.averageMaintenananceUnitPrice   
				PurchaseCost			:aggregation.totalPurchaseCost            
				MaintenanceCost			:aggregation.totalMaintenanceCost         
				TotalCost				:aggregation.totalCost   
				StartOfMaintenance		:aggregation.startOfMaintenance                 
				EndOfMaintenance		:aggregation.endOfMaintenance 
				NumOfApplications		:val(tapps)
				NumOfEquipments			:val(tequips)
				ProductIDs				:aggregation.products{uid}
			}
	}
	`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("GetAggregationDetails - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("repo/GetAggregationDetails - cannot complete query")
	}

	type aggregation struct {
		ID                   int
		Name                 string
		SKU                  string
		ProductNames         []string
		Swidtags             []string
		Editor               string
		Metric               []string
		Licenses             int32
		MaintenanceLicenses  int32
		UnitPrice            float64
		MaintenanceUnitPrice float64
		PurchaseCost         float64
		MaintenanceCost      float64
		TotalCost            float64
		StartOfMaintenance   string
		EndOfMaintenance     string
		NumOfApplications    int32
		NumOfEquipments      int32
		ProductIDs           []*id
	}

	type dataTemp struct {
		AggregatedRight []*aggregation
	}
	data := dataTemp{}
	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("GetAggregationDetails - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("repo/GetAggregationDetails - cannot unmarshal Json object")
	}
	if len(data.AggregatedRight) == 0 {
		return nil, v1.ErrNodeNotFound
	}
	aggresp := &v1.AggregationInfo{
		ID:                   data.AggregatedRight[0].ID,
		Name:                 data.AggregatedRight[0].Name,
		SKU:                  data.AggregatedRight[0].SKU,
		ProductNames:         data.AggregatedRight[0].ProductNames,
		Swidtags:             data.AggregatedRight[0].Swidtags,
		Editor:               data.AggregatedRight[0].Editor,
		Metric:               data.AggregatedRight[0].Metric,
		Licenses:             data.AggregatedRight[0].Licenses,
		MaintenanceLicenses:  data.AggregatedRight[0].MaintenanceLicenses,
		UnitPrice:            data.AggregatedRight[0].UnitPrice,
		MaintenanceUnitPrice: data.AggregatedRight[0].MaintenanceUnitPrice,
		PurchaseCost:         data.AggregatedRight[0].PurchaseCost,
		MaintenanceCost:      data.AggregatedRight[0].MaintenanceCost,
		TotalCost:            data.AggregatedRight[0].TotalCost,
		StartOfMaintenance:   data.AggregatedRight[0].StartOfMaintenance,
		EndOfMaintenance:     data.AggregatedRight[0].EndOfMaintenance,
		NumOfApplications:    data.AggregatedRight[0].NumOfApplications,
		NumOfEquipments:      data.AggregatedRight[0].NumOfEquipments,
		ProductIDs:           convertUIDToString(data.AggregatedRight[0].ProductIDs),
	}
	return aggresp, nil
}

func (l *LicenseRepository) AggregationIndividualRights(ctx context.Context, productIDs, metrics []string, scopes ...string) ([]*v1.AcqRightsInfo, error) {
	q := `{
			var(func: uid(` + strings.Join(productIDs, ",") + `)){
				product.acqRights @filter(eq(acqRights.metric,[` + strings.Join(metrics, ",") + `])){
					individualAcqs as uid
				}
			}
			IndividualRights(func: uid(individualAcqs)) {                        
				SKU						:acqRights.SKU
				Swidtag					:acqRights.swidtag
				ProductName				:acqRights.productName
				ProductEditor			:acqRights.editor   
				ProductVersion			:acqRights.version
				Metric					:acqRights.metric                                           
			    Licenses				:acqRights.numOfAcqLicences             
				MaintenanceLicenses		:acqRights.numOfLicencesUnderMaintenance
				UnitPrice				:acqRights.averageUnitPrice             
				MaintenanceUnitPrice	:acqRights.averageMaintenananceUnitPrice   
				PurchaseCost			:acqRights.totalPurchaseCost            
				MaintenanceCost			:acqRights.totalMaintenanceCost         
				TotalCost				:acqRights.totalCost      
				StartOfMaintenance		:acqRights.startOfMaintenance              
				EndOfMaintenance		:acqRights.endOfMaintenance 
			}
	}
	`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("AggregationIndividualRights - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("repo/AggregationIndividualRights - cannot complete query ")
	}

	type dataTemp struct {
		IndividualRights []*v1.AcqRightsInfo
	}

	data := dataTemp{}

	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("AggregationIndividualRights - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("repo/AggregationIndividualRights - cannot unmarshal Json object")
	}

	if len(data.IndividualRights) == 0 {
		return data.IndividualRights, v1.ErrNodeNotFound
	}
	return data.IndividualRights, nil
}

func convertUIDToString(ids []*id) []string {
	strids := []string{}
	for _, id := range ids {
		strids = append(strids, id.ID)
	}
	return strids
}

// func productAggregationsSortBy(sortBy v1.ProductAggSortBy) (predProductAgg, error) {
// 	switch sortBy {
// 	case v1.ProductAggSortByName:
// 		return predProductAggName, nil
// 	case v1.ProductAggSortByEditor:
// 		return predProductAggEditor, nil
// 	case v1.ProductAggSortByNumApp:
// 		return predProductAggNumApps, nil
// 	case v1.ProductAggSortByNumEquips:
// 		return predProductAggNumEquips, nil
// 	case v1.ProductAggSortByProductName:
// 		return predProductAggName, nil
// 	default:
// 		return "", fmt.Errorf("productAggregationsSortBy - unknown sortBy: %v", sortBy)
// 	}
// }

// func productAggPredForFilteringKey(key v1.ProductAggSearchKey) (predProductAgg, error) {
// 	switch key {
// 	case v1.ProductAggSearchKeyName:
// 		return predProductAggName, nil
// 	case v1.ProductAggSearchKeyEditor:
// 		return predProductAggEditor, nil
// 	case v1.ProductAggSearchKeyProductName:
// 		return predProductAggProductName, nil
// 	default:
// 		return "", fmt.Errorf("productAggPredForFilteringKey - unknown filtering key: %v", key)
// 	}
// }

// func metricFilter(filter *v1.AggregateFilter) []string {
// 	if filter == nil {
// 		return nil
// 	}
// 	filters := make([]string, 0, len(filter.Filters))
// 	for _, f := range filter.Filters {
// 		pred, err := predMetricForSearchKey(v1.MetricSearchKey(f.Key()))
// 		if err != nil {
// 			logger.Log.Error("dgraph - productAggFilter - ", zap.String("reason", err.Error()))
// 			continue
// 		}
// 		switch pred {
// 		case predMetricName:
// 			filters = append(filters, stringFilter(pred.String(), f))
// 		}
// 	}
// 	return filters
// }

// func predMetricForSearchKey(key v1.MetricSearchKey) (predMetric, error) {
// 	switch key {
// 	case v1.MetricSearchKeyName:
// 		return predMetricName, nil
// 	default:
// 		return "", fmt.Errorf("search is not supported on metric - %v", key)
// 	}
// }
