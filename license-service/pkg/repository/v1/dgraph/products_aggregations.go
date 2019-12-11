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
	"strings"

	"github.com/dgraph-io/dgo/protos/api"
	"go.uber.org/zap"
)

// CreateProductAggregation implements Licence CreateProductAggregation function
func (r *LicenseRepository) CreateProductAggregation(ctx context.Context, pa *v1.ProductAggregation, scopes []string) (retPa *v1.ProductAggregation, retErr error) {
	blankID := blankID(pa.Name)
	nquads := []*api.NQuad{
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "type",
			ObjectValue: stringObjectValue("product_aggregation"),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "product_aggreagtion.editor",
			ObjectValue: stringObjectValue(pa.Editor),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "product_aggreagtion.product_name",
			ObjectValue: stringObjectValue(pa.Product),
		},
		&api.NQuad{
			Subject:     blankID,
			Predicate:   "product_aggreagtion.name",
			ObjectValue: stringObjectValue(pa.Name),
		},
		&api.NQuad{
			Subject:   blankID,
			Predicate: "product_aggreagtion.metric",
			ObjectId:  pa.Metric,
		},
	}

	nquads = append(nquads, productsNquad(pa.Products, blankID)...)
	nquads = append(nquads, scopesNquad(scopes, blankID)...)

	mu := &api.Mutation{
		Set: nquads,
		//	CommitNow: true,
	}
	txn := r.dg.NewTxn()

	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/CreateProductAggregation - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/CreateProductAggregation - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/CreateProductAggregation - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/CreateProductAggregation - cannot commit txn")
		}
	}()

	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/CreateProductAggregation - failed to create aggregation", zap.String("reason", err.Error()), zap.Any("aggregation", pa))
		return nil, errors.New("cannot create aggregation")
	}
	id, ok := assigned.Uids[pa.Name]
	if !ok {
		logger.Log.Error("dgraph/CreateProductAggregation - failed to create aggregation", zap.String("reason", "cannot find id in assigned Uids map"), zap.Any("aggregation", pa))
		return nil, errors.New("cannot create aggregation")
	}
	pa.ID = id
	return pa, nil
}

// ListProductAggregations implements Licence ListProductAggregations function
func (r *LicenseRepository) ListProductAggregations(ctx context.Context, scopes []string) ([]*v1.ProductAggregation, error) {
	q := `   {
		Aggregations(func:eq(type,"product_aggregation"))` + agregateFilters(scopeFilters(scopes)) + ` {
		  ID:uid
		  Name: product_aggreagtion.name
		  Editor: product_aggreagtion.editor
		  Product:product_aggreagtion.product_name
		  Metric:product_aggreagtion.metric{
			  MID: uid
		  }
		  Products:product_aggreagtion.products{
			  PID: product.swidtag
		  }
	  }
   }

	 `
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListProductAggregations - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListProductAggregations - cannot complete query transaction")
	}

	type Data struct {
		//Aggregations []*v1.ProductAggregation
		Aggregations []struct {
			ID      string
			Name    string
			Editor  string
			Product string
			Metric  []struct {
				MID string
			}
			Products []struct {
				PID string
			}
		}
	}
	var aggList Data
	if err := json.Unmarshal(resp.GetJson(), &aggList); err != nil {
		logger.Log.Error("ListProductAggregations - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListProductAggregations - cannot unmarshal Json object")
	}

	prodAgg := make([]*v1.ProductAggregation, len(aggList.Aggregations))
	aggs := aggList.Aggregations
	for i := range aggs {
		prodAgg[i] = &v1.ProductAggregation{}
		prodAgg[i].ID = aggs[i].ID
		prodAgg[i].Name = aggs[i].Name
		prodAgg[i].Editor = aggs[i].Editor
		prodAgg[i].Product = aggs[i].Product

		if len(aggs[i].Metric) > 0 {
			prodAgg[i].Metric = aggs[i].Metric[0].MID
		}
		prodAgg[i].Products = make([]string, len(aggs[i].Products))
		for j := range aggs[i].Products {
			prodAgg[i].Products[j] = aggs[i].Products[j].PID
		}

	}

	return prodAgg, nil
}

