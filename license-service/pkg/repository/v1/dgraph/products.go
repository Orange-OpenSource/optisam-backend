package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"regexp"
	"sort"
	"strings"

	"go.uber.org/zap"
)

type prodPred string

func (p prodPred) String() string {
	return string(p)
}

const (
	prodPredName    prodPred = "product.name"
	prodPredSwidTag prodPred = "product.swidtag"
	prodPredEditor  prodPred = "product.editor"
)

type productAcquiredRight struct {
	SKU               string
	Metric            []string
	AcqLicenses       uint64
	TotalCost         float64
	TotalPurchaseCost float64
	AvgUnitPrice      float64
	Repartition       bool
}

// IsProductPurchasedInAggregation return aggregation name is swidtag is part of aggregation
func (l *LicenseRepository) IsProductPurchasedInAggregation(ctx context.Context, swidtag string, scope string) (string, error) {
	q := ` query {
		aggregation(func:type(Aggregation))@filter(eq(scopes, ` + scope + ` ) and eq(aggregation.swidtags, "` + swidtag + `" )){
  			 aggregation.name
			}
		}
		`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("GetProductInformation - ", zap.String("reason", err.Error()), zap.String("query", q))
		return "", fmt.Errorf("getProductInformation - cannot complete query transaction")
	}

	type Aggregation struct {
		Aggregation []struct {
			Name string `json:"aggregation.name"`
		} `json:"aggregation"`
	}

	// fmt.Println(string(resp.GetJson()))
	out := Aggregation{}
	if err = json.Unmarshal(resp.GetJson(), &out); err != nil {
		logger.Log.Error("Failed to marshal the product-aggregation link", zap.Error(err))
		return "", err
	}
	// fmt.Println(out)
	if len(out.Aggregation) > 0 {
		return out.Aggregation[0].Name, nil
	}
	return "", nil
}

