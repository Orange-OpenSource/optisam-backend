// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "optisam-backend/acqrights-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	"strconv"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type Worker struct {
	id string
	dg *dgo.Dgraph
}

type MessageType string

const (
	UpsertAcqRightsRequest MessageType = "UpsertAcqRights"
	UpsertAggregation      MessageType = "UpsertAggregation"
	DeleteAggregation      MessageType = "DeleteAggregation"
)

type Envelope struct {
	Type MessageType `json:"message_type"`
	JSON json.RawMessage
}

func NewWorker(id string, dg *dgo.Dgraph) *Worker {
	return &Worker{id: id, dg: dg}
}

func (w *Worker) ID() string {
	return w.id
}

//Dowork will load products/linked applications,linked equipments data into Dgraph
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	var e Envelope
	_ = json.Unmarshal(j.Data, &e)
	logger.Log.Info("Operation Type", zap.String("message_type", string(e.Type)))
	switch e.Type {
	case UpsertAcqRightsRequest:
		var uar v1.UpsertAcqRightsRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(acqRights.SKU,"` + uar.GetSku() + `")) @filter(eq(type_name,"acqRights")){
				acRights as uid
				}
			var(func: eq(product.swidtag,"` + uar.GetSwidtag() + `")) @filter(eq(type_name,"product")){
				product as uid
				}
			}
			`
		upsertAcqRights := `
			uid(acRights) <acqRights.SKU> "` + uar.GetSku() + `" .
			uid(acRights) <type_name> "acqRights" .
			uid(acRights) <dgraph.type> "AcquiredRights" .
			uid(acRights) <scopes> "` + uar.GetScope() + `" .
			uid(acRights) <acqRights.swidtag> "` + uar.GetSwidtag() + `" .
			uid(acRights) <acqRights.productName> "` + uar.GetProductName() + `" .
			uid(acRights) <acqRights.editor> "` + uar.GetProductEditor() + `" .
			uid(acRights) <acqRights.entity> "` + uar.GetEntity() + `" .
			uid(acRights) <acqRights.metric> "` + uar.GetMetricType() + `" .
			uid(acRights) <acqRights.numOfAcqLicences> "` + strconv.Itoa(int(uar.GetNumLicensesAcquired())) + `" .
			uid(acRights) <acqRights.numOfLicencesUnderMaintenance> "` + strconv.Itoa(int(uar.GetNumLicencesMaintainance())) + `" .
			uid(acRights) <acqRights.averageUnitPrice> "` + strconv.Itoa(int(uar.GetAvgUnitPrice())) + `" .
			uid(acRights) <acqRights.averageMaintenantUnitPrice> "` + strconv.Itoa(int(uar.GetAvgMaintenanceUnitPrice())) + `" .
			uid(acRights) <acqRights.totalPurchaseCost> "` + strconv.Itoa(int(uar.GetTotalPurchaseCost())) + `" .
			uid(acRights) <acqRights.totalMaintenanceCost> "` + strconv.Itoa(int(uar.GetTotalMaintenanceCost())) + `" .
			uid(acRights) <acqRights.totalCost> "` + strconv.Itoa(int(uar.GetTotalCost())) + `" .
			uid(product) <product.swidtag> "` + uar.GetSwidtag() + `" .
			uid(product) <product.acqRights> uid(acRights) .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			uid(product) <scopes> "` + uar.GetScope() + `" .
		`

		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{{SetNquads: []byte(upsertAcqRights)}},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case UpsertAggregation:
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
			var(func: eq(product_aggregation.name,"` + uar.GetName() + `")) @filter(eq(type_name,"product_aggregation")){
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
			query += `
			var(func: eq(product.swidtag,"` + product + `")) @filter(eq(type_name,"product")){
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
			var(func: eq(product_aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `")) @filter(eq(type_name,"product_aggregation")){
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
	default:
		fmt.Println(e.JSON)
	}

	//Everything's fine, we're done here
	return nil
}
