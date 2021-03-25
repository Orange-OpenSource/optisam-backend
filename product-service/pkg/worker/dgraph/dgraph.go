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
	"log"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"strconv"
	"strings"
	"sync"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

var errRetry = errors.New("RETRY")
var mu sync.Mutex

//Worker ...
type Worker struct {
	id string
	dg *dgo.Dgraph
}

// MessageType ...
type MessageType string

const (
	//UpsertProductRequest is request to upsert products in dgraph
	UpsertProductRequest MessageType = "UpsertProduct"
	//UpsertAcqRightsRequest is request to upsert acqrights in dgraph
	UpsertAcqRightsRequest MessageType = "UpsertAcqRights"
	//UpsertAggregation is request to upsert aggregation in dgraph
	UpsertAggregation MessageType = "UpsertAggregation"
	//DeleteAggregation is request to delete aggregation from dgraph
	DeleteAggregation MessageType = "DeleteAggregation"
	//DropProductDataRequest is request to drop complete products, acquired rights,
	//aggregation, editors, linked applications,linked equipments of a particular scope from Dgraph
	DropProductDataRequest MessageType = "DropProductData"
)

//Envelope ...
type Envelope struct {
	Type MessageType `JSON:"message_type"`
	JSON json.RawMessage
}

//NewWorker ...
func NewWorker(id string, dg *dgo.Dgraph) *Worker {
	return &Worker{id: id, dg: dg}
}

//ID gives worker id
func (w *Worker) ID() string {
	return w.id
}