// GetProductInformation ...
func (l *LicenseRepository) GetProductInformation(ctx context.Context, swidtag string, scopes ...string) (*v1.ProductAdditionalInfo, error) {

	variables := make(map[string]string)
	variables["$tag"] = swidtag

	q := `query OrderProducts($tag:string) {
		Products(func: eq(product.swidtag, $tag)) ` + agregateFilters(scopeFilters(scopes), typeFilters("type_name", "product")) + ` {    
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
	resp, err := l.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("GetProductInformation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("getProductInformation - cannot complete query transaction")
	}

	var ProductDetails v1.ProductAdditionalInfo

	if err := json.Unmarshal(resp.GetJson(), &ProductDetails); err != nil {
		logger.Log.Error("GetProductInformation - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", variables))
		return nil, fmt.Errorf("getProductInformation - cannot unmarshal Json object")
	}
	return &ProductDetails, nil
}

// ProductAcquiredRights implements Licence ProductAcquiredRights function
func (l *LicenseRepository) ProductAcquiredRights(ctx context.Context, swidTag string, metrics []*v1.Metric, isSimulation bool, scopes ...string) (string, string, []*v1.ProductAcquiredRight, error) {
	q := `
	{
		Products(func: eq(product.swidtag,"` + swidTag + `")) ` + agregateFilters(scopeFilters(scopes), typeFilters("type_name", "product")) + ` {
		  ID: uid
		  ProductName: product.name
		  AcquiredRights: product.acqRights{
		  SKU: acqRights.SKU
		  Metric: acqRights.metric
		  AcqLicenses: acqRights.numOfAcqLicences
		  TotalCost: acqRights.totalCost
		  TotalPurchaseCost: acqRights.totalPurchaseCost
		  AvgUnitPrice: acqRights.averageUnitPrice
		  Repartition: acqRights.repartition 
		}
		}
	  }
	`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ProductAcquiredRights - query failed", zap.Error(err), zap.String("query", q))
		return "", "", nil, errors.New("dgraph/ProductAcquiredRights -  failed to fetch acquired rights")
	}

	type product struct {
		ID             string
		ProductName    string
		AcquiredRights []*productAcquiredRight
	}

	type products struct {
		Products []*product
	}

	data := &products{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/ProductAcquiredRights - unmarshal failed", zap.Error(err), zap.String("query", q), zap.String("response", string(resp.Json)))
		return "", "", nil, errors.New("dgraph/ProductAcquiredRights -  failed unmarshal response")
	}

	if len(data.Products) == 0 {
		logger.Log.Error("dgraph/ProductAcquiredRights -failed", zap.String("reason", "no data found i"))
		return "", "", nil, v1.ErrNodeNotFound
	}
	productName := ""
	if data.Products[0].ProductName != "" {
		productName = data.Products[0].ProductName
	}
	prodRights := l.ConcatAcqRightForSameMetric(ctx, metrics, data.Products[0].AcquiredRights, isSimulation, scopes[0])
	return data.Products[0].ID, productName, prodRights, nil
}

// ProdAllocatedMetric
func (l *LicenseRepository) GetProdAllocatedMetric(ctx context.Context, pID []string, scopes ...string) ([]*v1.ProductAllocationEquipmentMetrics, error) {
	q := `
		{
			Products(func: uid(` + strings.Join(pID, ",") + `)) {
				SwidTag : product.swidtag
				MetricAllocation: product.allocation @filter(eq(scopes,` + scopes[0] + `)){ 
					EquipmentId: equipment.id
					MetricAllocated : allocation.metric
				}
				ProductEquipment: product.equipment @filter(eq(scopes,[` + scopes[0] + `])){
					EUID : 	uid
					EquipmentId  :	equipment.id
					EquipmentType 	: equipment.type
				}
			}
		}`

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ProdAllocatedMetric - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New("dgraph/ProdAllocatedMetric -  failed to fetch allocated metric")
	}

	type products struct {
		Products []*v1.ProductAllocationEquipmentMetrics
	}

	data := &products{}

	if err := json.Unmarshal(resp.Json, data); err != nil {
		logger.Log.Error("dgraph/ProdAllocatedMetric - unmarshal failed", zap.Error(err), zap.String("query", q), zap.String("response", string(resp.Json)))
		return nil, errors.New("dgraph/ProdAllocatedMetric -  failed unmarshal response")
	}

	return data.Products, nil
}

// ProductEquipments implements Licence ProductEquipments function
func productFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := searchKeyForProduct(v1.ProductSearchKey(f.Key()))
		if err != nil {
			logger.Log.Error("dgraph - productFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case prodPredSwidTag, prodPredName, prodPredEditor:
			filters = append(filters, stringFilter(pred.String(), f))
		}
	}
	return filters
}

func stringFilterValString(typ v1.Filtertype, pred string, val string) string {
	switch typ {
	case v1.EqFilter:
		return fmt.Sprintf(" (eq(%v,\"%v\")) ", pred, val)
	case v1.RegexFilter:
		val = regexp.QuoteMeta(val)
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	default:
		val = regexp.QuoteMeta(val)
		// By default, regex filter is used.
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	}
}

func stringFilterSingle(typ v1.Filtertype, pred string, val interface{}) string {
	strVal, ok := val.(string)
	if ok {
		return stringFilterValString(typ, pred, strVal)
	}
	switch typ {
	case v1.EqFilter:
		return fmt.Sprintf(" (eq(%v,\"%v\")) ", pred, val)
	case v1.RegexFilter:
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	default:
		// By default, regex filter is used.
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	}
}

func stringFilter(pred string, q v1.Queryable) string {
	vals := q.Values()
	if len(vals) == 0 {
		return stringFilterSingle(q.Type(), pred, q.Value())
	}
	filters := make([]string, 0, len(vals))
	for _, val := range vals {
		filters = append(filters, stringFilterSingle(q.Type(), pred, val))
	}
	return " ( " + strings.Join(filters, "OR") + " ) "
}

func searchKeyForProduct(key v1.ProductSearchKey) (prodPred, error) {
	switch key {
	case v1.ProductSearchKeySwidTag:
		return prodPredSwidTag, nil
	case v1.ProductSearchKeyName:
		return prodPredName, nil
	case v1.ProductSearchKeyEditor:
		return prodPredEditor, nil
	default:
		return "", fmt.Errorf("searchKeyForProduct - unknown product search key, %v ", key)
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
	return ` @filter(eq(aggregation.metric,["` + fmt.Sprintf("%v", filter.Filters[0].Value()) + `"]))`
}

// ConcatAcqRightForSameMetric accmodate same type metrics
func (l *LicenseRepository) ConcatAcqRightForSameMetric(ctx context.Context, metrics []*v1.Metric, acqRight []*productAcquiredRight, isSimulation bool, scope string) []*v1.ProductAcquiredRight {
	resAcqRight := make([]*v1.ProductAcquiredRight, 0, len(acqRight))
	if isSimulation {
		for _, acq := range acqRight {
			sort.Strings(acq.Metric)
			resAcqRight = append(resAcqRight, &v1.ProductAcquiredRight{
				SKU:               acq.SKU,
				Metric:            strings.Join(acq.Metric, ","),
				AcqLicenses:       acq.AcqLicenses,
				TotalCost:         acq.TotalCost,
				TotalPurchaseCost: acq.TotalPurchaseCost,
				AvgUnitPrice:      acq.AvgUnitPrice,
			})
		}
		return resAcqRight
	}
	encountered := map[string]int{}
	metricType := map[string]string{}
	for i := range acqRight {
		metric := strings.Join(acqRight[i].Metric, ",")
		if len(acqRight[i].Metric) == 1 {
			for _, met := range metrics {
				if met.Name == metric {
					metricType[metric] = met.Type.String()
				}
			}
		}
		if acqRight[i].Repartition {
			resAcqRight = append(resAcqRight, &v1.ProductAcquiredRight{
				SKU:               acqRight[i].SKU,
				Metric:            metric,
				AcqLicenses:       acqRight[i].AcqLicenses,
				TotalCost:         acqRight[i].TotalCost,
				TotalPurchaseCost: acqRight[i].TotalPurchaseCost,
				AvgUnitPrice:      acqRight[i].AvgUnitPrice,
				Repartition:       acqRight[i].Repartition,
			})
		} else {
			idx, ok := encountered[metric]
			if ok {
				// Add values to original.
				resAcqRight[idx].SKU = strings.Join([]string{resAcqRight[idx].SKU, acqRight[i].SKU}, ",")
				resAcqRight[idx].AcqLicenses += acqRight[i].AcqLicenses
				resAcqRight[idx].TotalCost += acqRight[i].TotalCost
				resAcqRight[idx].TotalPurchaseCost += acqRight[i].TotalPurchaseCost
				resAcqRight[idx].AvgUnitPrice += acqRight[i].AvgUnitPrice
			} else {
				// check all keys if it matches unordered list
				found := false
				for k, j := range encountered {
					encMet := strings.Split(k, ",")
					if helper.CompareSlices(encMet, acqRight[i].Metric) {
						resAcqRight[j].SKU = strings.Join([]string{resAcqRight[j].SKU, acqRight[i].SKU}, ",")
						resAcqRight[j].AcqLicenses += acqRight[i].AcqLicenses
						resAcqRight[j].TotalCost += acqRight[i].TotalCost
						resAcqRight[j].TotalPurchaseCost += acqRight[i].TotalPurchaseCost
						resAcqRight[j].AvgUnitPrice += acqRight[i].AvgUnitPrice
						found = true
						break
					}
				}
				if found {
					continue
				}
				// Record this element as an encountered element.
				encountered[metric] = len(resAcqRight)
				// Append to result slice.
				resAcqRight = append(resAcqRight, &v1.ProductAcquiredRight{
					SKU:               acqRight[i].SKU,
					Metric:            metric,
					AcqLicenses:       acqRight[i].AcqLicenses,
					TotalCost:         acqRight[i].TotalCost,
					TotalPurchaseCost: acqRight[i].TotalPurchaseCost,
					AvgUnitPrice:      acqRight[i].AvgUnitPrice,
					Repartition:       acqRight[i].Repartition,
				})
			}
		}
	}
	for _, acq := range resAcqRight {
		if acq.AcqLicenses != 0 {
			acq.AvgUnitPrice = acq.TotalPurchaseCost / float64(acq.AcqLicenses)
		} else {
			acq.AvgUnitPrice = acq.TotalPurchaseCost / float64(len(strings.Split(acq.SKU, ",")))
		}
	}
	if len(metricType) > 1 {
		for nupMetric, indexNup := range metricType {
			if indexNup == v1.MetricOracleNUPStandard.String() {
				// When NUP metric is present then get all info of metric
				metricNup, err := l.GetMetricConfigNUPID(ctx, nupMetric, scope)
				if err != nil {
					logger.Log.Sugar().Debugw("service/v1 - ListAcqRightsForProduct - error while getting transformed metric config", "error", err.Error())
				}
				if metricNup.Transform {
					for opsMetric, opsType := range metricType {
						if metricNup.TransformMetricName == opsMetric && opsType == v1.MetricOPSOracleProcessorStandard.String() {
							// opsidx := encountered[opsMetric]
							// nupidx := encountered[nupMetric]
							opsidx, nupidx := getAcqRightsIndex(resAcqRight, opsMetric, nupMetric)
							resAcqRight[opsidx].SKU = strings.Join([]string{resAcqRight[opsidx].SKU, resAcqRight[nupidx].SKU}, ",")
							//resAcqRight[opsidx].AcqLicenses += uint64(math.Ceil(float64(resAcqRight[nupidx].AcqLicenses) / 50))
							// resAcqRight[opsidx].TotalCost = (resAcqRight[opsidx].TotalCost + acqRight[nupidx].TotalCost) / 2
							// resAcqRight[opsidx].TotalPurchaseCost = (resAcqRight[opsidx].TotalPurchaseCost + acqRight[nupidx].TotalPurchaseCost) / 2
							// resAcqRight[opsidx].AvgUnitPrice = resAcqRight[opsidx].AvgUnitPrice
							resAcqRight[opsidx].TransformDetails = nupMetric + " is transformed into " + opsMetric
							resAcqRight[nupidx] = resAcqRight[len(resAcqRight)-1] // Copy last element to index.
							resAcqRight = resAcqRight[:len(resAcqRight)-1]        // Truncate slice.
							break
						}
					}
				}
			}
		}
	}

	return resAcqRight
}

func getAcqRightsIndex(resAcqRight []*v1.ProductAcquiredRight, metricOps string, metricNUP string) (int, int) {
	var opsIndex, nupIndex int
	for index, acqRight := range resAcqRight {
		if strings.EqualFold(acqRight.Metric, metricOps) {
			opsIndex = index
		}
		if strings.EqualFold(acqRight.Metric, metricNUP) {
			nupIndex = index
		}
	}
	return opsIndex, nupIndex
}

// GetProductsByEditorProductName will return all product version of product & editor
func (l *LicenseRepository) GetProductsByEditorProductName(ctx context.Context, metrics []*v1.Metric, scope, editorName, productName string) ([]*v1.ProductDetail, error) {
	q := `
		{
			Products(func:eq(product.editor,"` + editorName + `"),orderasc: product.version) @filter(eq(scopes,"` + scope + `") AND eq(product.name,"` + productName + `") AND eq(type_name,"product")) {
				ID : uid
				SwidTag : 		   product.swidtag
				Name :    		   product.name
				Version : 		   product.version
				Editor :  		   product.editor
				Edition : 			product.edition
				NumofEquipments:   count(product.equipment)
				AcquiredRights: product.acqRights{
					SKU: acqRights.SKU
					Metric: acqRights.metric
					AcqLicenses: acqRights.numOfAcqLicences
					TotalCost: acqRights.totalCost
					TotalPurchaseCost: acqRights.totalPurchaseCost
					AvgUnitPrice: acqRights.averageUnitPrice
					Repartition: acqRights.repartition 
				}
			}
		}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("GetProductsByEditorProductName - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", editorName))
		return nil, fmt.Errorf("GetProductsByEditorProductName - cannot complete query transaction")
	}
	type product struct {
		ID              string
		SwidTag         string
		Name            string
		Edition         string
		Editor          string
		Version         string
		NumOfEquipments int32
		AcquiredRights  []*productAcquiredRight
	}

	type products struct {
		Products []*product
	}
	data := &products{}
	if err := json.Unmarshal(resp.GetJson(), &data); err != nil {
		logger.Log.Error("GetProductsByEditorProductName - ", zap.String("reason", err.Error()), zap.String("query", q), zap.Any("query params", editorName))
		return nil, fmt.Errorf("GetProductsByEditorProductName - cannot unmarshal Json object")
	}
	var Products []*v1.ProductDetail
	for _, product := range data.Products {
		var productTmp v1.ProductDetail
		productTmp.ID = product.ID
		productTmp.SwidTag = product.SwidTag
		productTmp.Name = product.Name
		productTmp.Edition = product.Edition
		productTmp.Editor = product.Editor
		productTmp.Version = product.Version
		productTmp.NumOfEquipments = product.NumOfEquipments
		prodRights := l.ConcatAcqRightForSameMetric(ctx, metrics, product.AcquiredRights, false, scope)
		productTmp.AcquiredRights = prodRights
		Products = append(Products, &productTmp)
	}
	// fmt.Println("format string", Products)
	return Products, nil
}

