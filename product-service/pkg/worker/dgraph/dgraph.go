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
	"time"

	e_v1 "optisam-backend/equipment-service/pkg/api/v1"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var errRetry = errors.New("RETRY")
var mu sync.Mutex

// Worker ...
type Worker struct {
	id              string
	dg              *dgo.Dgraph
	equipmentClient e_v1.EquipmentServiceClient
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
	// aggregated rights, editors, linked applications,linked equipments of a particular scope from Dgraph
	DropProductDataRequest MessageType = "DropProductData"
	// DropAggregationData is request to drop complete aggregations of a particular scope from Dgraph
	DropAggregationData MessageType = "DropAggregationData"
	// DeleteAcqrightRequest is request to delete acqright from dgraph
	DeleteAcqright MessageType = "DeleteAcqright"
	// UpsertAggregationRights is to upsert aggregation rights in dgraph
	UpsertAggregatedRights MessageType = "UpsertAggregatedRights"
	// DeleteAggregationRights is to delete aggregation rights in dgraph
	DeleteAggregatedRights MessageType = "DeleteAggregatedRights"
	// UpsertNominativeUserRequest is request to upsert nominative user in dgraph
	UpsertNominativeUserRequest MessageType = "UpsertNominativeUser"
	// UpsertConcurrentUserRequest is request to upsert concurrent user in dgraph
	UpsertConcurrentUserRequest MessageType = "UpsertConcurrentUser"
	// DeleteConcurrentUserRequest is request to delete concurrent user in dgraph
	DeleteConcurrentUserRequest MessageType = "DeleteConcurrentUser"
	// DeleteNominativeUserRequest is request to delete nominative user in dgraph
	DeleteNominativeUserRequest MessageType = "DeleteNominativeUser"
	// DeletSaaSProductRequest is request to delete SaaS Product in dgraph
	DeletSaaSProductRequest MessageType = "DeletSaaSProduct"
)

// Envelope ...
type Envelope struct {
	Type MessageType `JSON:"message_type"`
	JSON json.RawMessage
}

