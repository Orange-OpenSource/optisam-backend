package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
		Products(func: eq(product.swidtag,"` + swidTag + `"))` + agregateFilters(scopeFilters(scopes)) + `{
		  ID: uid
		  ProductName: product.name
		  AcquiredRights: product.acqRights{
		  SKU: acqRights.SKU
		  Metric: acqRights.metric
		  AcqLicenses: acqRights.numOfAcqLicences
		  TotalCost: acqRights.totalCost
		  TotalPurchaseCost: acqRights.totalPurchaseCost
		  AvgUnitPrice: acqRights.averageUnitPrice
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

	return data.Products[0].ID, data.Products[0].ProductName, concatAcqRightForSameMetric(metrics, data.Products[0].AcquiredRights, isSimulation), nil
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

func concatAcqRightForSameMetric(metrics []*v1.Metric, acqRight []*productAcquiredRight, isSimulation bool) []*v1.ProductAcquiredRight {
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
	metricType := map[v1.MetricType]string{}
	for i := range acqRight {
		metric := strings.Join(acqRight[i].Metric, ",")
		if len(acqRight[i].Metric) == 1 {
			for _, met := range metrics {
				if met.Name == metric {
					metricType[met.Type] = metric
				}
			}
		}
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
			})
		}
	}
	for _, acq := range resAcqRight {
		if acq.AcqLicenses != 0 {
			acq.AvgUnitPrice = acq.TotalPurchaseCost / float64(acq.AcqLicenses)
		} else {
			acq.AvgUnitPrice = acq.TotalPurchaseCost / float64(len(strings.Split(acq.SKU, ",")))
		}
	}
	nupMetric, ok := metricType[v1.MetricOracleNUPStandard]
	if ok {
		opsMetric, ok := metricType[v1.MetricOPSOracleProcessorStandard]
		if ok {
			opsidx, _ := encountered[opsMetric]
			nupidx, _ := encountered[nupMetric]
			resAcqRight[opsidx].SKU = strings.Join([]string{resAcqRight[opsidx].SKU, acqRight[nupidx].SKU}, ",")
			resAcqRight[opsidx].AcqLicenses += uint64(math.Ceil(float64(resAcqRight[nupidx].AcqLicenses) / 50))
			// resAcqRight[opsidx].TotalCost = (resAcqRight[opsidx].TotalCost + acqRight[nupidx].TotalCost) / 2
			// resAcqRight[opsidx].TotalPurchaseCost = (resAcqRight[opsidx].TotalPurchaseCost + acqRight[nupidx].TotalPurchaseCost) / 2
			// resAcqRight[opsidx].AvgUnitPrice = resAcqRight[opsidx].AvgUnitPrice
			resAcqRight[nupidx] = resAcqRight[len(resAcqRight)-1] // Copy last element to index.
			resAcqRight = resAcqRight[:len(resAcqRight)-1]        // Truncate slice.
		}
	}
	return resAcqRight
}
