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

// AggregationDetails gives aggregatation info along with it's associated rights
func (l *LicenseRepository) AggregationDetails(ctx context.Context, name string, metrics []*v1.Metric, isSimulation bool, scopes ...string) (*v1.AggregationInfo, []*v1.ProductAcquiredRight, error) {
	q := `{
			var(func:eq(aggregation.name,"` + name + `")) ` + agregateFilters(scopeFilters(scopes)) + ` {
				aggUID as uid
				aggregationID as aggregation.id
				aggregation.products{
					apps as count(~application.product)
					equips as count(product.equipment)
				}
				tapps as sum(val(apps))
				tequips as sum(val(equips))
			}
			var(func:eq(aggregatedRights.aggregationId,val(aggregationID))) ` + agregateFilters(scopeFilters(scopes)) + ` {
				aggRights as uid
			}
			Aggregation(func:uid(aggUID)) {
				ID						:aggregation.id
				Name					:aggregation.name   
				ProductNames			:aggregation.product_names
				Swidtags				:aggregation.swidtags
				Editor					:aggregation.editor   
				NumOfApplications		:val(tapps)
				NumOfEquipments			:val(tequips)
				ProductIDs				:aggregation.products{uid}  
			}
			AggregatedRight(func: uid(aggRights)){                                
				SKU						:aggregatedRights.SKU
		 		Metric					:aggregatedRights.metric
		  		AcqLicenses				:aggregatedRights.numOfAcqLicences
		  		TotalCost				:aggregatedRights.totalCost
		  		TotalPurchaseCost		:aggregatedRights.totalPurchaseCost
		  		AvgUnitPrice			:aggregatedRights.averageUnitPrice
			}
	}
	`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("GetAggregationDetails - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, nil, fmt.Errorf("repo/GetAggregationDetails - cannot complete query")
	}

	type aggregation struct {
		ID                int
		Name              string
		ProductNames      []string
		Swidtags          []string
		Editor            string
		NumOfApplications int32
		NumOfEquipments   int32
		ProductIDs        []*id
	}

	type dataTemp struct {
		Aggregation     []*aggregation
		AggregatedRight []*productAcquiredRight
	}
	data := dataTemp{}
	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("GetAggregationDetails - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, nil, fmt.Errorf("repo/GetAggregationDetails - cannot unmarshal Json object")
	}
	if len(data.Aggregation) == 0 {
		return nil, nil, v1.ErrNodeNotFound
	}
	respAggInfo := &v1.AggregationInfo{
		ID:                data.Aggregation[0].ID,
		Name:              data.Aggregation[0].Name,
		ProductNames:      data.Aggregation[0].ProductNames,
		Swidtags:          data.Aggregation[0].Swidtags,
		Editor:            data.Aggregation[0].Editor,
		NumOfApplications: data.Aggregation[0].NumOfApplications,
		NumOfEquipments:   data.Aggregation[0].NumOfEquipments,
		ProductIDs:        convertUIDToString(data.Aggregation[0].ProductIDs),
	}
	if len(data.AggregatedRight) == 0 {
		return respAggInfo, nil, nil
	}
	return respAggInfo, concatAcqRightForSameMetric(metrics, data.AggregatedRight, isSimulation), nil
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