// NewWorker ...
func NewWorker(id string, dg *dgo.Dgraph, grpcServers map[string]*grpc.ClientConn) *Worker {
	return &Worker{id: id, dg: dg, equipmentClient: e_v1.NewEquipmentServiceClient(grpcServers["equipment"])}
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
		// re-establish the link between product and aggregations
		query = `var(func: eq(aggregation.swidtags,[` + upr.GetSwidTag() + `])) @filter(eq(type_name,"aggregation") AND eq(scopes,"` + upr.Scope + `")){
			aggregation as uid
		}`
		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(aggregation),0))`,
			SetNquads: []byte(`
				uid(aggregation)  <aggregation.products> uid(product) .
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
					if equipUser.GetAllocatedUsers() > 0 {
						userID := `user_` + upr.GetSwidTag() + equipUser.GetEquipmentId()
						userUID := `users` + strconv.Itoa(i)
						query = `var(func: eq(users.id,"` + userID + `")) @filter(eq(type_name,"instance_users") AND eq(scopes,"` + upr.Scope + `")){
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
							uid(` + userUID + `) <users.count> "` + strconv.Itoa(int(equipUser.GetAllocatedUsers())) + `" .
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
			logger.Log.Error("Failed to upsert to Dgraph - UpsertProductRequest", zap.Error(err), zap.String("query", req.Query), zap.Any("mutation", req.Mutations))
			return errRetry
		}
		for _, v := range upr.GetEquipments().GetEquipmentusers() {
			_, err := w.equipmentClient.UpsertAllocMetricByFile(ctx, &e_v1.UpsertAllocMetricByFileRequest{
				Scope:            upr.GetScope(),
				Swidtag:          upr.GetSwidTag(),
				EquipmentId:      v.GetEquipmentId(),
				AllocatedMetrics: v.GetAllocatedMetrics(),
			})
			if err != nil {
				logger.Log.Error("Failed to allocate metric by file upload", zap.Error(err))
				return err
			}
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
			uid(acRights) <acqRights.numOfLicencesUnderMaintenance> "` + strconv.Itoa(int(uar.NumLicencesMaintenance)) + `" .
			uid(acRights) <acqRights.averageUnitPrice> "` + fmt.Sprintf("%.2f", uar.AvgUnitPrice) + `" .
			uid(acRights) <acqRights.averageMaintenantUnitPrice> "` + fmt.Sprintf("%.2f", uar.AvgMaintenanceUnitPrice) + `" .
			uid(acRights) <acqRights.totalPurchaseCost> "` + fmt.Sprintf("%.2f", uar.TotalPurchaseCost) + `" .
			uid(acRights) <acqRights.totalMaintenanceCost> "` + fmt.Sprintf("%.2f", uar.TotalMaintenanceCost) + `" .
			uid(acRights) <acqRights.totalCost> "` + fmt.Sprintf("%.2f", uar.TotalCost) + `" .
			uid(acRights) <acqRights.startOfMaintenance> "` + uar.StartOfMaintenance + `" .
			uid(acRights) <acqRights.endOfMaintenance> "` + uar.EndOfMaintenance + `" .
			uid(acRights) <acqRights.orderingDate> "` + uar.OrderingDate + `" .
			uid(acRights) <acqRights.corporateSourcingContract> "` + uar.CorporateSourcingContract + `" .
			uid(acRights) <acqRights.softwareProvider> "` + uar.SoftwareProvider + `" .
			uid(acRights) <acqRights.lastPurchasedOrder> "` + uar.LastPurchasedOrder + `" .
			uid(acRights) <acqRights.supportNumber> "` + uar.SupportNumber + `" .
			uid(acRights) <acqRights.maintenanceProvider> "` + uar.MaintenanceProvider + `" .
			uid(acRights) <acqRights.repartition> "` + strconv.FormatBool(uar.Repartition) + `" .
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
		var uar UpsertAggregationRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(aggregation.name,"` + uar.Name + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,"` + uar.Scope + `") ){
				aggregation as uid
				}
			
			`
		set := `
		uid(aggregation) <aggregation.id> "` + strconv.Itoa(int(uar.ID)) + `" .
		uid(aggregation) <aggregation.name> "` + uar.Name + `" .
		uid(aggregation) <type_name> "aggregation" .
		uid(aggregation) <dgraph.type> "Aggregation" .
		uid(aggregation) <scopes> "` + uar.Scope + `" .
		uid(aggregation) <aggregation.editor> "` + uar.ProductEditor + `" .
		`
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
		upsertreq := &api.Request{
			Query: query,
			Mutations: []*api.Mutation{
				{
					DelNquads: []byte(`uid(aggregation) * * .`),
				},
				{
					SetNquads: []byte(set),
				},
			},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, upsertreq); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}

	case UpsertAggregatedRights:
		var uar UpsertAggregatedRight
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {
			var(func: eq(aggregatedRights.SKU,"` + uar.Sku + `")) @filter(eq(type_name,"aggregatedRights") AND eq(scopes,"` + uar.Scope + `") ){
				aggregatedRights as uid
				}
			
			`
		set := `
		uid(aggregatedRights) <aggregatedRights.SKU> "` + uar.Sku + `" .
		uid(aggregatedRights) <aggregatedRights.aggregationId> "` + strconv.Itoa(int(uar.AggregationID)) + `" .
		uid(aggregatedRights) <type_name> "aggregatedRights" .
		uid(aggregatedRights) <dgraph.type> "AggregatedRights" .
		uid(aggregatedRights) <scopes> "` + uar.Scope + `" .
		uid(aggregatedRights) <aggregatedRights.numOfAcqLicences> "` + strconv.Itoa(int(uar.NumLicensesAcquired)) + `" .
		uid(aggregatedRights) <aggregatedRights.numOfLicencesUnderMaintenance> "` + strconv.Itoa(int(uar.NumLicencesMaintenance)) + `" .
		uid(aggregatedRights) <aggregatedRights.averageUnitPrice> "` + fmt.Sprintf("%.2f", uar.AvgUnitPrice) + `" .
		uid(aggregatedRights) <aggregatedRights.averageMaintenanceUnitPrice> "` + fmt.Sprintf("%.2f", uar.AvgMaintenanceUnitPrice) + `" .
		uid(aggregatedRights) <aggregatedRights.totalPurchaseCost> "` + fmt.Sprintf("%.2f", uar.TotalPurchaseCost) + `" .
		uid(aggregatedRights) <aggregatedRights.totalMaintenanceCost> "` + fmt.Sprintf("%.2f", uar.TotalMaintenanceCost) + `" .
		uid(aggregatedRights) <aggregatedRights.totalCost> "` + fmt.Sprintf("%.2f", uar.TotalCost) + `" .
		uid(aggregatedRights) <aggregatedRights.startOfMaintenance> "` + uar.StartOfMaintenance + `" .
		uid(aggregatedRights) <aggregatedRights.endOfMaintenance> "` + uar.EndOfMaintenance + `" .
		uid(aggregatedRights) <aggregatedRights.orderingDate> "` + uar.OrderingDate + `" .
		uid(aggregatedRights) <aggregatedRights.corporateSourcingContract> "` + uar.CorporateSourcingContract + `" .
		uid(aggregatedRights) <aggregatedRights.softwareProvider> "` + uar.SoftwareProvider + `" .
		uid(aggregatedRights) <aggregatedRights.lastPurchasedOrder> "` + uar.LastPurchasedOrder + `" .
		uid(aggregatedRights) <aggregatedRights.supportNumber> "` + uar.SupportNumber + `" .
		uid(aggregatedRights) <aggregatedRights.maintenanceProvider> "` + uar.MaintenanceProvider + `" .
		uid(aggregatedRights) <aggregatedRights.repartition> "` + strconv.FormatBool(uar.Repartition) + `" .
		`
		reqmetrics := strings.Split(uar.Metric, ",")
		for _, met := range reqmetrics {
			set += `
				uid(aggregatedRights) <aggregatedRights.metric> "` + met + `" .
			`
		}
		query += `}`
		logger.Log.Info(query)
		upsertreq := &api.Request{
			Query: query,
			Mutations: []*api.Mutation{
				{
					DelNquads: []byte(`uid(aggregatedRights) * * .`),
				},
				{
					SetNquads: []byte(set),
				},
			},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, upsertreq); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}

	case DeleteAggregation:
		var dar v1.DeleteAggregationRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
			var(func: eq(aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,` + dar.Scope + `) ){
					aggregation as uid
				}
			var(func: eq(aggregatedRights.aggregationId,"` + strconv.Itoa(int(dar.GetID())) + `")) @filter(eq(type_name,"aggregatedRights") AND eq(scopes,` + dar.Scope + `) ){
					aggRights as uid
				}
			var(func: eq(type_name,"nominative_user")) @filter( eq(nominative.user.aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `") AND eq(scopes,` + dar.Scope + `) ){
					nomUsers as uid
				}
			var(func: eq(type_name,"concurrentUser")) @filter( eq(concurrent.user.aggregation.id,"` + strconv.Itoa(int(dar.GetID())) + `") AND eq(scopes,` + dar.Scope + `) ){
					cuUsers as uid
				}
			}
			`
		delete := `
				uid(aggregation) * * .
				uid(aggRights) * * .
				uid(nomUsers) * * .
				uid(cuUsers) * * .
		`
		set := `
				uid(aggregation) <Recycle> "true" .
				uid(aggRights) <Recycle> "true" .
				uid(nomUsers) <Recycle> "true" .
				uid(cuUsers) <Recycle> "true" .
		`
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
					productAllocations as product.allocation
					productEditors as product.editor
					productApplications as ~product.application
				}
				editorType as var(func: type(Editor)) @filter(eq(scopes,` + dar.Scope + `)){
					editors as editor.name
				}
				var(func: type(User)) @filter(eq(scopes,` + dar.Scope + `)){
					userId as  uid
				}
				var(func: eq(type_name,"concurrentUser")) @filter(eq(concurrent.user.is_aggregations,"false") AND  eq(concurrent.user.aggregation.id,"0") AND eq(scopes,` + dar.Scope + `) ){
					cuUsers as uid
				}
				var(func: eq(type_name,"nominative_user")) @filter(eq(nominative.user.aggregation.id,"0") AND eq(scopes,` + dar.Scope + `) ){
					nomUsers as uid
				}
				`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL { //nolint
			query += ` 
				acqrightsType as var(func: type(AcquiredRights)) @filter(eq(scopes,` + dar.Scope + `)){
				acqrights as acqRights.SKU
				}
				aggrightsType as var(func: type(AggregatedRights)) @filter(eq(scopes,` + dar.Scope + `)){
					aggrights as aggregatedRights.SKU
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
				uid(productAllocations) * * .
				uid(productApplications) * * .
				uid(editorType) * * .
				uid(editors) * * .
				uid(cuUsers) * * .
				uid(nomUsers) * * .
			`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL {
			delete += `
					uid(acqrightsType) * * .
					uid(acqrights) * * .
					uid(aggrightsType) * * .
					uid(aggrights) * * .
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
				uid(productAllocations) <Recycle> "true" .
				uid(productApplications) <Recycle> "true" .
				uid(editorType) <Recycle> "true" .
				uid(editors) <Recycle> "true" .
				uid(cuUsers) <Recycle> "true" .
				uid(nomUsers) <Recycle> "true" .
		`
		}
		if dar.DeletionType == v1.DropProductDataRequest_ACQRIGHTS || dar.DeletionType == v1.DropProductDataRequest_FULL {
			set += `
			uid(acqrightsType) <Recycle> "true" .
			uid(acqrights) <Recycle> "true" .
			uid(aggrightsType) <Recycle> "true" .
			uid(aggrights) <Recycle> "true" .
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
	case DropAggregationData:
		var dagg v1.DropAggregationDataRequest
		_ = json.Unmarshal(e.JSON, &dagg)
		query := `query { 
				aggregationType as var(func: type(Aggregation)) @filter(eq(scopes,` + dagg.Scope + `)){
					aggregation as aggregation.id
				}
				var(func: eq(type_name,"concurrentUser")) @filter(eq(concurrent.user.is_aggregations,"true") AND  gt(concurrent.user.aggregation.id,"0") AND eq(scopes,` + dagg.Scope + `) ){
					cuUsers as uid
				}
				var(func: eq(type_name,"nominative_user")) @filter(gt(nominative.user.aggregation.id,"0") AND eq(scopes,` + dagg.Scope + `) ){
					nomUsers as uid
				}
			}`

		delete := `
				uid(aggregationType) * * .
				uid(aggregation) * * .
				uid(cuUsers) * * .
				uid(nomUsers) * * .
			`
		set := `
				uid(aggregationType) <Recycle> "true" .
				uid(aggregation) <Recycle> "true" .
				uid(cuUsers) <Recycle> "true" .
				uid(nomUsers) <Recycle> "true" .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query, zap.Any("set", set), zap.Any("delete", delete))
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to drop aggregation data from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeleteAggregatedRights:
		var dar DeleteAggregatedRightRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
			var(func: eq(aggregatedRights.SKU,"` + dar.Sku + `")) @filter(eq(scopes,` + dar.Scope + `)){
				aggregatedRights as uid
				}
			}
			`
		delete := `
				uid(aggregatedRights) * * .
		`
		set := `
				uid(aggregatedRights) <Recycle> "true" .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query)
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to delete aggregatedRights from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case UpsertNominativeUserRequest:
		var unur UpserNominativeUserRequest
		_ = json.Unmarshal(e.JSON, &unur)
		var mutations []*api.Mutation
		queries = append(queries, "query", "{")
		for i, v := range unur.UserDetails {
			query := `var(func: eq(nominative.user.email,"` + v.Email + `")) @filter(eq(type_name,"nominative_user") 
			AND eq(scopes,"` + unur.Scope + `") AND eq(nominative.user.profile,"` + v.Profile + `")
			AND eq(nominative.user.swidtag,"` + unur.SwidTag + `") AND eq(nominative.user.aggregation.id,"` + strconv.Itoa(int(unur.AggregationId)) + `")){
				user_` + strconv.Itoa(i) + ` as uid
			}`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
				SetNquads: []byte(`
				uid(user_` + strconv.Itoa(i) + `) <type_name> "nominative_user" .
				uid(user_` + strconv.Itoa(i) + `) <dgraph.type> "NominativeUser" .
				uid(user_` + strconv.Itoa(i) + `) <scopes> "` + unur.Scope + `" .
				uid(user_` + strconv.Itoa(i) + `) <created> "` + unur.CreatedBy + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.swidtag> "` + unur.SwidTag + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.editor> "` + unur.Editor + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.aggregation.id> "` + strconv.Itoa(int(unur.AggregationId)) + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.email> "` + v.Email + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.first_name> "` + v.FirstName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.name> "` + v.UserName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.profile> "` + v.Profile + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.activation.date> "` + v.ActivationDate.String() + `" .
				`),
				CommitNow: true,
			})
			mutations = append(mutations, &api.Mutation{
				Cond: `@if(NOT eq(len(user_` + strconv.Itoa(i) + `),0))`,
				SetNquads: []byte(`
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.first_name> "` + v.FirstName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.name> "` + v.UserName + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.profile> "` + v.Profile + `" .
				uid(user_` + strconv.Itoa(i) + `) <nominative.user.activation.date> "` + v.ActivationDate.String() + `" .
				`),
				CommitNow: true,
			})
			if unur.AggregationId > 0 {
				query = `var(func: eq(aggregation.id,"` + strconv.Itoa(int(unur.AggregationId)) + `")) @filter(eq(type_name,"aggregation") AND eq(scopes,"` + unur.Scope + `")){
					aggregation_` + strconv.Itoa(i) + ` as uid
				}`
				queries = append(queries, query)
				mutations = append(mutations, &api.Mutation{
					Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
					SetNquads: []byte(`
					uid(aggregation_` + strconv.Itoa(i) + `)  <aggregation.nominative.users> uid(user_` + strconv.Itoa(i) + `) .
				`),
					CommitNow: true,
				})
			} else if unur.SwidTag != "" {
				query = `var(func: eq(product.swidtag,"` + unur.SwidTag + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + unur.Scope + `")){
					product_` + strconv.Itoa(i) + ` as uid
				}`
				queries = append(queries, query)
				mutations = append(mutations, &api.Mutation{
					Cond: `@if(eq(len(user_` + strconv.Itoa(i) + `),0))`,
					SetNquads: []byte(`
					uid(product_` + strconv.Itoa(i) + `)  <product.nominative.users> uid(user_` + strconv.Itoa(i) + `) .
				`),
					CommitNow: true,
				})
			}

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
	case UpsertConcurrentUserRequest:
		var uar UpserConcurrentUserRequest
		_ = json.Unmarshal(e.JSON, &uar)
		agID := strconv.Itoa(int(uar.AggregationID))
		currentDateTime := time.Now()

		var queries []string
		var mutations []*api.Mutation
		queries = append(queries, "query", "{")
		query := `
			var(func: eq(type_name,"concurrentUser")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(concurrent.user.purchase_date,"` + uar.PurchaseDate + `") AND eq(concurrent.user.aggregation.id,"` + agID + `")  AND eq(concurrent.user.swidtag,"` + uar.SwidTag + `")  ){
				cuUser as uid
				}
			
			`
		queries = append(queries, query)

		mutations = append(mutations, &api.Mutation{
			Cond: `@if(eq(len(cuUser),0))`,
			SetNquads: []byte(`
				uid(cuUser) <concurrent.user.is_aggregations> "` + strconv.FormatBool(uar.IsAggregations) + `" .
				uid(cuUser) <concurrent.user.aggregation.id> "` + agID + `" .
				uid(cuUser) <concurrent.user.swidtag> "` + uar.SwidTag + `" .
				uid(cuUser) <type_name> "concurrentUser" .
				uid(cuUser) <dgraph.type> "ProductConcurrentUser" .
				uid(cuUser) <scopes> "` + uar.Scope + `" .
				
				uid(cuUser) <concurrent.user.purchase_date> "` + uar.PurchaseDate + `" .
				uid(cuUser) <created> "` + currentDateTime.String() + `" .
			`),
			DelNquads: []byte(`uid(cuUser) * * .`),
			CommitNow: true,
		})

		mutations = append(mutations, &api.Mutation{
			SetNquads: []byte(`
				uid(cuUser) <concurrent.user.number_of_users> "` + strconv.Itoa(int(uar.NumberOfUsers)) + `" .
				uid(cuUser) <concurrent.user.profile_user> "` + uar.ProfileUser + `" .
				uid(cuUser) <concurrent.user.team> "` + uar.Team + `" .
				uid(cuUser) <updated> "` + currentDateTime.String() + `" .
			`),
		})

		// re-establish the link between concurrent user and product or aggregations
		if uar.IsAggregations && uar.AggregationID > 0 {
			query := `
			var(func: eq(type_name,"aggregation")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(aggregation.id,"` + agID + `")){
				aggUID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				SetNquads: []byte(`
					uid(aggUID) <aggregation.concurrent.users> uid(cuUser) .
				`),
			})

		} else {
			query := `
			var(func: eq(type_name,"product")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(product.swidtag,"` + uar.SwidTag + `")){
				pID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				SetNquads: []byte(`
					uid(pID) <product.concurrent.users> uid(cuUser) .
				`),
			})
		}

		queries = append(queries, "}")
		q := strings.Join(queries, "\n")
		upsertreq := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, upsertreq); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph - UpsertConcurrentUserRequest", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeleteConcurrentUserRequest:
		var uar UpserConcurrentUserRequest
		_ = json.Unmarshal(e.JSON, &uar)
		agID := strconv.Itoa(int(uar.AggregationID))
		var queries []string
		var mutations []*api.Mutation
		queries = append(queries, "query", "{")

		query := `
			var(func: eq(type_name,"concurrentUser")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(concurrent.user.purchase_date,"` + uar.PurchaseDate + `") AND (eq(concurrent.user.aggregation.id,"` + agID + `")  OR eq(concurrent.user.swidtag,"` + uar.SwidTag + `")  )){
				cuUser as uid
				}
			`

		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(cuUser),0))`,
			SetNquads: []byte(`
				uid(cuUser) <Recycle> "true" .
			`),
			DelNquads: []byte(`uid(cuUser) * * .`),
			CommitNow: true,
		})

		// un-establish the link between concurrent user and product or aggregations
		if uar.IsAggregations && uar.AggregationID > 0 {
			query := `
			var(func: eq(type_name,"aggregation")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(aggregation.id,"` + agID + `")){
				aggUID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
					uid(aggUID) <aggregation.concurrent.users> uid(cuUser) .
				`),
			})

		} else {
			query := `
			var(func: eq(type_name,"product")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(product.swidtag,"` + uar.SwidTag + `")){
				pID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
					uid(pID) <product.concurrent.users> uid(cuUser) .
				`),
			})
		}

		queries = append(queries, "}")
		q := strings.Join(queries, "\n")
		dropreq := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}

		if _, err := w.dg.NewTxn().Do(ctx, dropreq); err != nil {
			logger.Log.Error("Failed to delete concurrent user from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeleteNominativeUserRequest:
		var uar UpserNominativeUserRequest
		_ = json.Unmarshal(e.JSON, &uar)
		agID := strconv.Itoa(int(uar.AggregationId))
		var queries []string
		var mutations []*api.Mutation
		queries = append(queries, "query", "{")

		query := `
			var(func: eq(nominative.user.email,"` + uar.UserDetails[0].Email + `")) @filter(eq(type_name,"nominative_user") 
			AND eq(scopes,"` + uar.Scope + `") AND eq(nominative.user.profile,"` + uar.UserDetails[0].Profile + `")
			AND eq(nominative.user.swidtag,"` + uar.SwidTag + `") AND eq(nominative.user.aggregation.id,"` + agID + `")){
				nuUser as uid
				}
			`

		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(nuUser),0))`,
			SetNquads: []byte(`
				uid(nuUser) <Recycle> "true" .
			`),
			DelNquads: []byte(`uid(nuUser) * * .`),
			CommitNow: true,
		})

		// un-establish the link between nominative user and product or aggregations
		if uar.AggregationId > 0 {
			query := `
			var(func: eq(type_name,"aggregation")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(aggregation.id,"` + agID + `")){
				aggUID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
					uid(aggUID) <aggregation.nominative.users> uid(nuUser) .
				`),
			})

		} else {
			query := `
			var(func: eq(type_name,"product")) @filter(eq(scopes,"` + uar.Scope + `")  AND eq(product.swidtag,"` + uar.SwidTag + `")){
				pID as uid
			}
			`
			queries = append(queries, query)
			mutations = append(mutations, &api.Mutation{
				DelNquads: []byte(`
					uid(pID) <product.nominative.users> uid(nuUser) .
				`),
			})
		}

		queries = append(queries, "}")
		q := strings.Join(queries, "\n")
		dropreq := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}

		if _, err := w.dg.NewTxn().Do(ctx, dropreq); err != nil {
			logger.Log.Error("Failed to delete nominative user from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case DeletSaaSProductRequest:
		var uar DeleteProductRequest
		_ = json.Unmarshal(e.JSON, &uar)
		var queries []string
		var mutations []*api.Mutation
		queries = append(queries, "query", "{")

		query := `var(func: eq(product.swidtag,"` + uar.SwidTag + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + uar.Scope + `")){
			product as uid
		}`
		queries = append(queries, query)
		mutations = append(mutations, &api.Mutation{
			Cond: `@if(NOT eq(len(product),0))`,
			SetNquads: []byte(`
				uid(product) <Recycle> "true" .
			`),
			DelNquads: []byte(`uid(product) * * .`),
			CommitNow: true,
		})

		queries = append(queries, "}")
		q := strings.Join(queries, "\n")
		dropreq := &api.Request{
			Query:     q,
			Mutations: mutations,
			CommitNow: true,
		}

		if _, err := w.dg.NewTxn().Do(ctx, dropreq); err != nil {
			logger.Log.Sugar().Errorw("Failed to delete product user from Dgraph-DeletSaaSProductRequest",
				"scope", uar.Scope,
				"swidtag", uar.SwidTag,
				"error", err.Error(),
				"query", q,
			)
			return errors.New("RETRY")
		}
	default:
		fmt.Println(e.JSON)
	}

	// Everything's fine, we're done here
	return nil
}