// ProductAggregationsByName implements Licence ProductAggregationsByName function
func (r *LicenseRepository) ProductAggregationsByName(ctx context.Context, name string, scopes []string) (*v1.ProductAggregation, error) {

	variables := make(map[string]string)

	variables["$name"] = name

	q := `  query ProductAggByName($name:string) {
		Aggregations(func:eq(product_aggreagtion.name,$name))` + agregateFilters(scopeFilters(scopes)) + ` {
		  ID:uid
		  Name: product_aggreagtion.name
		  Editor: product_aggreagtion.editor
		  Product:product_aggreagtion.product_name
		  Metric:product_aggreagtion.metric{
			  MID:uid
		  }
		  Products:product_aggreagtion.products{
			  PID:uid
		  }
	  }
   }

	 `

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("ProductAggregationsByName - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ProductAggregationsByName - cannot complete query transaction")
	}

	type Data struct {
		//Aggregations []*v1.ProductAggregation
		Aggregations []struct {
			ID      string
			Name    string
			Editor  string
			Product string
			Metric  []struct {
				MID string
			}
			Products []struct {
				PID string
			}
		}
	}
	var aggList Data
	if err := json.Unmarshal(resp.GetJson(), &aggList); err != nil {
		logger.Log.Error("ProductAggregationsByName - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ProductAggregationsByName - cannot unmarshal Json object")
	}
	if len(aggList.Aggregations) == 0 {
		return nil, v1.ErrNodeNotFound
	}

	prodAgg := &v1.ProductAggregation{}

	prodAgg.ID = aggList.Aggregations[0].ID
	prodAgg.Name = aggList.Aggregations[0].Name
	prodAgg.Editor = aggList.Aggregations[0].Editor
	prodAgg.Product = aggList.Aggregations[0].Product

	if len(aggList.Aggregations[0].Metric) > 0 {
		prodAgg.Metric = aggList.Aggregations[0].Metric[0].MID
	}
	prodAgg.Products = make([]string, len(aggList.Aggregations[0].Products))
	for j := range aggList.Aggregations[0].Products {
		prodAgg.Products[j] = aggList.Aggregations[0].Products[j].PID
	}

	return prodAgg, nil
}

// DeleteProductAggregation implements Licence DeleteProductAggregation function
func (r *LicenseRepository) DeleteProductAggregation(ctx context.Context, id string, scopes []string) (retPa []*v1.ProductAggregation, retErr error) {

	if err := deleteProductAgg(r, ctx, id); err != nil {
		return nil, err
	}
	prodAgg, err := r.ListProductAggregations(ctx, scopes)
	if err != nil {
		return nil, errors.New("DeleteProductAggregation - ListProductAggregations - cannot fetch product aggregations")
	}
	return prodAgg, nil
}

func deleteProductAgg(r *LicenseRepository, ctx context.Context, id string) (retErr error) {

	d := map[string]string{"uid": id}
	pb, err := json.Marshal(d)
	if err != nil {
		return errors.New("deleteProductAgg - cannot marshal Json object")
	}

	mu := &api.Mutation{
		//	CommitNow: true,
		//DeleteJson: []byte(`{"uid": "` + id + `"}`),
		DeleteJson: pb,
	}
	txn := r.dg.NewTxn()
	defer func() {
		if retErr != nil {
			if err := txn.Discard(ctx); err != nil {
				logger.Log.Error("dgraph/deleteProductAgg - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf("dgraph/deleteProductAgg - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(ctx); err != nil {
			logger.Log.Error("dgraph/deleteProductAgg - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf("dgraph/deleteProductAgg - cannot commit txn")
		}
	}()
	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		logger.Log.Error("dgraph/deleteProductAgg - failed to delete aggregation", zap.String("reason", err.Error()), zap.Any("aggregation id", id))
		return errors.New("cannot delete aggregation")
	}
	return nil
}

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
		~product_aggreagtion.products {
			product_aggreagtion.metric` + aggFilter(filter) + `
		}
	  }`
}

// ProductIDForSwidtag implements Licence ProductIDForSwidtag function
func (r *LicenseRepository) ProductIDForSwidtag(ctx context.Context, id string, params *v1.QueryProducts, scopes []string) (string, error) {
	variables := make(map[string]string)

	variables["$id"] = id
	uids := []string{}
	aggQuery := ""
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

	resp, err := r.dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		logger.Log.Error("ProductIDForSwidtag - ", zap.String("reason", err.Error()), zap.String("query", q))
		return "", errors.New("ProductIDForSwidtag - cannot complete query transaction")
	}

	type Data struct {
		Products []struct {
			ID string
		}
	}
	var prodList Data
	if err := json.Unmarshal(resp.GetJson(), &prodList); err != nil {
		logger.Log.Error("ProductIDForSwidtag - ", zap.String("reason", err.Error()), zap.String("query", q))
		return "", errors.New("ProductIDForSwidtag - cannot unmarshal Json object")
	}
	if len(prodList.Products) == 0 {
		return "", v1.ErrNodeNotFound
	}

	return prodList.Products[0].ID, nil
}

func scopeNquad(scope, uid string) []*api.NQuad {
	return []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
	}
}

func productNquad(pID, uid string) []*api.NQuad {
	return []*api.NQuad{
		&api.NQuad{
			Subject:   uid,
			Predicate: "product_aggreagtion.products",
			ObjectId:  pID,
		},
	}
}

func productsNquad(prod []string, blankID string) []*api.NQuad {
	nquads := []*api.NQuad{}
	for _, pID := range prod {
		nquads = append(nquads, productNquad(pID, blankID)...)
	}
	return nquads
}

func scopesNquad(scp []string, blankID string) []*api.NQuad {
	nquads := []*api.NQuad{}
	for _, sID := range scp {
		nquads = append(nquads, scopeNquad(sID, blankID)...)
	}
	return nquads
}
