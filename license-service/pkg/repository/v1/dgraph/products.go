// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	prodPredName      prodPred = "product.name"
	prodPredSwidTag   prodPred = "product.swidtag"
	prodPredVersion   prodPred = "product.version"
	prodPredEditor    prodPred = "product.editor"
	prodPredNumOfApp  prodPred = "val(numOfApplications)"
	prodPredNumOfEqp  prodPred = "val(numOfEquipments)"
	prodPredTotalCost prodPred = "val(totalCost)"
)

const (
	offset   string = "$offset"
	pagesize string = "$pagesize"
)

// GetProductInformation ...
func (r *LicenseRepository) GetProductInformation(ctx context.Context, swidtag string, scopes ...string) (*v1.ProductAdditionalInfo, error) {

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

// ProductAcquiredRights implements Licence ProductAcquiredRights function
func (r *LicenseRepository) ProductAcquiredRights(ctx context.Context, swidTag string, scopes ...string) (string, []*v1.ProductAcquiredRight, error) {
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
		logger.Log.Error("dgraph/ProductAcquiredRights -failed", zap.String("reason", "no data found i"))
		return "", nil, v1.ErrNodeNotFound
	}

	return data.Products[0].ID, data.Products[0].AcquiredRights, nil
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
	case "totalCost":
		return prodPredTotalCost, nil
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
