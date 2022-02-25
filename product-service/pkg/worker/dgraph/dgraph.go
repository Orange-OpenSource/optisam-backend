package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// Worker ...
type Worker struct {
	id string
	dg *dgo.Dgraph
}

// MessageType ...
type MessageType string

const (
	// UpsertProductRequest is request to upsert products in dgraph
	UpsertProductRequest MessageType = "UpsertProduct"
	// UpsertAcqRightsRequest is request to upsert acqrights in dgraph
	UpsertAcqRights MessageType = "UpsertAcqRights"
	// UpsertAggregation is request to upsert aggregation in dgraph
	UpsertAggregation MessageType = "UpsertAggregation"
	// DeleteAggregation is request to delete aggregation from dgraph
	DeleteAggregation MessageType = "DeleteAggregation"
	// DropProductDataRequest is request to drop complete products, acquired rights,
	// aggregation, editors, linked applications,linked equipments of a particular scope from Dgraph
	DropProductDataRequest MessageType = "DropProductData"
	// DeleteAcqrightRequest is request to delete acqright from dgraph
	DeleteAcqright MessageType = "DeleteAcqright"
)

// Envelope ...
type Envelope struct {
	Type MessageType `JSON:"message_type"`
	JSON json.RawMessage
}

// NewWorker ...
func NewWorker(id string, dg *dgo.Dgraph) *Worker {
	return &Worker{id: id, dg: dg}
}

// ID gives worker id
func (w *Worker) ID() string {
	return w.id
}

