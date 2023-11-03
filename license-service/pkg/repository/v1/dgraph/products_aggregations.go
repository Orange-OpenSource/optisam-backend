package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

// // CreateProductAggregation implements Licence CreateProductAggregation function
// func (l *LicenseRepository) CreateProductAggregation(ctx context.Context, pa *v1.ProductAggregation, scopes []string) (retPa *v1.ProductAggregation, retErr error) {
// 	blankID := blankID(pa.Name)
// 	nquads := []*api.NQuad{
// 		{
// 			Subject:     blankID,
// 			Predicate:   "type_name",
// 			ObjectValue: stringObjectValue("product_aggregation"),
// 		},
// 		{
// 			Subject:     blankID,
// 			Predicate:   "product_aggregation.editor",
// 			ObjectValue: stringObjectValue(pa.Editor),
// 		},
// 		{
// 			Subject:     blankID,
// 			Predicate:   "product_aggregation.product_name",
// 			ObjectValue: stringObjectValue(pa.Product),
// 		},
// 		{
// 			Subject:     blankID,
// 			Predicate:   "product_aggregation.name",
// 			ObjectValue: stringObjectValue(pa.Name),
// 		},
// 		{
// 			Subject:   blankID,
// 			Predicate: "product_aggregation.metric",
// 			ObjectId:  pa.Metric,
// 		},
// 		{
// 			Subject:     blankID,
// 			Predicate:   "dgraph.type",
// 			ObjectValue: stringObjectValue("ProductAggregation"),
// 		},
// 	}

// 	nquads = append(nquads, productsNquad(pa.Products, blankID)...)
// 	nquads = append(nquads, scopesNquad(scopes, blankID)...)

// 	mu := &api.Mutation{
// 		Set: nquads,
// 		//	CommitNow: true,
// 	}
// 	txn := l.dg.NewTxn()

// 	defer func() {
// 		if retErr != nil {
// 			if err := txn.Discard(ctx); err != nil {
// 				logger.Log.Error("dgraph/CreateProductAggregation - failed to discard txn", zap.String("reason", err.Error()))
// 				retErr = fmt.Errorf("dgraph/CreateProductAggregation - cannot discard txn")
// 			}
// 			return
// 		}
// 		if err := txn.Commit(ctx); err != nil {
// 			logger.Log.Error("dgraph/CreateProductAggregation - failed to commit txn", zap.String("reason", err.Error()))
// 			retErr = fmt.Errorf("dgraph/CreateProductAggregation - cannot commit txn")
// 		}
// 	}()

// 	assigned, err := txn.Mutate(ctx, mu)
// 	if err != nil {
// 		logger.Log.Error("dgraph/CreateProductAggregation - failed to create aggregation", zap.String("reason", err.Error()), zap.Any("aggregation", pa))
// 		return nil, errors.New("cannot create aggregation")
// 	}
// 	id, ok := assigned.Uids[pa.Name]
// 	if !ok {
// 		logger.Log.Error("dgraph/CreateProductAggregation - failed to create aggregation", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("aggregation", pa))
// 		return nil, errors.New("cannot create aggregation")
// 	}
// 	pa.ID = id
// 	return pa, nil
// }

// // ProductAggregationsByName implements Licence ProductAggregationsByName function
// func (l *LicenseRepository) ProductAggregationsByName(ctx context.Context, name string, scopes []string) (*v1.ProductAggregation, error) {

// 	variables := make(map[string]string)

// 	variables["$name"] = name

// 	q := `  query ProductAggByName($name:string) {
// 		Aggregations(func:eq(product_aggregation.name,$name))` + agregateFilters(scopeFilters(scopes)) + ` {
// 		  ID:uid
// 		  Name: product_aggregation.name
// 		  Editor: product_aggregation.editor
// 		  Product:product_aggregation.product_name
// 		  Metric:product_aggregation.metric{
// 			  MID:uid
// 			  Name:metric.name
// 		  }
// 		  Products:product_aggregation.products{
// 			  PID:product.swidtag
// 			  ProductName:product.name
// 		  }
// 	  }
//    }

// 	 `

// 	resp, err := l.dg.NewTxn().QueryWithVars(ctx, q, variables)
// 	if err != nil {
// 		logger.Log.Error("ProductAggregationsByName - ", zap.String("reason", err.Error()), zap.String("query", q))
// 		return nil, errors.New("productAggregationsByName - cannot complete query transaction")
// 	}

// 	type Data struct {
// 		// Aggregations []*v1.ProductAggregation
// 		Aggregations []struct {
// 			ID      string
// 			Name    string
// 			Editor  string
// 			Product string
// 			Metric  []struct {
// 				MID  string
// 				Name string
// 			}
// 			Products []struct {
// 				PID         string
// 				ProductName string
// 			}
// 		}
// 	}
// 	var aggList Data
// 	if err := json.Unmarshal(resp.GetJson(), &aggList); err != nil {
// 		logger.Log.Error("ProductAggregationsByName - ", zap.String("reason", err.Error()), zap.String("query", q))
// 		return nil, errors.New("productAggregationsByName - cannot unmarshal Json object")
// 	}
// 	if len(aggList.Aggregations) == 0 {
// 		return nil, v1.ErrNodeNotFound
// 	}

