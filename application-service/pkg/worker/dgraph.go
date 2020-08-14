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
	v1 "optisam-backend/application-service/pkg/api/v1"
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
	UpsertApplicationRequest MessageType = "UpsertApplication"
	UpsertInstanceRequest    MessageType = "UpsertInstance"
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

func (w *Worker) DoWork(ctx context.Context, j *job.Job) error {
	//do something cool!
	var e Envelope
	_ = json.Unmarshal(j.Data, &e)
	fmt.Println(e.Type)
	switch e.Type {
	case UpsertApplicationRequest:
		var uar v1.UpsertApplicationRequest
		_ = json.Unmarshal(e.JSON, &uar)
		query := `query {application as var(func: eq(application.id,` + uar.GetApplicationId() + `)) @filter(eq(type_name,"application"))}`
		mu := &api.Mutation{SetNquads: []byte(`
		uid(application) <application.id> "` + uar.GetApplicationId() + `" .
		uid(application) <application.name>"` + uar.GetName() + `" .
		uid(application) <application.version>"` + uar.GetVersion() + `" .
		uid(application) <application.owner>"` + uar.GetOwner() + `" .
		uid(application) <scopes> "` + uar.GetScope() + `" .
		uid(application) <type_name> "application" .
		uid(application) <dgraph.type> "Application" .
		`)}
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{mu},
			CommitNow: true,
		}
		if _, err := w.dg.NewTxn().Do(ctx, req); err != nil {
			logger.Log.Error("Failed to upsert to Dgraph", zap.Error(err))
			return errors.New("RETRY")
		}
	case UpsertInstanceRequest:
		var uir v1.UpsertInstanceRequest
		_ = json.Unmarshal(e.JSON, &uir)
		fmt.Println(uir)
		query := `query {
			  var(func: eq(instance.id,"` + uir.GetInstanceId() + `")) @filter(eq(type_name,"instance")){
				  instance as uid
			  }
			`
		addProductEquipment := `
		uid(instance) <instance.id> "` + uir.GetInstanceId() + `" .
		uid(instance) <scopes> "` + uir.GetScope() + `" .
		uid(instance) <type_name> "instance" .
		uid(instance) <dgraph.type> "Instance" .
		`

		if uir.Products.GetOperation() == "add" {
			for i, product := range uir.GetProducts().GetProductId() {
				query += `
				var(func: eq(product.swidtag,"` + product + `")) @filter(eq(type_name,"product")){
					product` + strconv.Itoa(i) + ` as uid
				}
				`
				addProductEquipment += `
				uid(product` + strconv.Itoa(i) + `) <product.swidtag> "` + product + `" .
				uid(product` + strconv.Itoa(i) + `) <type_name> "product" .
				uid(product` + strconv.Itoa(i) + `) <dgraph.type> "Product" .
				uid(product` + strconv.Itoa(i) + `) <scopes> "` + uir.GetScope() + `" .
				uid(instance) <instance.product> uid(product` + strconv.Itoa(i) + `) .
				`
			}
		}

		if uir.Equipments.GetOperation() == "add" {
			for i, equipment := range uir.GetEquipments().GetEquipmentId() {
				query += `
				var(func: eq(equipment.id,"` + equipment + `")) @filter(eq(type_name,"equipment")){
					equipment` + strconv.Itoa(i) + ` as uid
				}
				`
				addProductEquipment += `
				uid(equipment` + strconv.Itoa(i) + `) <equipment.id> "` + equipment + `" .
				uid(equipment` + strconv.Itoa(i) + `) <type_name> "equipment" .
				uid(equipment` + strconv.Itoa(i) + `) <dgraph.type> "Equipment" .
				uid(equipment` + strconv.Itoa(i) + `) <scopes> "` + uir.GetScope() + `" .
				uid(instance) <instance.equipment> uid(equipment` + strconv.Itoa(i) + `) .
				`
			}
		}

		if uir.GetApplicationId() != "" {
			query += `
			var(func: eq(application.id,"` + uir.GetApplicationId() + `")) @filter(eq(type_name,"application")){
				application as uid
			}`
			addProductEquipment += `
			uid(instance) <instance.environment>"` + uir.GetInstanceName() + `" .
			uid(application) <application.instance> uid(instance) .
			uid(application) <application.id> "` + uir.GetApplicationId() + `" .
			uid(application) <type_name> "application" .
			uid(application) <dgraph.type> "Application" .
			uid(application) <scopes> "` + uir.GetScope() + `" .
			`
		}
		//end query block
		query += `
		}`
		logger.Log.Info(query)

		muAdd := &api.Mutation{SetNquads: []byte(addProductEquipment)}
		req := &api.Request{
			Query:     query,
			Mutations: []*api.Mutation{muAdd},
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