// DoWork will load products/linked applications,linked equipments data into Dgraph
// nolint: funlen, gocyclo
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	mu.Lock()
	defer mu.Unlock()
	var e Envelope
	var updatePartialFlag bool
	_ = json.Unmarshal(j.Data, &e)
	var queries []string
	switch e.Type {
	case UpsertProductRequest:
		queries = append(queries, "query", "{")
		var upr v1.UpsertProductRequest
		_ = json.Unmarshal(e.JSON, &upr)
		var mutations []*api.Mutation
		query := `var(func: eq(product.swidtag,"` + upr.GetSwidTag() + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + upr.Scope + `")){
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
		// re-establish the link between product and acqRights
		query = `var(func: eq(acqRights.swidtag,"` + upr.GetSwidTag() + `")) @filter(eq(type_name,"acqRights") AND eq(scopes,"` + upr.Scope + `")){
			acqright as uid
		}`
		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(acqright),0))`,
			SetNquads: []byte(`
				uid(product)  <product.acqRights> uid(acqright) .
			`),
			CommitNow: true,
		})
		if upr.GetOptionOf() != "" {
			query = `var(func: eq(product.swidtag,"` + upr.GetOptionOf() + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + upr.Scope + `")){
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
					query = `var(func: eq(application.id,"` + app + `")) @filter(eq(type_name,"application") AND eq(scopes,"` + upr.Scope + `")){
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
					query = `var(func: eq(equipment.id,"` + equipUser.GetEquipmentId() + `")) @filter(eq(type_name,"equipment") AND eq(scopes,"` + upr.Scope + `")){
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
						query = `var(func: eq(users.id,"` + userID + `")) @filter(eq(type_name,"users") AND eq(scopes,"` + upr.Scope + `")){
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
			query = `var(func: eq(editor.name,"` + upr.GetEditor() + `")) @filter(eq(type_name,"editor") AND eq(scopes,"` + upr.Scope + `")){
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
		// fmt.Println(q)
		req := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query), zap.Any("mutation", req.Mutations))
			return errRetry
		}
	case UpsertAcqRights:
		var uar UpsertAcqRightsRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(acqRights.SKU,"` + uar.Sku + `")) @filter(eq(type_name,"acqRights") AND eq(scopes,"` + uar.Scope + `")){
				acRights as uid `
		if uar.IsSwidtagModified {
			query += `	
					oldSwidtag as acqRights.swidtag 
				`
		}
		// if uar.IsMetricModifed {
		// 	query += `
		// 			oldMetrics as acqRights.metric
		// 		`
		// }
		query += `}
			var(func: eq(product.swidtag,"` + uar.Swidtag + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + uar.Scope + `") ){
				product as uid
				}
			`
		if uar.IsSwidtagModified {
			query += `
				var(func: eq(product.swidtag,val(oldSwidtag))) @filter(eq(type_name,"product") AND eq(scopes,"` + uar.Scope + `") ){
					oldProduct as uid
				}
				`
		}
		query += `}`

		var mutations []*api.Mutation
		mutations = append(mutations, &api.Mutation{
			Cond: "@if(eq(len(acRights),0))",
			SetNquads: []byte(`
			uid(acRights) <acqRights.SKU> "` + uar.Sku + `" .
			uid(acRights) <type_name> "acqRights" .
			uid(acRights) <dgraph.type> "AcquiredRights" .
			uid(acRights) <scopes> "` + uar.Scope + `" .
			`),
		})

		mutations = append(mutations, &api.Mutation{
			Cond: "@if(eq(len(product),0))",
			SetNquads: []byte(`
			uid(product) <product.swidtag> "` + uar.Swidtag + `" .
			uid(product) <type_name> "product" .
			uid(product) <dgraph.type> "Product" .
			uid(product) <scopes> "` + uar.Scope + `" .
			`),
		})
		//	log.Println("add new product", string(mutations[1].SetNquads))

		if uar.IsSwidtagModified {
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
			uid(oldProduct) <product.acqRights> uid(acRights) .
			`),
			})

			//	log.Println("delete link with old swidtag", string(mutations[2].DelNquads))
		}
		if uar.IsMetricModifed {
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
			uid(acRights) <acqRights.metric> * .
			`),
			})
		}
		set := `
			uid(acRights) <acqRights.swidtag> "` + uar.Swidtag + `" .
			uid(acRights) <acqRights.productName> "` + uar.ProductName + `" .
			uid(acRights) <acqRights.editor> "` + uar.ProductEditor + `" .
			uid(acRights) <acqRights.numOfAcqLicences> "` + strconv.Itoa(int(uar.NumLicensesAcquired)) + `" .
			uid(acRights) <acqRights.averageUnitPrice> "` + strconv.Itoa(int(uar.AvgUnitPrice)) + `" .
			uid(acRights) <acqRights.averageMaintenantUnitPrice> "` + strconv.Itoa(int(uar.AvgMaintenanceUnitPrice)) + `" .
			uid(acRights) <acqRights.totalPurchaseCost> "` + strconv.Itoa(int(uar.TotalPurchaseCost)) + `" .
			uid(acRights) <acqRights.totalMaintenanceCost> "` + strconv.Itoa(int(uar.TotalMaintenanceCost)) + `" .
			uid(acRights) <acqRights.totalCost> "` + strconv.Itoa(int(uar.TotalCost)) + `" .
			uid(acRights) <acqRights.startOfMaintenance> "` + uar.StartOfMaintenance + `" .
			uid(acRights) <acqRights.endOfMaintenance> "` + uar.EndOfMaintenance + `" .
			uid(product) <product.swidtag> "` + uar.Swidtag + `" .
			uid(product) <product.acqRights> uid(acRights) .
		`
		reqmetrics := strings.Split(uar.MetricType, ",")
		for _, met := range reqmetrics {
			set += `
				uid(acRights) <acqRights.metric> "` + met + `" .
			`
		}
		mutations = append(mutations, &api.Mutation{
			SetNquads: []byte(set),
		})

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
		var uar UpsertAggregatedRightsRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(aggregation.name,"` + uar.Name + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,"` + uar.Scope + `") ){
				aggregation as uid
				}
			
			`
		set := `
		uid(aggregation) <aggregation.id> "` + strconv.Itoa(int(uar.ID)) + `" .
		uid(aggregation) <aggregation.name> "` + uar.Name + `" .
		uid(aggregation) <aggregation.SKU> "` + uar.Sku + `" .
		uid(aggregation) <type_name> "aggregation" .
		uid(aggregation) <dgraph.type> "Aggregation" .
		uid(aggregation) <scopes> "` + uar.Scope + `" .
		uid(aggregation) <aggregation.editor> "` + uar.ProductEditor + `" .
		uid(aggregation) <aggregation.numOfAcqLicences> "` + strconv.Itoa(int(uar.NumLicensesAcquired)) + `" .
		uid(aggregation) <aggregation.averageUnitPrice> "` + strconv.Itoa(int(uar.AvgUnitPrice)) + `" .
		uid(aggregation) <aggregation.averageMaintenanceUnitPrice> "` + strconv.Itoa(int(uar.AvgMaintenanceUnitPrice)) + `" .
		uid(aggregation) <aggregation.totalPurchaseCost> "` + strconv.Itoa(int(uar.TotalPurchaseCost)) + `" .
		uid(aggregation) <aggregation.totalMaintenanceCost> "` + strconv.Itoa(int(uar.TotalMaintenanceCost)) + `" .
		uid(aggregation) <aggregation.totalCost> "` + strconv.Itoa(int(uar.TotalCost)) + `" .
		uid(aggregation) <aggregation.startOfMaintenance> "` + uar.StartOfMaintenance + `" .
		uid(aggregation) <aggregation.endOfMaintenance> "` + uar.EndOfMaintenance + `" .
		`
		reqmetrics := strings.Split(uar.Metric, ",")
		for _, met := range reqmetrics {
			set += `
				uid(aggregation) <aggregation.metric> "` + met + `" .
			`
		}
		for _, prodname := range uar.Products {
			set += `
				uid(aggregation) <aggregation.product_names> "` + prodname + `" .
			`
		}
		for i, product := range uar.Swidtags {
			query += `
			var(func: eq(product.swidtag,"` + product + `")) @filter(eq(type_name,"product")AND eq(scopes,"` + uar.Scope + `") ){
				product` + strconv.Itoa(i) + ` as uid
			}
			`
			set += `
			uid(aggregation) <aggregation.products> uid(product` + strconv.Itoa(i) + `) .
			uid(aggregation) <aggregation.swidtags> "` + product + `" .
			`
		}
		query += `}`
		logger.Log.Info(query)
		muUpsert := &api.Mutation{
			DelNquads: []byte(`
			uid(aggregation) * * .
			`),
			SetNquads: []byte(set),
		}
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
			var(func: eq(aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,` + dar.Scope + `) ){
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
		var dar v1.DropProductDataRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query { 
			  `
		if dar.DeletionType == v1.DropProductDataRequest_PARK || dar.DeletionType == v1.DropProductDataRequest_FULL { //nolint
			query += `
					productType as var(func: type(Product)) @filter(eq(scopes,` + dar.Scope + `)){
					products as product.swidtag
					productEquipments as product.equipment
					productEditors as product.editor
					productApplications as ~product.application
				}
				editorType as var(func: type(Editor)) @filter(eq(scopes,` + dar.Scope + `)){
					editors as editor.name
				}
				aggregationType as var(func: type(Aggregation)) @filter(eq(scopes,` + dar.Scope + `)){
					aggregations as aggregation.id
				}
				var(func: type(User)) @filter(eq(scopes,` + dar.Scope + `)){
					userId as  uid
				}
				`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL { //nolint
			query += ` 
				acqrightsType as var(func: type(AcquiredRights)) @filter(eq(scopes,` + dar.Scope + `)){
				acqrights as acqRights.SKU
			}
			`
		}
		query += `}`

		delete := ``
		if dar.DeletionType == v1.DropProductDataRequest_PARK || dar.DeletionType == v1.DropProductDataRequest_FULL {
			delete += `
				uid(userId) * * .
				uid(productType) * * .
				uid(products) * * .
				uid(productEditors) * * .
				uid(productEquipments) * * .
				uid(productApplications) * * .
				uid(editorType) * * .
				uid(editors) * * .
				uid(aggregationType) * * .
				uid(aggregations) * * .
			`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL {
			delete += `
					uid(acqrightsType) * * .
					uid(acqrights) * * .
					`
		}
		set := ``
		if dar.DeletionType == v1.DropProductDataRequest_PARK || dar.DeletionType == v1.DropProductDataRequest_FULL {
			set += `
				uid(userId) <Recycle> "true" .
				uid(productType) <Recycle> "true" .
				uid(products) <Recycle> "true" .
				uid(productEditors) <Recycle> "true" .
				uid(productEquipments) <Recycle> "true" .
				uid(productApplications) <Recycle> "true" .
				uid(editorType) <Recycle> "true" .
				uid(editors) <Recycle> "true" .
				uid(aggregationType) <Recycle> "true" .
				uid(aggregations) <Recycle> "true" .
		`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL {
			set += `
			uid(acqrightsType) <Recycle> "true" .
			uid(acqrights) <Recycle> "true" .
			`
		}
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query, zap.Any("set", set), zap.Any("delete", delete))
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to drop products data from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeleteAcqright:
		var dar DeleteAcqRightRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
			var(func: eq(acqRights.SKU,"` + dar.Sku + `")) @filter(eq(scopes,` + dar.Scope + `)){
				acqright as uid
				}
			}
			`
		delete := `
				uid(acqright) * * .
		`
		set := `
				uid(acqright) <Recycle> "true" .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query)
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to delete acqight from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	default:
		fmt.Println(e.JSON)
	}

	// Everything's fine, we're done here
	return nil
}