// GetProductInformationFromAcqRight will fetch product info from acq rights
func (l *LicenseRepository) GetProductInformationFromAcqRight(ctx context.Context, swidtag string, scopes ...string) (*v1.ProductAdditionalInfo, error) {

	q := `{
		Products(func: eq(acqRights.swidtag, ` + swidtag + `)) ` + agregateFilters(scopeFilters(scopes), typeFilters("type_name", "acqRights")) + ` {    
			Swidtag : 		   acqRights.swidtag
			Name :    		   acqRights.productName 
			Version : 		   acqRights.version
			Editor :  		   acqRights.editor
		}

	}`

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Sugar().Errorw("GetProductInformationFromAcqRight - Error while getting product info from acqights",
			"error", err.Error(),
			"query", q,
		)
		return nil, fmt.Errorf("GetProductInformationFromAcqRight - cannot complete query transaction")
	}

	var ProductDetails v1.ProductAdditionalInfo

	if err := json.Unmarshal(resp.GetJson(), &ProductDetails); err != nil {
		logger.Log.Sugar().Errorw("GetProductInformationFromAcqRight - Error while unmarshal JSON for getting product info from acqights",
			"error", err.Error(),
			"query", q,
		)
		return nil, fmt.Errorf("GetProductInformationFromAcqRight - cannot unmarshal Json object")
	}
	return &ProductDetails, nil
}