//DoWork will load products/linked applications,linked equipments data into Dgraph
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	var updatePartialFlag bool
	_ = json.Unmarshal(j.Data, &e)
	var queries []string
	switch e.Type {
	case UpsertProductRequest:
		mu.Lock()
		defer mu.Unlock()
		queries = append(queries, "query", "{")
		var upr v1.UpsertProductRequest
		_ = json.Unmarshal(e.JSON, &upr)
		var mutations []*api.Mutation
		query := `var(func: eq(product.swidtag,"` + upr.GetSwidTag() + `")) @filter(eq(type_name,"product")){
				product as uid
			}`
		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(eq(len(product),0))`,
			SetNquads: []byte(`
			uid(product) <product.swidtag> "` + upr.GetSwidTag() + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			uid(product) <scopes> "` + upr.GetScope() + `" .
			`),
			CommitNow: true,
		})
		if upr.GetOptionOf() != "" {
			query := `var(func: eq(product.swidtag,"` + upr.GetOptionOf() + `")) @filter(eq(type_name,"product")){
					child as uid
				}`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				Cond: `@if(eq(len(child),0))`,
				SetNquads: []byte(`
					uid(child) <product.swidtag> "` + upr.GetOptionOf() + `" .
					uid(child) <type_name> "product" .
					uid(child) <dgraph.type> "Product" .
					uid(child) <scopes> "` + upr.GetScope() + `" .
					`),
				CommitNow: true,
			})

			mutations = append(mutations, &api.Mutation{
				SetNquads: []byte(`uid(child) <product.child> uid(product) .`),
			})
		}
		// Application Upsert
		if len(upr.GetApplications().GetApplicationId()) > 0 {
			updatePartialFlag = true
			if upr.GetApplications().GetOperation() == "add" {

				for i, app := range upr.GetApplications().GetApplicationId() {
					appID := `application` + strconv.Itoa(i)
					query := `var(func: eq(application.id,"` + app + `")) @filter(eq(type_name,"application")){
					` + appID + ` as uid
					}
					`
					queries = append(queries, query)
					mutations = append(mutations, &api.Mutation{
						Cond: `@if(eq(len(` + appID + `),0))`,
						SetNquads: []byte(`
						uid(` + appID + `) <application.id> "` + app + `" .
						uid(` + appID + `) <type_name> "application" .
						uid(` + appID + `) <dgraph.type> "Application" .
						uid(` + appID + `) <scopes> "` + upr.GetScope() + `" .
						`),
						CommitNow: true,
					})
					mutations = append(mutations, &api.Mutation{
						SetNquads: []byte(`uid(` + appID + `) <application.product> uid(product) .`),
					})
				}

			}
		}
		// Equipments Upsert
		if len(upr.GetEquipments().GetEquipmentusers()) > 0 {
			updatePartialFlag = true
			if upr.GetEquipments().GetOperation() == "add" {
				for i, equipUser := range upr.GetEquipments().GetEquipmentusers() {
					eqUID := "equipment" + strconv.Itoa(i)
					query := `var(func: eq(equipment.id,"` + equipUser.GetEquipmentId() + `")) @filter(eq(type_name,"equipment")){
						` + eqUID + ` as uid
					}`
					queries = append(queries, query)
					mutations = append(mutations, &api.Mutation{
						Cond: `@if(eq(len(` + eqUID + `),0))`,
						SetNquads: []byte(`
						uid(` + eqUID + `) <equipment.id> "` + equipUser.GetEquipmentId() + `" .
						uid(` + eqUID + `) <type_name> "equipment" .
						uid(` + eqUID + `) <dgraph.type> "Equipment" .
						uid(` + eqUID + `) <scopes> "` + upr.GetScope() + `" .
							`),
						CommitNow: true,
					})

					mutations = append(mutations, &api.Mutation{
						SetNquads: []byte(`uid(product) <product.equipment> uid(` + eqUID + `) .`),
					})
					if equipUser.GetNumUser() > 0 {
						userID := `user_` + upr.GetSwidTag() + equipUser.GetEquipmentId()
						userUID := `users` + strconv.Itoa(i)
						query := `var(func: eq(users.id,"` + userID + `")) @filter(eq(type_name,"users")){
							` + userUID + ` as uid
						}`
						queries = append(queries, query)
						mutations = append(mutations, &api.Mutation{
							Cond: `@if(eq(len(` + userUID + `),0))`,
							SetNquads: []byte(`
							uid(` + userUID + `) <users.id> "` + userID + `" .
							uid(` + userUID + `) <type_name> "instance_users" .
							uid(` + userUID + `) <dgraph.type> "User" .
							uid(` + userUID + `) <scopes> "` + upr.GetScope() + `" .
							`),
							CommitNow: true,
						})

						mutations = append(mutations, &api.Mutation{
							SetNquads: []byte(`
							uid(product) <product.users> uid(` + userUID + `) .
							uid(` + eqUID + `) <equipment.users>  uid(` + userUID + `) .
							uid(` + userUID + `) <users.count> "` + strconv.Itoa(int(equipUser.GetNumUser())) + `" .
							`),
							CommitNow: true,
						})
					}
				}
			}
		}

		if !updatePartialFlag {
			query = `var(func: eq(editor.name,"` + upr.GetEditor() + `")) @filter(eq(type_name,"editor")){
			editor as uid
			}`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				Cond: "@if(eq(len(editor),0))",
				SetNquads: []byte(`
				uid(editor) <type_name> "editor" .
				uid(editor) <dgraph.type> "Editor" .
				uid(editor) <editor.name> "` + upr.GetEditor() + `" .
				uid(editor) <scopes> "` + upr.GetScope() + `" .
				`),
			})
			mutations = append(mutations, &api.Mutation{
				SetNquads: []byte(`
				uid(product) <product.name> "` + upr.GetName() + `" .
				uid(product) <product.version> "` + upr.GetVersion() + `" .
				uid(product) <product.category> "` + upr.GetCategory() + `" .
				uid(product) <product.editor> "` + upr.GetEditor() + `" .
				uid(editor) <editor.product> uid(product) .
				`),
			})
		}
		queries = append(queries, "}")
		q := strings.Join(queries, "\n")
		//fmt.Println(q)
		req := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query), zap.Any("mutation", req.Mutations))
			return errRetry
		}
	case UpsertAcqRightsRequest:
		mu.Lock()
		defer mu.Unlock()
		var uar v1.UpsertAcqRightsRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(acqRights.SKU,"` + uar.GetSku() + `")) @filter(eq(type_name,"acqRights") AND eq(scopes,"` + uar.GetScope() + `")){
				acRights as uid
				}
			var(func: eq(product.swidtag,"` + uar.GetSwidtag() + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + uar.GetScope() + `") ){
				product as uid
				}
			}
			`
		var mutations []*api.Mutation
		mutations = append(mutations, &api.Mutation{
			Cond: "@if(eq(len(acRights),0))",
			SetNquads: []byte(`
			uid(acRights) <acqRights.SKU> "` + uar.GetSku() + `" .
			uid(acRights) <type_name> "acqRights" .
			uid(acRights) <dgraph.type> "AcquiredRights" .
			uid(acRights) <scopes> "` + uar.GetScope() + `" .
			`),
		})
		log.Println("S!", string(mutations[0].SetNquads))

		mutations = append(mutations, &api.Mutation{
			Cond: "@if(eq(len(product),0))",
			SetNquads: []byte(`
			uid(product) <product.swidtag> "` + uar.GetSwidtag() + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			uid(product) <scopes> "` + uar.GetScope() + `" .
			`),
		})
		log.Println("S2", string(mutations[1].SetNquads))
		mutations = append(mutations, &api.Mutation{
			SetNquads: []byte(`
			uid(acRights) <acqRights.swidtag> "` + uar.GetSwidtag() + `" .
			uid(acRights) <acqRights.productName> "` + uar.GetProductName() + `" .
			uid(acRights) <acqRights.editor> "` + uar.GetProductEditor() + `" .
			uid(acRights) <acqRights.entity> "` + uar.GetEntity() + `" .
			uid(acRights) <acqRights.metric> "` + uar.GetMetricType() + `" .
			uid(acRights) <acqRights.numOfAcqLicences> "` + strconv.Itoa(int(uar.GetNumLicensesAcquired())) + `" .
			uid(acRights) <acqRights.averageUnitPrice> "` + strconv.Itoa(int(uar.GetAvgUnitPrice())) + `" .
			uid(acRights) <acqRights.averageMaintenantUnitPrice> "` + strconv.Itoa(int(uar.GetAvgMaintenanceUnitPrice())) + `" .
			uid(acRights) <acqRights.totalPurchaseCost> "` + strconv.Itoa(int(uar.GetTotalPurchaseCost())) + `" .
			uid(acRights) <acqRights.totalMaintenanceCost> "` + strconv.Itoa(int(uar.GetTotalMaintenanceCost())) + `" .
			uid(acRights) <acqRights.totalCost> "` + strconv.Itoa(int(uar.GetTotalCost())) + `" .
			uid(acRights) <acqRights.startOfMaintenance> "` + uar.GetStartOfMaintenance() + `" .
			uid(acRights) <acqRights.endOfMaintenance> "` + uar.GetEndOfMaintenance() + `" .
			uid(product) <product.swidtag> "` + uar.GetSwidtag() + `" .
			uid(product) <product.acqRights> uid(acRights) .
		`),
		})
		log.Println("S3", string(mutations[2].SetNquads))
		req := &api.Request{
			Query:     query,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query))
			return errors.New("RETRY")
		}
	case UpsertAggregation:
		// TODO: Do conditional upserts in this block also skipping this for now as bulk load is not needed\.
		var uar v1.ProductAggregationMessage
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(product_aggregation.name,"` + uar.GetName() + `")) @filter(eq(type_name,"product_aggregation")){
				aggregation as uid
				}
			var(func: eq(metric.name,"` + uar.GetMetric() + `")) @filter(eq(type_name,"metric")){
				metric as uid
			}
			`
		deleteQuery := `query {
			var(func: eq(product_aggregation.name,"` + uar.GetName() + `")) @filter(eq(type_name,"product_aggregation") ){
				aggregation as uid
				}
			}
			`
		delete := `
				uid(aggregation) <product_aggregation.products> * .
		`
		set := `
		uid(aggregation) <product_aggregation.id> "` + strconv.Itoa(int(uar.GetID())) + `" .
		uid(aggregation) <product_aggregation.name> "` + uar.GetName() + `" .
		uid(aggregation) <type_name> "product_aggregation" .
		uid(aggregation) <dgraph.type> "ProductAggregation" .
		uid(aggregation) <scopes> "` + uar.GetScope() + `" .
		uid(aggregation) <product_aggregation.editor> "` + uar.GetEditor() + `" .
		uid(aggregation) <product_aggregation.metric> uid(metric) .
		`
		for i, product := range uar.GetProducts() {
			//SCOPE BASED CHANGE
			query += `
			var(func: eq(product.swidtag,"` + product + `")) @filter(eq(type_name,"product")AND eq(scopes,"` + uar.GetScope() + `") ){
				product` + strconv.Itoa(i) + ` as uid
			}
			`
			set += `
			uid(aggregation) <product_aggregation.products> uid(product` + strconv.Itoa(i) + `) .
			`
		}
		query += `
		}`
		muDelete := &api.Mutation{DelNquads: []byte(delete)}
		logger.Log.Info(query)
		delreq := &api.Request{
			Query:     deleteQuery,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, delreq); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
		muUpsert := &api.Mutation{SetNquads: []byte(set)}
		logger.Log.Info(query)
		upsertreq := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muUpsert},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, upsertreq); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeleteAggregation:
		var dar v1.DeleteProductAggregationRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
			var(func: eq(product_aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `")) @filter(eq(type_name,"product_aggregation") ){
				aggregation as uid
				}
			`
		delete := `
				uid(aggregation) * * .
		`
		set := `
				uid(aggregation) <Recycle> "true" .
		`
		query += `
		}`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query)
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DropProductDataRequest:
		mu.Lock()
		defer mu.Unlock()
		var dar v1.DropProductDataRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
				productType as var(func: type(Product)) @filter(eq(scopes,` + dar.Scope + `)){
					products as product.swidtag
					productAcqrights as product.acqRights
					productEquipments as product.equipment
					productEditors as product.editor
					productApplications as ~product.application
				}
				acqrightsType as var(func: type(AcquiredRights)) @filter(eq(scopes,` + dar.Scope + `)){
					acqrights as acqRights.SKU
				}
				editorType as var(func: type(Editor)) @filter(eq(scopes,` + dar.Scope + `)){
					editors as editor.name
				}
				aggregationType as var(func: type(ProductAggregation)) @filter(eq(scopes,` + dar.Scope + `)){
					aggregations as product_aggregation.id
				}
			}
			`
		delete := `
				uid(productType) * * .
				uid(products) * * .
				uid(productAcqrights) * * .
				uid(productEditors) * * .
				uid(productEquipments) * * .
				uid(productApplications) * * .
				uid(acqrightsType) * * .
				uid(acqrights) * * .
				uid(editorType) * * .
				uid(editors) * * .
				uid(aggregationType) * * .
				uid(aggregations) * * .
		`
		set := `
				uid(productType) <Recycle> "true" .
				uid(products) <Recycle> "true" .
				uid(productAcqrights) <Recycle> "true" .
				uid(productEditors) <Recycle> "true" .
				uid(productEquipments) <Recycle> "true" .
				uid(productApplications) <Recycle> "true" .
				uid(acqrightsType) <Recycle> "true" .
				uid(acqrights) <Recycle> "true" .
				uid(editorType) <Recycle> "true" .
				uid(editors) <Recycle> "true" .
				uid(aggregationType) <Recycle> "true" .
				uid(aggregations) <Recycle> "true" .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query)
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to drop products data from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	default:
		fmt.Println(e.JSON)
	}

	//Everything's fine, we're done here
	return nil
}
