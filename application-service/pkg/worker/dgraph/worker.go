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
	v1 "optisam-backend/application-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	"strconv"
	"sync"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

var mu sync.Mutex

//Worker ...
type Worker struct {
	id string
	dg *dgo.Dgraph
}

//MessageType ...
type MessageType string

const (
	//UpsertApplicationRequest is request type for upsert application in dgraph
	UpsertApplicationRequest MessageType = "UpsertApplication"
	//UpsertInstanceRequest is request type for upsert instances in dgraph
	UpsertInstanceRequest MessageType = "UpsertInstance"
	//DropApplicationDataRequest is request type for delete applications and applications instances in dgraph for a particular scope
	DropApplicationDataRequest MessageType = "DropApplicationData"
)

//Envelope ...
type Envelope struct {
	Type MessageType `json:"message_type"`
	JSON json.RawMessage
}

//NewWorker ...
func NewWorker(id string, dg *dgo.Dgraph) *Worker {
	return &Worker{id: id, dg: dg}
}

//ID ...
func (w *Worker) ID() string {
	return w.id
}

//DoWork impletation of work
func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	mu.Lock()
	defer mu.Unlock()
	//do something cool!
	var e Envelope
	_ = json.Unmarshal(j.Data, &e)
	fmt.Println(e.Type)
	switch e.Type {
	case UpsertApplicationRequest:
		var mutations []*api.Mutation
		var uar v1.UpsertApplicationRequest
		_ = json.Unmarshal(e.JSON, &uar)
		//SCOPE BASED CHANGE
		query := `query {application as var(func: eq(application.id,` + uar.GetApplicationId() + `)) @filter(eq(type_name,"application") AND eq(scopes,"` + uar.GetScope() + `"))}`
		mu := &api.Mutation{SetNquads: []byte(`
		uid(application) <application.id> "` + uar.GetApplicationId() + `" .
		uid(application) <application.name>"` + uar.GetName() + `" .
		uid(application) <application.version>"` + uar.GetVersion() + `" .
		uid(application) <application.owner>"` + uar.GetOwner() + `" .
		uid(application) <application.domain>"` + uar.GetDomain() + `" .
		uid(application) <scopes> "` + uar.GetScope() + `" .
		uid(application) <type_name> "application" .
		uid(application) <dgraph.type> "Application" .
		`)}
		mutations = append(mutations, mu)
		req := &api.Request{
			Query:     query,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err), zap.String("query", req.Query), zap.Any("mutation", req.Mutations))
			return errors.New("RETRY")
		}
	case UpsertInstanceRequest:
		var mutations []*api.Mutation
		var uir v1.UpsertInstanceRequest
		_ = json.Unmarshal(e.JSON, &uir)
		fmt.Println(uir)
		//SCOPE BASED CHANGE
		query := `query {
			  var(func: eq(instance.id,"` + uir.GetInstanceId() + `")) @filter(eq(type_name,"instance") AND eq(scopes,"` + uir.GetScope() + `") ){
				  instance as uid
			  }
			`
		mutations = append(mutations, &api.Mutation{
			Cond: "@if(eq(len(instance),0))",
			SetNquads: []byte(`
			uid(instance) <instance.id> "` + uir.GetInstanceId() + `" .
			uid(instance) <scopes> "` + uir.GetScope() + `" .
			uid(instance) <type_name> "instance" .
			uid(instance) <dgraph.type> "Instance" .
			`),
		})

		if uir.Products.GetOperation() == "add" {
			for i, product := range uir.GetProducts().GetProductId() {
				prodUID := `product` + strconv.Itoa(i)
				//SCOPE BASED CHANGE
				query += `
				var(func: eq(product.swidtag,"` + product + `")) @filter(eq(type_name,"product") AND eq(scopes,"` + uir.GetScope() + `")){
					product` + strconv.Itoa(i) + ` as uid
				}
				`
				//queries = append(queries, query)
				mutations = append(mutations, &api.Mutation{
					Cond: "@if(eq(len(" + prodUID + "),0))",
					SetNquads: []byte(`
					uid(` + prodUID + `) <product.swidtag> "` + product + `" .
					uid(` + prodUID + `) <type_name> "product" .
					uid(` + prodUID + `) <dgraph.type> "Product" .
					uid(` + prodUID + `) <scopes> "` + uir.GetScope() + `" .
					`),
				})
				mutations = append(mutations, &api.Mutation{
					SetNquads: []byte(`
					uid(instance) <instance.product> uid(` + prodUID + `) .
					`),
				})
			}
		}

		if uir.Equipments.GetOperation() == "add" {
			for i, equipment := range uir.GetEquipments().GetEquipmentId() {
				eqUID := `equipment` + strconv.Itoa(i)

				//SCOPE BASED CHANGE
				query += `
				var(func: eq(equipment.id,"` + equipment + `")) @filter(eq(type_name,"equipment") AND eq(scopes,"` + uir.GetScope() + `")){
					equipment` + strconv.Itoa(i) + ` as uid
				}
				`
				//queries = append(queries, query)
				mutations = append(mutations, &api.Mutation{
					Cond: "@if(eq(len(" + eqUID + "),0))",
					SetNquads: []byte(`
					uid(` + eqUID + `) <equipment.id> "` + equipment + `" .
					uid(` + eqUID + `) <type_name> "equipment" .
					uid(` + eqUID + `) <dgraph.type> "Equipment" .
					uid(` + eqUID + `) <scopes> "` + uir.GetScope() + `" .
					`),
				})
				mutations = append(mutations, &api.Mutation{
					SetNquads: []byte(`
					uid(instance) <instance.equipment> uid(` + eqUID + `) .
					`),
				})
			}
		}

		if uir.GetApplicationId() != "" {
			//SCOPE BASED CHANGE
			query += `
			var(func: eq(application.id,"` + uir.GetApplicationId() + `")) @filter(eq(type_name,"application") AND eq(scopes,"` + uir.GetScope() + `")){
				application as uid
			}`

			mutations = append(mutations, &api.Mutation{
				Cond: "@if(eq(len(application),0))",
				SetNquads: []byte(`
				uid(application) <application.id> "` + uir.GetApplicationId() + `" .
				uid(application) <type_name> "application" .
				uid(application) <dgraph.type> "Application" .
				uid(application) <scopes> "` + uir.GetScope() + `" .
				`),
			})
			mutations = append(mutations, &api.Mutation{
				SetNquads: []byte(`
				uid(instance) <instance.environment>"` + uir.GetInstanceName() + `" .
				uid(application) <application.instance> uid(instance) .
				`),
			})
		}
		//end query block
		query += "}"
		req := &api.Request{
			Query:     query,
			Mutations: mutations,
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed upsert application instances to Dgraph", zap.Error(err), zap.String("Query", req.Query), zap.Any("mutation", req.Mutations))
			return errors.New("RETRY")
		}
	case DropApplicationDataRequest:
		var dar v1.DropApplicationDataRequest
		_ = json.Unmarshal(e.JSON, &dar)
		query := `query {
			  applicationType as var(func: type(Application)) @filter(eq(scopes,` + dar.Scope + `)){
				applications as application.id
				applicationInstances as application.instance
			  }
			  instanceType as var(func: type(Instance)) @filter(eq(scopes,` + dar.Scope + `)){
				instances as instance.id
			  }
			}
			`
		delete := `
				uid(applicationType) * * .
				uid(applications) * * .
				uid(applicationInstances) * * .
				uid(instanceType) * * .
				uid(instances) * * .
		`
		set := `
				uid(applicationType) <Recycle> "true" .
				uid(applications) <Recycle> "true" .
				uid(applicationInstances) <Recycle> "true" .
				uid(instanceType) <Recycle> "true" .
				uid(instances) <Recycle> "true" .
		`
		muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
		logger.Log.Info(query)
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muDelete},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to delete applications from Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	default:
		fmt.Println(e.JSON)
	}

	//Everything's fine, we're done here
	return nil
}