// 	prodAgg := &v1.ProductAggregation{}

// 	prodAgg.ID = aggList.Aggregations[0].ID
// 	prodAgg.Name = aggList.Aggregations[0].Name
// 	prodAgg.Editor = aggList.Aggregations[0].Editor
// 	prodAgg.Product = aggList.Aggregations[0].Product

// 	if len(aggList.Aggregations[0].Metric) > 0 {
// 		prodAgg.Metric = aggList.Aggregations[0].Metric[0].MID
// 		prodAgg.MetricName = aggList.Aggregations[0].Metric[0].Name
// 	}
// 	prodAgg.Products = make([]string, len(aggList.Aggregations[0].Products))
// 	prodAgg.ProductsFull = make([]*v1.ProductData, len(aggList.Aggregations[0].Products))
// 	for j := range aggList.Aggregations[0].Products {
// 		prodAgg.Products[j] = aggList.Aggregations[0].Products[j].PID
// 		prodAgg.ProductsFull[j] = &v1.ProductData{
// 			Swidtag: aggList.Aggregations[0].Products[j].PID,
// 			Name:    aggList.Aggregations[0].Products[j].ProductName,
// 		}
// 	}

// 	return prodAgg, nil
// }

func uidNotFilter(uids []string) string {
	if len(uids) == 0 {
		return ""
	}
	filters := make([]string, len(uids))
	for i := range uids {
		filters[i] = " NOT uid( " + uids[i] + ")"
	}
	return "@filter( " + strings.Join(filters, " AND ") + " )"
}

func aggQueryFromFilterWithID(uid, id string, filter *v1.AggregateFilter) string {
	if filter == nil && len(filter.Filters) == 0 {
		return ""
	}
	return id + ` as var(func:uid(` + uid + `))@cascade{
		~aggregation.products {
			aggregation.metric` + aggFilter(filter) + `
		}
	  }`
}

// ProductIDForSwidtag implements Licence ProductIDForSwidtag function
func (l *LicenseRepository) ProductIDForSwidtag(ctx context.Context, id string, params *v1.QueryProducts, scopes ...string) (string, error) {
	variables := make(map[string]string)

	variables["$id"] = id
	uids := []string{}
	aggQuery := ""
	if params == nil {
		params = &v1.QueryProducts{}
	}
	if params.AggFilter != nil && len(params.AcqFilter.Filters) != 0 {
		uids = append(uids, "IID_AGG")
		aggQuery = aggQueryFromFilterWithID("IID", "IID_AGG", params.AggFilter)
	}

	q := `   query ProductByID($id:string){
		IID as var(func:eq(product.swidtag,$id))@cascade ` + agregateFilters(scopeFilters(scopes), productFilter(params.Filter)) + ` {
			` + acqFilter(params.AcqFilter) + `
	  }

	  ` + aggQuery + `

	  Products(func: uid(IID))@cascade` + uidNotFilter(uids) + `{
		  ID: uid
	  }
   }

	 `
	fmt.Println(q)

	resp, err := l.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("ProductIDForSwidtag - ", zap.String("reason", err.Error()), zap.String("query", q))
		return "", errors.New("productIDForSwidtag - cannot complete query transaction")
	}

	type Data struct {
		Products []struct {
			ID string
		}
	}
	var prodList Data
	if err := json.Unmarshal(resp.GetJson(), &prodList); err != nil {
		logger.Log.Error("ProductIDForSwidtag - ", zap.String("reason", err.Error()), zap.String("query", q))
		return "", errors.New("productIDForSwidtag - cannot unmarshal Json object")
	}
	if len(prodList.Products) == 0 {
		return "", v1.ErrNodeNotFound
	}

	return prodList.Products[0].ID, nil
}

// func scopeNquad(scope, uid string) []*api.NQuad {
// 	return []*api.NQuad{
// 		{
// 			Subject:     uid,
// 			Predicate:   "scopes",
// 			ObjectValue: stringObjectValue(scope),
// 		},
// 	}
// }

// func productNquad(pID, uid string) []*api.NQuad {
// 	return []*api.NQuad{
// 		{
// 			Subject:   uid,
// 			Predicate: "product_aggregation.products",
// 			ObjectId:  pID,
// 		},
// 	}
// }

// func productsNquad(prod []string, blankID string) []*api.NQuad {
// 	nquads := []*api.NQuad{}
// 	for _, pID := range prod {
// 		nquads = append(nquads, productNquad(pID, blankID)...)
// 	}
// 	return nquads
// }

// func scopesNquad(scp []string, blankID string) []*api.NQuad {
// 	nquads := []*api.NQuad{}
// 	for _, sID := range scp {
// 		nquads = append(nquads, scopeNquad(sID, blankID)...)
// 	}
// 	return nquads
// }
